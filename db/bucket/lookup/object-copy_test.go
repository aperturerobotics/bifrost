package bucket_lookup

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/block/byteslice"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
	"github.com/s4wave/spacewave/db/bucket"
	bucket_mock "github.com/s4wave/spacewave/db/bucket/mock"
)

// copyDestination observes the copy's storage boundary and injects failures.
type copyDestination struct {
	// StoreOps supplies real in-memory block storage.
	block.StoreOps
	// batchSizes records the submitted ownership batches.
	batchSizes []int
	// syncs counts completed-copy durability requests.
	syncs int
	// batchErr rejects a destination write batch when set.
	batchErr error
	// syncErr rejects the final durability fence when set.
	syncErr error
}

// TestCopyObjectLargerThanBuffer preserves valid single-block copies that exceed
// the batching budget, rather than waiting forever for impossible capacity.
func TestCopyObjectLargerThanBuffer(t *testing.T) {
	// Store an encoded block larger than the copy's pending-byte budget.
	ctx := t.Context()
	source := block_mock.NewMockStore(0)
	destination := block_mock.NewMockStore(0)
	data := make([]byte, 5<<20)
	root, _, err := source.PutBlock(ctx, data, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Copy between distinct buckets using the raw block constructor.
	src := NewCursor(ctx, nil, nil, nil, source, nil,
		&bucket.ObjectRef{BucketId: "source", RootRef: root},
		&bucket.BucketOpArgs{BucketId: "source"}, nil)
	defer src.Release()
	dest := NewCursor(ctx, nil, nil, nil, destination, nil,
		&bucket.ObjectRef{BucketId: "destination"},
		&bucket.BucketOpArgs{BucketId: "destination"}, nil)
	defer dest.Release()
	ref, err := CopyObjectToBucket(ctx, dest, src, byteslice.NewByteSliceBlock, 1, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	// A completed root must resolve to the entire original payload.
	stored, found, err := destination.GetBlock(ctx, ref.GetRootRef())
	if err != nil || !found || len(stored) != len(data) {
		t.Fatalf("large copy: bytes=%d found=%t err=%v", len(stored), found, err)
	}
}

// PutBlockBatch records each batch before passing it to storage.
func (d *copyDestination) PutBlockBatch(ctx context.Context, entries []*block.PutBatchEntry) error {
	d.batchSizes = append(d.batchSizes, len(entries))
	if d.batchErr != nil {
		return d.batchErr
	}
	return d.StoreOps.PutBlockBatch(ctx, entries)
}

// Sync can fail the final fence independently of successful block writes.
func (d *copyDestination) Sync(ctx context.Context) (bool, error) {
	d.syncs++
	if d.syncErr != nil {
		return false, d.syncErr
	}
	return d.StoreOps.Sync(ctx)
}

// sharedCopyBlock references the same child twice to form a shared subtree.
type sharedCopyBlock struct {
	bucket_mock.Root
}

// GetBlockRefs exposes two edges to the encoded child.
func (b *sharedCopyBlock) GetBlockRefs() (map[uint32]*block.BlockRef, error) {
	ref := b.GetExamplePtr().GetRootRef()
	return map[uint32]*block.BlockRef{1: ref, 2: ref}, nil
}

// GetBlockRefCtor decodes every child with the same shared-edge representation.
func (b *sharedCopyBlock) GetBlockRefCtor(uint32) block.Ctor {
	return func() block.Block { return &sharedCopyBlock{} }
}

// TestCopyObjectPrunesSharedSubtrees checks bounded traversal and complete storage.
func TestCopyObjectPrunesSharedSubtrees(t *testing.T) {
	// Build a graph with two references to each successive child.
	const depth = 8
	ctx := t.Context()
	source := block_mock.NewMockStore(0)
	destination := &copyDestination{StoreOps: block_mock.NewMockStore(0)}
	refs := make([]*block.BlockRef, 0, depth+1)
	var root *block.BlockRef
	for range depth + 1 {
		blk := &sharedCopyBlock{Root: bucket_mock.Root{
			ExamplePtr: &bucket.ObjectRef{BucketId: "child", RootRef: root},
		}}
		data, err := blk.MarshalBlock()
		if err != nil {
			t.Fatal(err)
		}
		root, _, err = source.PutBlock(ctx, data, nil)
		if err != nil {
			t.Fatal(err)
		}
		refs = append(refs, root)
	}

	// Give the copy distinct source and destination bucket identities.
	src := NewCursor(ctx, nil, nil, nil, source, nil,
		&bucket.ObjectRef{BucketId: "source", RootRef: root},
		&bucket.BucketOpArgs{BucketId: "source"}, nil)
	defer src.Release()
	dest := NewCursor(ctx, nil, nil, nil, destination, nil,
		&bucket.ObjectRef{BucketId: "destination"},
		&bucket.BucketOpArgs{BucketId: "destination"}, nil)
	defer dest.Release()

	// Count logical visits while concurrent workers prune shared descendants.
	var visits atomic.Int64
	ref, stats, err := CopyObjectToBucketWithStats(ctx, dest, src,
		func() block.Block { return &sharedCopyBlock{} }, 2, false,
		func(ent *WalkObjectBlocksEntry) (bool, error) {
			if !ent.Ref.GetEmpty() {
				visits.Add(1)
			}
			return true, ent.Err
		})
	if err != nil {
		t.Fatal(err)
	}

	// Verify every distinct block reached a single fenced destination batch.
	if got, want := visits.Load(), int64(2*depth+1); got != want {
		t.Fatalf("visited %d blocks, want %d: shared descendants were revisited", got, want)
	}
	if stats.BlocksCopied != depth+1 || stats.BlocksDeduped != depth {
		t.Fatalf("unexpected copy accounting: %+v", stats)
	}
	if ref.GetBucketId() != "destination" || !ref.GetRootRef().EqualsRef(root) {
		t.Fatalf("unexpected destination ref: %v", ref)
	}
	for _, ref := range refs {
		_, found, err := destination.GetBlock(ctx, ref)
		if err != nil || !found {
			t.Fatalf("copied block %v: found=%v, err=%v", ref, found, err)
		}
	}
	if len(destination.batchSizes) != 1 || destination.batchSizes[0] != depth+1 || destination.syncs != 1 {
		t.Fatalf("copy did not batch and fence its writes: batches=%v syncs=%d", destination.batchSizes, destination.syncs)
	}

	// Existing payloads still need the destination's ownership write. A failed
	// final batch or durability fence must never return a completed root.
	for _, failure := range []string{"batch", "sync"} {
		t.Run(failure, func(t *testing.T) {
			injected := errors.New("destination " + failure + " failed")
			destination.batchErr, destination.syncErr = nil, nil
			if failure == "batch" {
				destination.batchErr = injected
			} else {
				destination.syncErr = injected
			}
			ref, stats, err := CopyObjectToBucketWithStats(ctx, dest, src,
				func() block.Block { return &sharedCopyBlock{} }, 2, false, nil)
			if !errors.Is(err, injected) || ref != nil || stats.BlocksExisting != depth+1 {
				t.Fatalf("copy accepted failed destination: ref=%v stats=%+v err=%v", ref, stats, err)
			}
		})
	}
}
