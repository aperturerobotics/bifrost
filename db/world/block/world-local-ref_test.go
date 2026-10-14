package world_block_test

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/db/testbed"
	"github.com/s4wave/spacewave/db/world"
	world_block "github.com/s4wave/spacewave/db/world/block"
	"github.com/sirupsen/logrus"
)

// TestBuildWorldStateFromCursorLocalObjectRef checks that an object root in
// the World's own bucket is stored as a block edge of the World DAG, so block
// walks and garbage collection reach the object's content.
func TestBuildWorldStateFromCursorLocalObjectRef(t *testing.T) {
	// Start the storage testbed.
	ctx := context.Background()
	le := logrus.NewEntry(logrus.New())
	tb, err := testbed.NewTestbed(ctx, le)
	if err != nil {
		t.Fatal(err)
	}

	// Build the World on an empty cursor.
	ocs, err := tb.BuildEmptyCursor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer ocs.Release()
	ws, err := world_block.BuildMockWorldState(ctx, le, true, ocs, false)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Discard()

	// Write a root in the World's bucket. An importer holding its external ref
	// creates objects from refs that name the bucket.
	written, err := world_block.BuildMockObject(ctx, ws, "written")
	if err != nil {
		t.Fatal(err)
	}
	defer world.ReleaseObjectState(written)

	// Name the bucket on the written ref and create the imported object.
	bucketID := ocs.GetRefWithOpArgs().GetBucketId()
	external, _, err := written.GetRootRef(ctx)
	if err != nil {
		t.Fatal(err)
	}
	external = external.Clone()
	external.BucketId = bucketID
	obj, err := ws.CreateObject(ctx, "imported", external)
	if err != nil {
		t.Fatal(err)
	}
	defer world.ReleaseObjectState(obj)

	// The stored Object names no bucket and references its root.
	root, err := obj.(*world_block.ObjectState).GetRoot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if stored := root.GetRootRef().GetBucketId(); stored != "" {
		t.Fatalf("stored object root names bucket %q, want a local edge", stored)
	}
	refs, err := root.GetRootRef().GetBlockRefs()
	if err != nil {
		t.Fatal(err)
	}
	if refs[1].GetEmpty() {
		t.Fatal("stored object root is not a block ref")
	}

	// The public ref still names the World's bucket.
	ref, _, err := obj.GetRootRef(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ref.GetBucketId() != bucketID {
		t.Fatalf("object ref bucket = %q, want %q", ref.GetBucketId(), bucketID)
	}
}
