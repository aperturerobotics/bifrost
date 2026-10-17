package block_gc

import (
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_store_test "github.com/s4wave/spacewave/db/block/store/test"
	"github.com/s4wave/spacewave/net/hash"
)

// TestGCStageHoldsCompleteRoots checks a GC wrapper around an existing stage.
func TestGCStageHoldsCompleteRoots(t *testing.T) {
	// Build a root to hold through the GC wrapper.
	ref, err := block.BuildBlockRef([]byte("held root"), &block.PutOpts{HashType: hash.HashType_HashType_SHA256})
	if err != nil {
		t.Fatal(err)
	}

	// Wrap an existing stage: OpenStage bypasses the GC wrapper.
	inner := &block_store_test.HoldingStage{}
	held, err := block.HoldCompleteRoots(t.Context(), NewGCStoreOps(inner, nil), []*block.BlockRef{ref})
	if err != nil {
		t.Fatal(err)
	}
	if len(held) != 1 || !held[0] || len(inner.Held) != 1 || !inner.Held[0].EqualVT(ref) {
		t.Fatalf("held %v, inner stage held %v, want the root held", held, inner.Held)
	}
}
