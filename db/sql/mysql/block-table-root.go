package mysql

import (
	"context"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/block/blob"
)

// NewTableRootBlock constructs an empty table root block.
func NewTableRootBlock() block.Block {
	return &TableRoot{}
}

// LoadTableRoot follows the database root cursor and may return nil.
func LoadTableRoot(ctx context.Context, cursor *block.Cursor) (*TableRoot, error) {
	// Decode the table root block from the supplied cursor.
	ni, err := cursor.Unmarshal(ctx, NewTableRootBlock)
	if err != nil {
		return nil, err
	}

	// Require a present table root before validating its contents.
	niv, ok := ni.(*TableRoot)
	if !ok || niv == nil {
		return nil, nil
	}

	// Validate the decoded table root before returning it.
	if err := niv.Validate(); err != nil {
		return nil, err
	}
	return niv, nil
}

// Validate checks the table schema, partitions, indexes and auto-increment value.
func (r *TableRoot) Validate() error {
	// Validate the table schema before its partitions and indexes.
	if err := r.GetTableSchema().Validate(); err != nil {
		return errors.Wrap(err, "schema")
	}

	// Validate every table storage partition.
	for i, pt := range r.GetTablePartitions() {
		if err := pt.Validate(); err != nil {
			return errors.Wrapf(err, "table_partitions[%d]", i)
		}
	}

	// Require names and columns for every stored table index.
	for i, index := range r.GetIndexes() {
		if index.GetName() == "" {
			return errors.Errorf("indexes[%d]: empty name", i)
		}
		if len(index.GetColumns()) == 0 {
			return errors.Errorf("indexes[%d]: empty columns", i)
		}
		for j, col := range index.GetColumns() {
			if col.GetName() == "" {
				return errors.Errorf("indexes[%d]: columns[%d]: empty name", i, j)
			}
		}
		if err := index.GetKeyValue().Validate(); err != nil {
			return errors.Wrapf(err, "indexes[%d]", i)
		}
	}

	// Locate the table auto-increment column.
	var autoIncrIdx int
	for i, c := range r.GetTableSchema().GetColumns() {
		if c.GetAutoIncrement() {
			autoIncrIdx = i + 1
			break
		}
	}

	// Validate the stored auto-increment value when present.
	autoIncrVal := r.GetAutoIncrVal()
	if autoIncrVal != nil {
		if err := autoIncrVal.Validate(); err != nil {
			return errors.Wrap(err, "auto_incr_val")
		}
	}

	// Require an auto-increment column when the table stores its value.
	hasAutoIncrCol := !autoIncrVal.IsEmpty()
	if autoIncrIdx == 0 && hasAutoIncrCol {
		return errors.New("expected empty auto_incr_val")
	}
	return nil
}

// FetchAutoIncrVal fetches and checks the auto-increment value.
//
// bcs should be located at the table root.
func (r *TableRoot) FetchAutoIncrVal(
	ctx context.Context,
	bcs *block.Cursor,
	expectedType sql.Type,
) (any, sql.ConvertInRange, error) {
	autoIncrVal, err := r.GetAutoIncrVal().FetchSqlColumn(ctx, bcs.FollowSubBlock(4))
	if err != nil {
		return nil, sql.InRange, err
	}
	return expectedType.Convert(sql.NewContext(ctx), autoIncrVal)
}

// StoreAutoIncrVal stores the auto-increment value.
//
// bcs should be located at the table root.
func (r *TableRoot) StoreAutoIncrVal(
	ctx context.Context,
	bcs *block.Cursor,
	buildBlobOpts *blob.BuildBlobOpts,
	val any,
) error {
	// Build the auto-increment column at its table-root sub-block.
	bcs = bcs.FollowSubBlock(4)
	var err error
	r.AutoIncrVal, err = BuildTableColumn(ctx, bcs, buildBlobOpts, val)
	if err != nil {
		return err
	}

	// Mark the updated auto-increment column for persistence.
	bcs.SetBlock(r.AutoIncrVal, true)
	return nil
}

// ApplySubBlock applies a sub-block change with a field id.
func (r *TableRoot) ApplySubBlock(id uint32, next block.SubBlock) error {
	var ok bool
	switch id {
	case 4:
		r.AutoIncrVal, ok = next.(*TableColumn)
		if !ok {
			return block.ErrUnexpectedType
		}
	}
	return nil
}

// GetSubBlocks returns all constructed sub-blocks by ID.
// May return nil, and values may also be nil.
func (r *TableRoot) GetSubBlocks() map[uint32]block.SubBlock {
	return map[uint32]block.SubBlock{
		2: newTableRootPartitionSetContainer(r, nil),
		4: r.GetAutoIncrVal(),
		8: newTableIndexSet(r),
	}
}

// GetSubBlockCtor returns a function which creates or returns the existing
// sub-block at reference id. Can return nil to indicate invalid reference id.
func (r *TableRoot) GetSubBlockCtor(id uint32) block.SubBlockCtor {
	switch id {
	case 8:
		return func(create bool) block.SubBlock {
			return newTableIndexSet(r)
		}
	case 2:
		return func(create bool) block.SubBlock {
			return newTableRootPartitionSetContainer(r, nil)
		}
	case 4:
		return func(create bool) block.SubBlock {
			v := r.GetAutoIncrVal()
			if v == nil && create {
				v = &TableColumn{}
				r.AutoIncrVal = v
			}
			return v
		}
	}
	return nil
}

// _ is a type assertion
var (
	_ block.Block              = (*TableRoot)(nil)
	_ block.BlockWithSubBlocks = (*TableRoot)(nil)
)
