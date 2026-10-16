package blocktype

import "github.com/s4wave/spacewave/db/block"

// BlockType provides construction and identification for a block.Block type.
type BlockType interface {
	// Constructor builds a new zero-value block.Block instance.
	Constructor() block.Block

	// GetBlockTypeID returns the unique identifier for this block type.
	//
	// Format: "github.com/s4wave/spacewave/db/block/mock.Root"
	GetBlockTypeID() string

	// MatchesBlockType checks if a block.Block is of this type.
	MatchesBlockType(b block.Block) bool
}

// blockType implements BlockType for blocks of type *T.
type blockType[T any, PT block.Ptr[T]] struct {
	typeID string
}

// New builds the BlockType of *T with the unique identifier typeID, such as
// "github.com/s4wave/spacewave/db/block/mock.Root".
func New[T any, PT block.Ptr[T]](typeID string) BlockType {
	return &blockType[T, PT]{typeID: typeID}
}

// Constructor builds a new zero-value block.Block instance.
func (t *blockType[T, PT]) Constructor() block.Block {
	return PT(new(T))
}

// GetBlockTypeID returns the unique identifier for this block type.
func (t *blockType[T, PT]) GetBlockTypeID() string {
	return t.typeID
}

// MatchesBlockType checks if a block.Block is of this type.
func (t *blockType[T, PT]) MatchesBlockType(b block.Block) bool {
	_, ok := b.(PT)
	return ok
}
