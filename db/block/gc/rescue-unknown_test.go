package block_gc

import (
	"context"
	"slices"
	"testing"

	store_kvtx_inmem "github.com/s4wave/spacewave/db/store/kvtx/inmem"
)

// TestRescueAdoptedUntrackedParent checks that a rescue counts a block adopted
// during a pass that has no child list as unknown, since a parent stored
// untracked after the mark may name a child the mark found unreached, and
// that a catch-up reaches the child once the parent's list is recorded.
func TestRescueAdoptedUntrackedParent(t *testing.T) {
	// A tracked child that nothing owns.
	ctx := context.Background()
	store := store_kvtx_inmem.NewStore()
	rg := NewRefGraph(store)
	t.Cleanup(func() { rg.Close() })
	parent, child := BlockIRI(testBlockRef(t, "untracked-parent")), BlockIRI(testBlockRef(t, "untracked-child"))
	if err := rg.ApplyRefBatch(ctx, []RefEdge{{Subject: NodeUnreferenced, Object: child}}, nil); err != nil {
		t.Fatal(err)
	}

	// Open a pass and mark: the child is unreached.
	if err := rg.BeginBlockPass(ctx); err != nil {
		t.Fatal(err)
	}
	mark, err := rg.MarkBlocks(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// A bucket adopts the parent, stored without a child list, which
	// references the child.
	if err := rg.AddRef(ctx, "bucket", parent); err != nil {
		t.Fatal(err)
	}
	deleteChildList(t, store, parent)
	if _, err := rg.RescueAdopted(ctx, mark); err != nil {
		t.Fatal(err)
	}

	// The rescue reports the parent, so the sweep removes nothing, and
	// lists it, so the sweep can wait for its list. A rescue with nothing
	// logged reads the list again.
	for range 2 {
		if mark.Unknown != 1 || !mark.Awaitable() || !slices.Equal(mark.UnknownSample, []string{parent}) {
			t.Fatalf("unknown %d %v, want the parent", mark.Unknown, mark.UnknownSample)
		}
		if _, err := rg.RescueAdopted(ctx, mark); err != nil {
			t.Fatal(err)
		}
	}

	// Once the parent's references are recorded, a catch-up in read
	// transactions reaches the child through it.
	record := []RefEdge{{Subject: NodeUnreferenced, Object: parent}, {Subject: parent, Object: child}}
	if err := rg.ApplyRefBatch(ctx, record, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := rg.CatchUpAdopted(ctx, mark); err != nil {
		t.Fatal(err)
	}
	if mark.Unknown != 0 || len(mark.UnknownSample) != 0 || !mark.Live.Has(child) {
		t.Fatalf("unknown %d %v, child live %v", mark.Unknown, mark.UnknownSample, mark.Live.Has(child))
	}

	// The rescue then consumes the log without walking it again.
	if n, err := rg.RescueAdopted(ctx, mark); err != nil || n != 0 || mark.logSeen == 0 {
		t.Fatalf("rescue after catch-up added %d, log seen %d, err=%v", n, mark.logSeen, err)
	}
}
