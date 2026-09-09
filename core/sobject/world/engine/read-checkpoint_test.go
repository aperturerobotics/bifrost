package sobject_world_engine

import (
	"errors"
	"testing"

	"github.com/s4wave/spacewave/core/sobject"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/tx"
	world_block "github.com/s4wave/spacewave/db/world/block"
)

// TestReadCheckpointFreezesWorld verifies the historical root and rejects write transactions.
func TestReadCheckpointFreezesWorld(t *testing.T) {
	ctx := t.Context()
	c, so, head := newProcessTestWorld(t, ctx)
	known := applyTransactionTestObject(t, c, so, head, "known-object")
	snapshot := newTestFinalizationSnapshot(t, &sobject.SORoot{InnerSeqno: 2}, known.GetHeadRef())
	_ = applyTransactionTestObject(t, c, so, known, "later-object")
	engine, release, err := OpenReadCheckpoint(ctx, c.le, c.bus, so, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		release()
		if _, err := engine.NewTransaction(ctx, false); !errors.Is(err, world_block.ErrEngineClosed) {
			t.Fatalf("released checkpoint retained its engine: %v", err)
		}
	}()
	read, err := engine.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Discard()
	if found, err := read.HasObject(ctx, "known-object"); err != nil || !found {
		t.Fatalf("checkpoint lost known object: found=%v err=%v", found, err)
	}
	if found, err := read.HasObject(ctx, "later-object"); err != nil || found {
		t.Fatalf("checkpoint exposed later object: found=%v err=%v", found, err)
	}
	if _, err := engine.NewTransaction(ctx, true); !errors.Is(err, tx.ErrNotWrite) {
		t.Fatalf("checkpoint accepted write transaction: %v", err)
	}
	cursor, err := engine.BuildStorageCursor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cursor.Release()
	if _, _, err := cursor.PutBlock(ctx, []byte("unreferenced history write"), nil); !errors.Is(err, tx.ErrNotWrite) {
		t.Fatalf("checkpoint storage cursor accepted a write: %v", err)
	}
	if err := engine.AccessWorldState(ctx, nil, func(cursor *bucket_lookup.Cursor) error {
		_, _, err := cursor.PutBlock(ctx, []byte("root history write"), nil)
		return err
	}); !errors.Is(err, tx.ErrNotWrite) {
		t.Fatalf("checkpoint World cursor accepted a write: %v", err)
	}
}
