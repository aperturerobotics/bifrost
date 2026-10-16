package bstore_world

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
)

// UnmarshalBlockStoreState unmarshals a BlockStoreState from a block cursor.
func UnmarshalBlockStoreState(ctx context.Context, bcs *block.Cursor) (*BlockStoreState, error) {
	return block.UnmarshalBlock[BlockStoreState](ctx, bcs)
}

// Validate validates the BlockStoreState.
func (i *BlockStoreState) Validate() error {
	return nil
}

// _ is a type assertion
var _ block.Block = (*BlockStoreState)(nil)
