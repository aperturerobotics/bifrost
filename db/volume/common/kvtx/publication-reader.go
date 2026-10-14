package kvtx

import (
	"bytes"
	"context"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/net/hash"
)

// publicationReader reads a Volume's blocks with its admitted publications
// laid over the store. An entry of a queued or unapplied publication answers
// before the store, so a reader sees admitted blocks before their group
// applies; the entry leaves the read set when its group resolves, applied or
// failed. Entries come from in-process publishers and are verified before
// they are applied, so reads return them without hashing.
type publicationReader struct {
	// writer holds the admitted publications, nil without atomic publication.
	writer *publicationWriter
	// store reads and writes the applied blocks.
	store block.StoreOps
}

// GetHashType returns the store's preferred hash type.
func (r publicationReader) GetHashType() hash.HashType {
	return r.store.GetHashType()
}

// GetSupportedFeatures returns the store's native features.
func (r publicationReader) GetSupportedFeatures() block.StoreFeature {
	return r.store.GetSupportedFeatures()
}

// BeginReadOperation opens a read scope on the store that still reads
// admitted entries first.
func (r publicationReader) BeginReadOperation(ctx context.Context) (block.StoreOps, func(), error) {
	scope, release, err := r.store.BeginReadOperation(ctx)
	if err != nil {
		return nil, nil, err
	}
	return publicationReader{writer: r.writer, store: scope}, release, nil
}

// PutBlock writes a block to the store.
func (r publicationReader) PutBlock(ctx context.Context, data []byte, opts *block.PutOpts) (*block.BlockRef, bool, error) {
	return r.store.PutBlock(ctx, data, opts)
}

// PutBlockBatch writes a batch of blocks to the store.
func (r publicationReader) PutBlockBatch(ctx context.Context, entries []*block.PutBatchEntry) ([]bool, error) {
	return r.store.PutBlockBatch(ctx, entries)
}

// GetBlock reads an admitted entry before the store.
func (r publicationReader) GetBlock(ctx context.Context, ref *block.BlockRef) ([]byte, bool, error) {
	// Read the store when no admitted publication writes the block.
	entry, err := r.writer.find(ref)
	if err != nil {
		return nil, false, err
	}
	if entry == nil {
		return r.store.GetBlock(ctx, ref)
	}

	// Serve the admitted bytes, or absence for a tombstone.
	if entry.Tombstone {
		return nil, false, nil
	}
	return bytes.Clone(entry.Data), true, nil
}

// GetStoredBlock reads a block without its refs, as the store keeps none.
func (r publicationReader) GetStoredBlock(ctx context.Context, ref *block.BlockRef) (*block.StoredBlock, error) {
	return block.GetBlockWithoutRefs(ctx, r, ref)
}

// GetBlockExists reports an admitted entry before the store.
func (r publicationReader) GetBlockExists(ctx context.Context, ref *block.BlockRef) (bool, error) {
	entry, err := r.writer.find(ref)
	if err != nil {
		return false, err
	}
	if entry == nil {
		return r.store.GetBlockExists(ctx, ref)
	}
	return !entry.Tombstone, nil
}

// GetBlockExistsBatch reports admitted entries and probes the store for the
// rest in one batch.
func (r publicationReader) GetBlockExistsBatch(ctx context.Context, refs []*block.BlockRef) ([]bool, error) {
	// Answer each reference an admitted entry writes.
	out := make([]bool, len(refs))
	var missing []*block.BlockRef
	var missingIdx []int
	for i, ref := range refs {
		entry, err := r.writer.find(ref)
		if err != nil {
			return nil, err
		}
		if entry != nil {
			out[i] = !entry.Tombstone
			continue
		}
		missing = append(missing, ref)
		missingIdx = append(missingIdx, i)
	}

	// Probe the store for the remaining references.
	if len(missing) == 0 {
		return out, nil
	}
	found, err := r.store.GetBlockExistsBatch(ctx, missing)
	if err != nil {
		return nil, err
	}
	for i, idx := range missingIdx {
		out[idx] = found[i]
	}
	return out, nil
}

// RmBlock removes a block from the store.
func (r publicationReader) RmBlock(ctx context.Context, ref *block.BlockRef) error {
	return r.store.RmBlock(ctx, ref)
}

// StatBlock reports an admitted entry's size before the store.
func (r publicationReader) StatBlock(ctx context.Context, ref *block.BlockRef) (*block.BlockStat, error) {
	// Stat the store when no admitted publication writes the block.
	entry, err := r.writer.find(ref)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return r.store.StatBlock(ctx, ref)
	}

	// Report the admitted size, or absence for a tombstone.
	if entry.Tombstone {
		return nil, nil
	}
	return &block.BlockStat{Ref: ref.Clone(), Size: int64(len(entry.Data))}, nil
}

// Sync syncs the store.
func (r publicationReader) Sync(ctx context.Context) (bool, error) {
	return r.store.Sync(ctx)
}

// _ is a type assertion
var _ block.StoreOps = publicationReader{}
