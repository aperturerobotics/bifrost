package kvtx

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
	hstore "github.com/s4wave/spacewave/db/store"
)

// publicationSnapshotStore is a store whose block reads see an empty snapshot
// taken before any publication applies. Unlike the in-memory store it never
// blocks a read on an open write transaction, so a test reads through the
// Volume while a group is mid-flight. GetStoredBlock reports a block with
// known refs, marking a read served by the store.
type publicationSnapshotStore struct {
	hstore.Store
}

// BeginReadOperation returns an empty snapshot of the store.
func (s publicationSnapshotStore) BeginReadOperation(context.Context) (block.StoreOps, func(), error) {
	return block.NopStoreOps{}, func() {}, nil
}

// GetBlock reports every block as missing.
func (s publicationSnapshotStore) GetBlock(context.Context, *block.BlockRef) ([]byte, bool, error) {
	return nil, false, nil
}

// GetStoredBlock reports every block as stored with known refs.
func (s publicationSnapshotStore) GetStoredBlock(context.Context, *block.BlockRef) (*block.StoredBlock, error) {
	return &block.StoredBlock{RefsKnown: true}, nil
}

// GetBlockExists reports every block as missing.
func (s publicationSnapshotStore) GetBlockExists(context.Context, *block.BlockRef) (bool, error) {
	return false, nil
}

// GetBlockExistsBatch reports every block as missing.
func (s publicationSnapshotStore) GetBlockExistsBatch(_ context.Context, refs []*block.BlockRef) ([]bool, error) {
	return make([]bool, len(refs)), nil
}

// StatBlock reports every block as missing.
func (s publicationSnapshotStore) StatBlock(context.Context, *block.BlockRef) (*block.BlockStat, error) {
	return nil, nil
}
