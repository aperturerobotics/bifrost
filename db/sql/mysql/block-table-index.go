package mysql

import (
	"github.com/s4wave/spacewave/db/block"
	kvtx_block "github.com/s4wave/spacewave/db/kvtx/block"
)

// IsNil reports whether the index sub-block is absent.
func (r *TableIndex) IsNil() bool {
	return r == nil
}

// BlockAliasIdentity returns the per-instance token for the inline index.
func (r *TableIndex) BlockAliasIdentity() *block.AliasIdentityToken {
	return &r.unknownFields
}

// ApplySubBlock applies changes to the index tree.
func (r *TableIndex) ApplySubBlock(id uint32, next block.SubBlock) error {
	if id == 5 {
		return block.ApplySubBlock(&r.KeyValue, next)
	}
	return nil
}

// GetSubBlocks returns the index's stored tree.
func (r *TableIndex) GetSubBlocks() map[uint32]block.SubBlock {
	return map[uint32]block.SubBlock{5: r.GetKeyValue()}
}

// GetSubBlockCtor returns the constructor for the index tree.
func (r *TableIndex) GetSubBlockCtor(id uint32) block.SubBlockCtor {
	if id == 5 {
		return kvtx_block.NewKeyValueStoreSubBlockCtor(&r.KeyValue)
	}
	return nil
}

// _ is a type assertion.
var (
	_ block.SubBlock           = (*TableIndex)(nil)
	_ block.BlockWithSubBlocks = (*TableIndex)(nil)
)
