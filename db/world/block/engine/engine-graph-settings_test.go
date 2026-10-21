package world_block_engine_test

import (
	"slices"
	"testing"

	cayley_kv "github.com/aperturerobotics/cayley/graph/kv"
	"github.com/aperturerobotics/cayley/quad"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	world_block "github.com/s4wave/spacewave/db/world/block"
	world_block_engine "github.com/s4wave/spacewave/db/world/block/engine"
)

// TestWorldEngineControllerGraphSettings checks that a local World created
// from a config with graph settings starts with them in its root.
func TestWorldEngineControllerGraphSettings(t *testing.T) {
	// Start the engine with a predicate index.
	indexes := append(slices.Clone(cayley_kv.DefaultQuadIndexes), cayley_kv.QuadIndex{
		Dirs: []quad.Direction{quad.Predicate, quad.Label, quad.Subject},
	})
	settings, err := world_block.NewGraphSettings(indexes)
	if err != nil {
		t.Fatal(err.Error())
	}
	ctx, _, _, startEngine := newEngineTestbed(t)
	eng, engRef := startEngine(true, func(c *world_block_engine.Config) {
		c.GraphSettings = settings
	})
	defer engRef.Release()

	// The World root records the settings.
	err = eng.AccessWorldState(ctx, nil, func(bls *bucket_lookup.Cursor) error {
		// Decode the root and compare its settings.
		_, bcs := bls.BuildTransaction(nil)
		wi, err := bcs.Unmarshal(ctx, world_block.NewWorldBlock)
		if err != nil {
			return err
		}
		if got := wi.(*world_block.World).GetGraphSettings(); !got.EqualVT(settings) {
			t.Errorf("world root graph settings %v", got)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err.Error())
	}
}
