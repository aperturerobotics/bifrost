package mysql

import (
	"bytes"
	"context"
	"encoding/binary"
	"slices"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block/blob"
)

// TableEditor implements row management operations against a table.
//
// Note: all table operations are (currently) not concurrency safe.
type TableEditor struct {
	// ctx supplies the default operation context.
	ctx context.Context
	// t owns the row and index trees.
	t *Table
	// buildBlobOpts controls row value encoding.
	buildBlobOpts *blob.BuildBlobOpts
	// statement indicates whether row edits need rollback records.
	statement bool
	// statementNonce is the hidden row counter before the statement.
	statementNonce uint64
	// statementAutoIncrement is the auto-increment counter before the statement.
	statementAutoIncrement uint64
	// changes records row edits in application order for statement rollback.
	changes []tableRowChange
}

// tableRowChange records the row identities needed to undo one editor operation.
type tableRowChange struct {
	// before is the replaced row, or nil for an insertion.
	before sql.Row
	// after is the replacement row, or nil for a deletion.
	after sql.Row
	// oldKey is the original stored identity.
	oldKey []byte
	// newKey is the replacement stored identity.
	newKey []byte
}

// NewTableEditor constructs a new table row inserter.
func NewTableEditor(ctx context.Context, t *Table) *TableEditor {
	if ctx == nil {
		ctx = context.Background()
	}
	return &TableEditor{
		ctx: ctx,
		t:   t,
	}
}

// SetBuildBlobOpts sets the build blob options.
func (i *TableEditor) SetBuildBlobOpts(opts *blob.BuildBlobOpts) {
	i.buildBlobOpts = opts
}

// StatementBegin starts a row-edit journal and snapshots the table counters.
func (i *TableEditor) StatementBegin(ctx *sql.Context) {
	i.statement = true
	i.statementNonce = i.t.root.RowNonce
	i.statementAutoIncrement = i.t.autoIncVal
	i.changes = nil
}

// Insert inserts the row given, returning an error if it cannot. Insert will be
// called once for each row to process for the insert operation, which may
// involve many rows. After all rows in an operation have been processed, Close
// is called.
func (i *TableEditor) Insert(sqlCtx *sql.Context, row sql.Row) error {
	// Use the SQL operation context for row access.
	cctx := i.ctx
	if sqlCtx != nil && sqlCtx.Context != nil {
		cctx = sqlCtx.Context
	}

	// Provide a SQL context for schema and constraint checks.
	checkCtx := sqlCtx
	if checkCtx == nil {
		checkCtx = sql.NewContext(cctx)
	}

	// Require the inserted row to match the table column count.
	schema := i.t.schema.Schema
	if len(row) != len(schema) {
		return sql.ErrInvalidColumnNumber.New(len(schema), len(row))
	}

	// Encode the primary row identity, retaining a nonce only for keyless duplicates.
	rowKey, err := i.t.rowKey(checkCtx, row)
	if err != nil {
		return err
	}
	if len(i.t.schema.PkOrdinals) == 0 {
		rowKey = binary.BigEndian.AppendUint64(rowKey, i.t.root.RowNonce)
	}

	// Check the inserted row against primary and unique keys.
	if err := i.ensureUniqueRow(checkCtx, row, nil); err != nil {
		return err
	}

	// Advance the table auto-increment value past the inserted row.
	autoIncIdx := i.t.autoIncIdx
	schemaCols := i.t.schema.Schema
	if autoIncIdx != 0 {
		// Locate the auto-increment column and compare its inserted value.
		autoIncIdx-- // 1-based index
		autoIncVal := i.t.autoIncVal
		if autoIncIdx >= len(schemaCols) {
			return errors.Errorf("auto increment index out of range: %d > %d", autoIncIdx, len(schemaCols)-1)
		}
		autoIncCol := schemaCols[autoIncIdx]
		cmp, err := autoIncCol.Type.Compare(sqlCtx, row[autoIncIdx], autoIncVal)
		if err != nil {
			return errors.Wrap(err, "auto increment type mismatch")
		}

		// Convert larger inserted values into the next auto-increment value.
		if cmp > 0 {
			v, _, err := types.Uint64.Convert(sqlCtx, row[autoIncIdx])
			if err != nil {
				return errors.Wrap(err, "auto increment type mismatch")
			}

			// Require the converted auto-increment value to be an unsigned integer.
			var ok bool
			autoIncVal, ok = v.(uint64)
			if !ok {
				return errors.Errorf("auto increment type mismatch: converted to %T", v)
			}
			autoIncVal++ // Move onto next autoIncVal
		} else if cmp == 0 {
			autoIncVal++
		}

		// Persist the next auto-increment value in the table root.
		err = i.SetAutoIncrementValue(sqlCtx, autoIncVal)
		if err != nil {
			return err
		}
	}

	// Journal the inserted identity before changing row or index storage.
	if i.statement {
		i.changes = append(i.changes, tableRowChange{after: slices.Clone(row), newKey: rowKey})
	}
	if err := i.storeRow(checkCtx, row, rowKey); err != nil {
		return err
	}

	// Reserve the next hidden identity only for a keyless row.
	if len(i.t.schema.PkOrdinals) == 0 {
		i.t.root.RowNonce++
	}
	i.t.bcs.SetBlock(i.t.root, true)
	return nil
}

// Update updates the old row to the new row.
func (i *TableEditor) Update(sqlCtx *sql.Context, oldRow, newRow sql.Row) error {
	// Use the SQL operation context for row access.
	cctx := i.ctx
	if sqlCtx != nil && sqlCtx.Context != nil {
		cctx = sqlCtx.Context
	}

	// Provide a SQL context for schema and constraint checks.
	checkCtx := sqlCtx
	if checkCtx == nil {
		checkCtx = sql.NewContext(cctx)
	}

	// Require both row versions to match the table schema.
	schema := i.t.schema.Schema
	if len(oldRow) != len(schema) {
		return sql.ErrInvalidColumnNumber.New(len(schema), len(oldRow))
	}
	if len(newRow) != len(schema) {
		return sql.ErrInvalidColumnNumber.New(len(schema), len(newRow))
	}
	if err := schema.CheckRow(checkCtx, oldRow); err != nil {
		return err
	}
	if err := schema.CheckRow(checkCtx, newRow); err != nil {
		return err
	}

	// Locate the existing row and check the replacement for key conflicts.
	rowKey, err := i.findRowKey(checkCtx, oldRow)
	if err != nil {
		return err
	}
	if err := i.ensureUniqueRow(checkCtx, newRow, rowKey); err != nil {
		return err
	}

	// Recompute the row identity when primary columns or keyless values change.
	newKey, err := i.t.rowKey(checkCtx, newRow)
	if err != nil {
		return err
	}
	if len(i.t.schema.PkOrdinals) == 0 {
		newKey = append(newKey, rowKey[len(rowKey)-8:]...)
	}

	// Journal both identities before replacing the row and its secondary references.
	if i.statement {
		i.changes = append(i.changes, tableRowChange{
			before: slices.Clone(oldRow), after: slices.Clone(newRow), oldKey: rowKey, newKey: newKey,
		})
	}
	if err := i.removeRow(checkCtx, oldRow, rowKey); err != nil {
		return err
	}
	return i.storeRow(checkCtx, newRow, newKey)
}

// Delete deletes the row given.
func (i *TableEditor) Delete(sqlCtx *sql.Context, row sql.Row) error {
	// Use the SQL operation context for row access.
	cctx := i.ctx
	if sqlCtx != nil && sqlCtx.Context != nil {
		cctx = sqlCtx.Context
	}

	// Provide a SQL context for schema and constraint checks.
	checkCtx := sqlCtx
	if checkCtx == nil {
		checkCtx = sql.NewContext(cctx)
	}

	// Require the deleted row to match the table schema.
	if len(row) != len(i.t.schema.Schema) {
		return sql.ErrInvalidColumnNumber.New(len(i.t.schema.Schema), len(row))
	}
	if err := i.t.schema.CheckRow(checkCtx, row); err != nil {
		return err
	}

	// Locate the deleted row in its table partition.
	rowKey, err := i.findRowKey(checkCtx, row)
	if err != nil {
		return err
	}

	// Journal the removed identity before changing row or index storage.
	if i.statement {
		i.changes = append(i.changes, tableRowChange{before: slices.Clone(row), oldKey: rowKey})
	}
	return i.removeRow(checkCtx, row, rowKey)
}

// storeRow writes one row and all of its secondary references without constraint checks.
func (i *TableEditor) storeRow(ctx *sql.Context, row sql.Row, key []byte) error {
	// Build the row independently so value encoding cannot partially change its tree.
	cursor := i.t.bcs.Detach(false)
	cursor.ClearAllRefs()
	if _, err := BuildTableRow(GetDbContext(ctx), cursor, row, i.buildBlobOpts); err != nil {
		return err
	}

	// Write the row reference into its primary partition.
	part, err := i.t.SelectPartition(key)
	if err != nil {
		return err
	}
	tree, err := part.BuildTreeTx(GetDbContext(ctx), false, true)
	if err != nil {
		return err
	}
	if err := tree.SetCursorAtKey(GetDbContext(ctx), key, cursor, false); err != nil {
		return err
	}
	return i.t.writeIndexes(ctx, row, key, false)
}

// removeRow removes one row and its secondary references without constraint checks.
func (i *TableEditor) removeRow(ctx *sql.Context, row sql.Row, key []byte) error {
	// Remove secondary references before deleting the primary row identity.
	if err := i.t.writeIndexes(ctx, row, key, true); err != nil {
		return err
	}
	part, err := i.t.SelectPartition(key)
	if err != nil {
		return err
	}
	tree, err := part.BuildTreeTx(GetDbContext(ctx), false, true)
	if err != nil {
		return err
	}
	return tree.Delete(GetDbContext(ctx), key)
}

// ensureUniqueRow checks row against the primary key and every unique index,
// skipping the stored row at skipKey. A conflict returns sql.UniqueKeyError
// carrying the existing row, which INSERT ... ON DUPLICATE KEY UPDATE and
// REPLACE use to update or delete it.
func (i *TableEditor) ensureUniqueRow(sqlCtx *sql.Context, row sql.Row, skipKey []byte) error {
	// Check the primary tree at the encoded row identity, excluding a replaced row.
	pkOrds := i.t.schema.PkOrdinals
	if len(pkOrds) != 0 {
		key, err := i.t.rowKey(sqlCtx, row)
		if err != nil {
			return err
		}
		if !bytes.Equal(key, skipKey) {
			existing, err := i.t.rowAtKey(sqlCtx, key)
			if err != nil {
				return err
			}
			if existing != nil {
				return sql.NewUniqueKeyErr(formatKey(row, pkOrds), true, existing)
			}
		}
	}

	// Probe each non-null unique secondary key once.
	for _, index := range i.t.root.GetIndexes() {
		if !index.GetUnique() {
			continue
		}
		key, unique, err := i.t.secondaryKey(sqlCtx, index, row, nil)
		if err != nil {
			return err
		}
		if !unique {
			continue
		}
		tree, err := i.t.indexTree(sqlCtx, index, false)
		if err != nil {
			return err
		}
		primary, found, err := tree.Get(GetDbContext(sqlCtx), key)
		tree.Discard()
		if err != nil {
			return err
		}
		if !found || bytes.Equal(primary, skipKey) {
			continue
		}

		// Duplicate-key updates need the existing row, not just its identity.
		existing, err := i.t.rowAtKey(sqlCtx, primary)
		if err != nil {
			return err
		}
		ordinals, err := i.t.indexColumnOrdinals(index.GetColumns())
		if err != nil {
			return err
		}
		return sql.NewUniqueKeyErr(formatKey(row, ordinals), false, existing)
	}
	return nil
}

// formatKey formats the key columns of row for a duplicate key message.
func formatKey(row sql.Row, ordinals []int) string {
	vals := make([]string, len(ordinals))
	for idx, ord := range ordinals {
		vals[idx] = errors.Errorf("%v", row[ord]).Error()
	}
	return "[" + strings.Join(vals, ",") + "]"
}

// findRowKey resolves the stored row through its primary key or keyless prefix.
func (i *TableEditor) findRowKey(ctx *sql.Context, row sql.Row) ([]byte, error) {
	// Encode the row identity and choose its one storage partition.
	key, err := i.t.rowKey(ctx, row)
	if err != nil {
		return nil, err
	}
	keyless := len(i.t.schema.PkOrdinals) == 0
	partitionKey := key
	if keyless {
		partitionKey = binary.BigEndian.AppendUint64(bytes.Clone(key), 0)
	}
	part, err := i.t.SelectPartition(partitionKey)
	if err != nil {
		return nil, err
	}

	// A keyless mutation selects one duplicate without decoding unrelated rows.
	if keyless {
		tree, err := part.BuildTreeTx(GetDbContext(ctx), false, false)
		if err != nil {
			return nil, err
		}
		defer tree.Discard()
		it := tree.Iterate(GetDbContext(ctx), key, true, false)
		defer it.Close()
		if !it.Next() {
			if err := it.Err(); err != nil {
				return nil, err
			}
			return nil, sql.ErrDeleteRowNotFound.New()
		}
		return bytes.Clone(it.Key()), nil
	}

	// Preserve the editor contract that the supplied old row must still exist.
	existing, err := i.t.rowAtKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, sql.ErrDeleteRowNotFound.New()
	}
	equal, err := row.Equals(ctx, existing, i.t.schema.Schema)
	if err != nil {
		return nil, err
	}
	if !equal {
		return nil, sql.ErrDeleteRowNotFound.New()
	}
	return key, nil
}

// SetAutoIncrementValue sets a new AUTO_INCREMENT value.
func (i *TableEditor) SetAutoIncrementValue(sqlCtx *sql.Context, val uint64) error {
	// Use the SQL operation context when storing the auto-increment value.
	cctx := i.ctx
	if sqlCtx != nil && sqlCtx.Context != nil {
		cctx = sqlCtx.Context
	}

	// Persist the table auto-increment value before updating its cached value.
	err := i.t.root.StoreAutoIncrVal(cctx, i.t.bcs, i.buildBlobOpts, val)
	if err != nil {
		return err
	}

	// Keep the loaded table value consistent with the persisted root.
	i.t.autoIncVal = val
	return nil
}

// AcquireAutoIncrementLock relies on the enclosing transaction's exclusive write lock.
func (i *TableEditor) AcquireAutoIncrementLock(ctx *sql.Context) (func(), error) {
	return func() {}, nil
}

// DiscardChanges is called if a statement encounters an error, and all current
// changes since the statement beginning should be discarded.
func (i *TableEditor) DiscardChanges(ctx *sql.Context, errorEncountered error) error {
	// Restore only failed statements whose edits were journaled.
	if errorEncountered == nil || !i.statement {
		return nil
	}
	i.statement = false

	// Restore row and index identities even when cancellation caused the failure.
	if ctx == nil {
		ctx = sql.NewContext(i.ctx)
	}
	ctx = ctx.WithContext(context.WithoutCancel(ctx.Context))
	for _, change := range slices.Backward(i.changes) {
		if change.after != nil {
			if err := i.removeRow(ctx, change.after, change.newKey); err != nil {
				return err
			}
		}
		if change.before != nil {
			if err := i.storeRow(ctx, change.before, change.oldKey); err != nil {
				return err
			}
		}
	}
	i.changes = nil

	// Restore counters without cloning or retaining the full table cursor graph.
	i.t.root.RowNonce = i.statementNonce
	i.t.bcs.SetBlock(i.t.root, true)
	if i.t.autoIncIdx != 0 {
		return i.SetAutoIncrementValue(ctx, i.statementAutoIncrement)
	}
	return nil
}

// StatementComplete releases the successful statement's rollback journal.
func (i *TableEditor) StatementComplete(ctx *sql.Context) error {
	i.statement = false
	i.changes = nil
	return nil
}

// Close releases the editor; writes already belong to the enclosing transaction.
func (i *TableEditor) Close(sqlCtx *sql.Context) error {
	return nil
}

// _ is a type assertion
var (
	_ sql.AutoIncrementSetter = (*TableEditor)(nil)
	_ sql.RowInserter         = (*TableEditor)(nil)
	_ sql.RowUpdater          = (*TableEditor)(nil)
	_ sql.RowDeleter          = (*TableEditor)(nil)
	_ sql.RowReplacer         = (*TableEditor)(nil)
)
