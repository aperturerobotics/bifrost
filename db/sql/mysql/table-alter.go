package mysql

import (
	"io"
	"slices"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/pkg/errors"
	kvtx_block "github.com/s4wave/spacewave/db/kvtx/block"
)

// AddColumn adds a column and rewrites stored rows to the new schema.
func (t *Table) AddColumn(ctx *sql.Context, column *sql.Column, order *sql.ColumnOrder) error {
	// Require a new column definition with an unused table column name.
	if column == nil {
		return errors.New("column is nil")
	}
	if _, ok := t.columnOrdinal(column.Name); ok {
		return sql.ErrColumnExists.New(column.Name)
	}

	// Resolve the new column source and its requested schema position.
	nextColumn := *column
	if nextColumn.Source == "" {
		nextColumn.Source = t.name
	}
	insertAt, err := t.columnOrderIndex(order, len(t.schema.Schema))
	if err != nil {
		return err
	}

	// Compute the new column default for existing table rows.
	defaultValue, err := columnDefaultValue(ctx, &nextColumn)
	if err != nil {
		return err
	}

	// Rewrite table rows with the new column inserted into their schema.
	newSchema := slices.Insert(slices.Clone(t.schema.Schema), insertAt, &nextColumn)
	newPK := primaryKeySchemaFor(newSchema)
	return t.rewriteRows(ctx, newPK, "", "", func(row sql.Row) (sql.Row, error) {
		// Build the row with the new column default at its schema position.
		next := make(sql.Row, len(row)+1)
		copy(next, row[:insertAt])
		next[insertAt] = defaultValue
		copy(next[insertAt+1:], row[insertAt:])
		return next, nil
	})
}

// DropColumn drops a column and rewrites stored rows to the new schema.
func (t *Table) DropColumn(ctx *sql.Context, columnName string) error {
	// Require an existing table column that is outside the primary key.
	dropAt, ok := t.columnOrdinal(columnName)
	if !ok {
		return sql.ErrTableColumnNotFound.New(t.name, columnName)
	}
	if t.schema.Schema[dropAt].PrimaryKey {
		return errors.Errorf("cannot drop primary key column %q", columnName)
	}

	// Reject removal of a column referenced by a secondary index.
	for _, index := range t.root.GetIndexes() {
		for _, indexCol := range index.GetColumns() {
			if strings.EqualFold(indexCol.GetName(), columnName) {
				return errors.Errorf("cannot drop column %q referenced by index %q", columnName, index.GetName())
			}
		}
	}

	// Rewrite table rows without the removed column.
	newSchema := slices.Delete(slices.Clone(t.schema.Schema), dropAt, dropAt+1)
	newPK := primaryKeySchemaFor(newSchema)
	return t.rewriteRows(ctx, newPK, "", "", func(row sql.Row) (sql.Row, error) {
		next := make(sql.Row, len(row)-1)
		copy(next, row[:dropAt])
		copy(next[dropAt:], row[dropAt+1:])
		return next, nil
	})
}

// ModifyColumn replaces an existing column definition and optionally moves it.
func (t *Table) ModifyColumn(ctx *sql.Context, columnName string, column *sql.Column, order *sql.ColumnOrder) error {
	// Require an existing column and its replacement definition.
	oldAt, ok := t.columnOrdinal(columnName)
	if !ok {
		return sql.ErrTableColumnNotFound.New(t.name, columnName)
	}
	if column == nil {
		return errors.New("column is nil")
	}

	// Resolve the replacement column source and requested position.
	nextColumn := *column
	if nextColumn.Source == "" {
		nextColumn.Source = t.name
	}
	withoutOld := slices.Delete(slices.Clone(t.schema.Schema), oldAt, oldAt+1)
	insertAt, err := t.columnOrderIndex(order, len(withoutOld))
	if err != nil {
		return err
	}

	// Rewrite table rows with the replacement column at its new position.
	newSchema := slices.Insert(withoutOld, insertAt, &nextColumn)
	newPK := primaryKeySchemaFor(newSchema)
	return t.rewriteRows(ctx, newPK, columnName, nextColumn.Name, func(row sql.Row) (sql.Row, error) {
		// Preserve the existing column value while moving it within the row.
		value := row[oldAt]
		next := make(sql.Row, 0, len(row))
		next = append(next, row[:oldAt]...)
		next = append(next, row[oldAt+1:]...)
		next = slices.Insert(next, insertAt, value)
		return next, nil
	})
}

// rewriteRows atomically rebuilds row identities and secondary trees for a schema change.
func (t *Table) rewriteRows(
	ctx *sql.Context,
	newSchema sql.PrimaryKeySchema,
	renamedFrom, renamedTo string,
	transform func(sql.Row) (sql.Row, error),
) error {
	// Build the new layout in a detached table, leaving the old tree readable.
	next := *t
	next.root = t.root.CloneVT()
	next.bcs = t.bcs.Detach(false)
	next.bcs.ClearAllRefs()
	next.root.TableSchema = NewTableSchema(newSchema.Schema)
	next.setPrimaryKeyOrdinals(newSchema.PkOrdinals)
	next.root.RowNonce = 0

	// Reset stored trees so every row and secondary reference is rebuilt.
	for i := range next.root.TablePartitions {
		next.root.TablePartitions[i] = NewTablePartitionRoot()
	}
	for _, index := range next.root.Indexes {
		index.KeyValue = kvtx_block.NewKeyValueStore(0)
	}
	if renamedFrom != "" {
		next.renameIndexColumn(renamedFrom, renamedTo)
	}

	// Load the replacement schema and counters on the detached table.
	next.bcs.SetBlock(next.root, true)
	if err := next.loadRootState(GetDbContext(ctx), next.root); err != nil {
		return err
	}

	// Transform each original row and insert it through normal key constraints.
	editor := next.NewTableEditor(ctx)
	for part := range t.root.TablePartitions {
		// Open the original partition without modifying its cursors or stored keys.
		partition, err := t.PartitionAtIndex(part)
		if err != nil {
			return err
		}
		rows, err := partition.IterateRows(ctx)
		if err != nil {
			return err
		}

		// Rebuild every row and index reference in the replacement table.
		err = func() error {
			defer rows.Close(ctx)
			for {
				// Transform and validate the row against its replacement schema.
				row, err := rows.Next(ctx)
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return err
				}
				row, err = transform(row)
				if err != nil {
					return err
				}
				for i, column := range newSchema.Schema {
					value, _, err := column.Type.Convert(ctx, row[i])
					if err != nil {
						return err
					}
					row[i] = value
				}
				if err := newSchema.CheckRow(ctx, row); err != nil {
					return err
				}

				// Reuse the editor so primary keys and unique indexes remain authoritative.
				if err := editor.Insert(ctx, row); err != nil {
					return err
				}
			}
		}()
		if err != nil {
			return err
		}
	}

	// Publish the complete replacement layout only after every row succeeds.
	return t.replaceRoot(GetDbContext(ctx), next.bcs)
}

// columnOrderIndex resolves a requested position in the table schema.
func (t *Table) columnOrderIndex(order *sql.ColumnOrder, defaultIndex int) (int, error) {
	// Use the default column position when the request omits an order.
	if order == nil {
		return defaultIndex, nil
	}

	// Resolve explicit first position or an empty after-column request.
	if order.First {
		return 0, nil
	}
	if order.AfterColumn == "" {
		return defaultIndex, nil
	}

	// Locate the named preceding column in the current table schema.
	ord, ok := t.columnOrdinal(order.AfterColumn)
	if !ok {
		return 0, sql.ErrTableColumnNotFound.New(t.name, order.AfterColumn)
	}
	return ord + 1, nil
}

// columnDefaultValue evaluates a new column's default for existing rows.
func columnDefaultValue(ctx *sql.Context, column *sql.Column) (any, error) {
	if column.Default != nil {
		return column.Default.Eval(ctx, nil)
	}
	if !column.Nullable {
		return nil, errors.Errorf("column %q requires a default", column.Name)
	}
	return nil, nil
}

// primaryKeySchemaFor derives primary ordinals from the current column positions.
func primaryKeySchemaFor(schema sql.Schema) sql.PrimaryKeySchema {
	ordinals := make([]int, 0)
	for i, col := range schema {
		if col.PrimaryKey {
			ordinals = append(ordinals, i)
		}
	}
	return sql.NewPrimaryKeySchema(schema, ordinals...)
}

// setPrimaryKeyOrdinals stores primary-key column positions.
func (t *Table) setPrimaryKeyOrdinals(ordinals []int) {
	t.root.PrimaryKeyOrdinals = make([]int32, len(ordinals))
	for i, ord := range ordinals {
		t.root.PrimaryKeyOrdinals[i] = int32(ord) //nolint:gosec
	}
}

// renameIndexColumn updates secondary definitions after a column rename.
func (t *Table) renameIndexColumn(oldName, newName string) {
	if strings.EqualFold(oldName, newName) {
		return
	}
	for _, index := range t.root.GetIndexes() {
		for _, col := range index.GetColumns() {
			if strings.EqualFold(col.GetName(), oldName) {
				col.Name = newName
			}
		}
	}
}
