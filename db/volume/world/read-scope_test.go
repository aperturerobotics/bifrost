package volume_world_test

import (
	"testing"

	"github.com/s4wave/spacewave/db/block"
	transform_all "github.com/s4wave/spacewave/db/block/transform/all"
	volume_world "github.com/s4wave/spacewave/db/volume/world"
	"github.com/s4wave/spacewave/db/world/testbed"
)

// TestWorldVolumeReadScopeKeepsBackingStore exercises cursor reads through a
// Volume whose own metadata and block bytes are stored in a parent World.
func TestWorldVolumeReadScopeKeepsBackingStore(t *testing.T) {
	ctx := t.Context()
	tb, err := testbed.Default(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tb.Release()

	vol, err := volume_world.NewVolumeWithEngine(ctx, tb.Logger, tb.Bus, transform_all.BuildFactorySet(),
		&volume_world.Config{ObjectKey: "nested-volume"}, tb.Engine)
	if err != nil {
		t.Fatal(err)
	}
	defer vol.Close()
	data := []byte("read through the nested volume and its parent World")
	ref, _, err := vol.PutBlock(ctx, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	scoped, release, err := vol.BeginReadOperation(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	_, cursor := block.NewTransaction(vol, nil, ref, nil)
	got, found, err := cursor.Fetch(block.WithReadOperationStore(ctx, scoped))
	if err != nil {
		t.Fatal(err)
	}
	if !found || string(got) != string(data) {
		t.Fatalf("nested block = %q, found %v", got, found)
	}
}
