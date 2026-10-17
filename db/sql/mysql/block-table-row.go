package mysql

import (
	"context"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/block/sbset"
)

// UnmarshalTableRow unmarshals the TableRow block, returning nil for an empty
// cursor.
func UnmarshalTableRow(ctx context.Context, bcs *block.Cursor) (*TableRow, error) {
	return block.UnmarshalBlock[TableRow](ctx, bcs)
}

// ApplySubBlock applies a sub-block change with a field id. The column set
// edits the row in place, so there is nothing to apply.
func (r *TableRow) ApplySubBlock(id uint32, next block.SubBlock) error {
	return nil
}

// GetSubBlocks returns all constructed sub-blocks by ID.
// May return nil, and values may also be nil.
func (r *TableRow) GetSubBlocks() map[uint32]block.SubBlock {
	m := make(map[uint32]block.SubBlock)
	m[1] = newTableRowColumnSetContainer(r, nil)
	return m
}

// GetSubBlockCtor returns a function which creates or returns the existing
// sub-block at reference id. Can return nil to indicate invalid reference id.
func (r *TableRow) GetSubBlockCtor(id uint32) block.SubBlockCtor {
	switch id {
	case 1:
		return func(create bool) block.SubBlock {
			return newTableRowColumnSetContainer(r, nil)
		}
	}
	return nil
}

// FetchSqlRow converts the encoded columns into a sql.Row of Go values, which
// the caller checks against its schema.
func (r *TableRow) FetchSqlRow(ctx context.Context, bcs *block.Cursor) (sql.Row, error) {
	// Size the SQL row to the stored columns.
	rowCols := r.GetColumns()
	cols := make(sql.Row, len(rowCols))

	// Decode each stored column into its SQL row position. Only a chunked
	// blob reads through a cursor, so the column set is followed at the
	// first one and rows of scalars and inline blobs follow none.
	var colSet *sbset.SubBlockSet
	for i, col := range rowCols {
		var colcs *block.Cursor
		if col.readsThroughCursor() {
			if colSet == nil {
				colSet = newTableRowColumnSetContainer(r, bcs)
			}
			_, colcs = colSet.Get(i)
		}
		val, err := col.FetchSqlColumn(ctx, colcs)
		if err != nil {
			return nil, err
		}
		cols[i] = val
	}
	return cols, nil
}

// _ is a type assertion
var (
	_ block.Block              = (*TableRow)(nil)
	_ block.BlockWithSubBlocks = (*TableRow)(nil)
)
