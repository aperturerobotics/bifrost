package publisher

import (
	"slices"
	"testing"

	"github.com/s4wave/spacewave/core/release"
	"github.com/s4wave/spacewave/db/block"
	block_gc "github.com/s4wave/spacewave/db/block/gc"
	"github.com/s4wave/spacewave/db/world"
	world_testbed "github.com/s4wave/spacewave/db/world/testbed"
)

// TestCollectBlocksUnknownRefs checks that release export recovers the refs of
// an untracked World root from its decoded block.
func TestCollectBlocksUnknownRefs(t *testing.T) {
	// Commit a World with a nonempty graph in the real in-memory testbed.
	ctx := t.Context()
	tb := world_testbed.MustDefault(t, ctx)
	tx, err := tb.Engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()

	// Create the quad's objects before linking them in the committed graph.
	for _, key := range []string{"source", "target"} {
		object, err := tx.CreateObject(ctx, key, nil)
		world.ReleaseObjectState(object)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.SetGraphQuad(ctx, world.NewGraphQuadWithKeys("source", "links", "target", "")); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// Read the complete release closure before removing the root's GC edges.
	head, before, err := collectBlocks(ctx, tb.Engine, &release.ReleaseMetadata{})
	if err != nil || head == nil || len(before) == 0 {
		t.Fatalf("initial closure = %v/%v/%v", head, before, err)
	}
	graph := tb.Volume.GetRefGraph()
	node := block_gc.BlockIRI(head.GetRootRef())
	outgoing, err := graph.GetOutgoingRefs(ctx, node)
	if err != nil || len(outgoing) == 0 {
		t.Fatalf("root edges = %v/%v, want children", outgoing, err)
	}
	incoming, err := graph.GetIncomingRefs(ctx, node)
	if err != nil {
		t.Fatal(err)
	}

	// Model a root stored without tracking, without changing its block bytes:
	// drop its child list and its owner edges.
	if _, err := graph.RemoveNodeRefs(ctx, node, false); err != nil {
		t.Fatal(err)
	}
	for _, owner := range incoming {
		if err := graph.RemoveRef(ctx, owner, node); err != nil {
			t.Fatal(err)
		}
	}

	// Export the same closure with the typed root's own refs, not a false leaf.
	_, after, err := collectBlocks(ctx, tb.Engine, &release.ReleaseMetadata{})
	if err != nil || len(after) != len(before) {
		t.Fatalf("untracked closure length = %d/%v, want %d", len(after), err, len(before))
	}
	for i, entry := range before {
		got := after[i]
		if !got.ref.EqualsRef(entry.ref) || !got.stored.RefsKnown || len(got.stored.Refs) != len(entry.stored.Refs) {
			t.Fatalf("untracked closure block %d = %+v, want %+v", i, got, entry)
		}
		for _, ref := range entry.stored.Refs {
			if !slices.ContainsFunc(got.stored.Refs, func(candidate *block.BlockRef) bool { return candidate.EqualsRef(ref) }) {
				t.Fatalf("untracked block %d refs = %v, want %v", i, got.stored.Refs, ref)
			}
		}
	}
}
