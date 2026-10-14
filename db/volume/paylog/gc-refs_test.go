//go:build !js

package paylog

import (
	"testing"

	block_gc "github.com/s4wave/spacewave/db/block/gc"
	"github.com/s4wave/spacewave/db/volume/device"
)

// TestGCStoreOpsJournaledLeaf checks that journaled ownership identifies a
// tracked leaf before its incoming edge reaches the graph.
func TestGCStoreOpsJournaledLeaf(t *testing.T) {
	for _, engine := range engines {
		t.Run(engine.name, func(t *testing.T) {
			// Write a leaf whose GC edges are journaled in the payload store.
			ctx := t.Context()
			store := openStore(t, engine, device.NewMemory())
			graph := newGraph(store)
			gcStore := block_gc.NewGCStoreOps(store, graph)
			gcStore.SetWALAppender(store)
			ref, _, err := gcStore.PutBlock(ctx, []byte("journaled leaf"), nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := gcStore.FlushPending(ctx); err != nil {
				t.Fatal(err)
			}

			// Require a known leaf before and after the journal reaches the graph.
			for _, replayed := range []bool{false, true} {
				if replayed {
					if _, err := store.ReplayWAL(ctx, graph); err != nil {
						t.Fatal(err)
					}
				}
				stored, err := gcStore.GetStoredBlock(ctx, ref)
				if err != nil || stored == nil || !stored.RefsKnown || len(stored.Refs) != 0 {
					t.Fatalf("replayed=%v: leaf = %+v/%v", replayed, stored, err)
				}
			}
		})
	}
}
