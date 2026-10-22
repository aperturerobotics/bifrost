package testbed

import (
	"strconv"
	"testing"

	block_gc "github.com/s4wave/spacewave/db/block/gc"
	"github.com/s4wave/spacewave/db/world"
)

// TestWorldBlockPass checks that a block pass over a World that created,
// relinked and deleted objects reaches no block of unknown references and
// sweeps the blocks the replaced states left behind.
func TestWorldBlockPass(t *testing.T) {
	// Start a World on the testbed volume.
	ctx := t.Context()
	tb := MustDefault(t, ctx)

	// Commit rounds that create objects, link them, then delete half.
	for round := range 8 {
		tx, err := tb.Engine.NewTransaction(ctx, true)
		if err != nil {
			t.Fatal(err)
		}
		for i := range 16 {
			key := "obj-" + strconv.Itoa(round) + "-" + strconv.Itoa(i)
			obj, err := tx.CreateObject(ctx, key, nil)
			world.ReleaseObjectState(obj)
			if err != nil {
				t.Fatal(err)
			}
			if i != 0 {
				q := world.NewGraphQuadWithKeys("obj-"+strconv.Itoa(round)+"-0", "<link>", key, "")
				if err := tx.SetGraphQuad(ctx, q); err != nil {
					t.Fatal(err)
				}
			}
		}
		if round != 0 {
			for i := range 8 {
				if _, err := tx.DeleteObject(ctx, "obj-"+strconv.Itoa(round-1)+"-"+strconv.Itoa(i)); err != nil {
					t.Fatal(err)
				}
			}
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}

	// Run a block pass through the volume's collector.
	rg := tb.Volume.GetRefGraph()
	if rg == nil {
		t.Fatal("volume has no ref graph")
	}
	collector := block_gc.NewCollector(rg, tb.Volume, nil)
	stats, err := collector.Collect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("live=%d unreached=%d swept=%d unknown=%d sample=%v", stats.LiveNodes, stats.UnreachedBlocks, stats.BlocksSwept, stats.UnknownBlocks, stats.UnknownSample)
	if stats.UnknownBlocks != 0 {
		t.Fatalf("pass reached %d blocks of unknown references, such as %v", stats.UnknownBlocks, stats.UnknownSample)
	}
	if stats.BlocksSwept == 0 {
		t.Fatal("pass swept no block")
	}
}
