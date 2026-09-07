package bucket_lookup

import (
	"sync/atomic"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
	"github.com/s4wave/spacewave/db/bucket"
	bucket_mock "github.com/s4wave/spacewave/db/bucket/mock"
)

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
	const depth = 8
	ctx := t.Context()
	source := block_mock.NewMockStore(0)
	destination := block_mock.NewMockStore(0)
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

	src := NewCursor(ctx, nil, nil, nil, source, nil,
		&bucket.ObjectRef{BucketId: "source", RootRef: root},
		&bucket.BucketOpArgs{BucketId: "source"}, nil)
	defer src.Release()
	dest := NewCursor(ctx, nil, nil, nil, destination, nil,
		&bucket.ObjectRef{BucketId: "destination"},
		&bucket.BucketOpArgs{BucketId: "destination"}, nil)
	defer dest.Release()

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
}
