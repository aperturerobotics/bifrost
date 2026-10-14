package world_block

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
	"github.com/s4wave/spacewave/db/world"
)

// errWindowRead ends WatchChanges after one report.
var errWindowRead = errors.New("read one change window")

// TestWatchChanges checks that the change feed compares the head a watcher saw
// against the current head and lists exactly the touched objects and quads.
func TestWatchChanges(t *testing.T) {
	// Create the objects the changes touch and one they leave alone.
	ctx := t.Context()
	e := newRetirementTestEngine(t, ctx)
	ws := world.NewEngineWorldState(e, true)
	createKeyWatchTestObject(t, ctx, ws, "watch/kept")
	createKeyWatchTestObject(t, ctx, ws, "watch/deleted")
	createKeyWatchTestObject(t, ctx, ws, "watch/linked")
	createKeyWatchTestObject(t, ctx, ws, "watch/untouched")

	// Link two of them with the quad the changes delete.
	unlinked := world.NewGraphQuadWithKeys("watch/kept", "<old>", "watch/linked", "")
	if err := ws.SetGraphQuad(ctx, unlinked); err != nil {
		t.Fatal(err)
	}

	// Start the feed and settle on a retained head.
	after := settleWatch(t, ctx, e, ws)

	// Create and delete an object, and set and delete a quad, over several
	// commits.
	createKeyWatchTestObject(t, ctx, ws, "watch/created")
	if _, err := ws.DeleteObject(ctx, "watch/deleted"); err != nil {
		t.Fatal(err)
	}
	linked := world.NewGraphQuadWithKeys("watch/kept", "<new>", "watch/created", "")
	if err := ws.SetGraphQuad(ctx, linked); err != nil {
		t.Fatal(err)
	}
	if err := ws.DeleteGraphQuad(ctx, unlinked); err != nil {
		t.Fatal(err)
	}

	// Expect one window covering every commit.
	_, changes := readChangeWindow(t, ctx, e, after)
	if changes.Unknown {
		t.Fatal("changes after retained head are unknown")
	}
	for _, key := range []string{"watch/created", "watch/deleted"} {
		if !slices.Contains(changes.Keys, key) {
			t.Fatalf("changed keys %q lack %q", changes.Keys, key)
		}
	}
	if slices.Contains(changes.Keys, "watch/untouched") {
		t.Fatalf("changed keys %q include an untouched object", changes.Keys)
	}
	checkQuads(t, changes.Quads, linked, unlinked)
}

// TestWatchChangesImportedRoot checks that a head installed from outside the
// Engine, as a SharedObject replay or a snapshot import installs it, yields a
// precise ChangeSet.
func TestWatchChangesImportedRoot(t *testing.T) {
	// Start the feed and settle on a retained head.
	ctx := t.Context()
	e := newRetirementTestEngine(t, ctx)
	ws := world.NewEngineWorldState(e, true)
	after := settleWatch(t, ctx, e, ws)

	// Import a World with more operations than the head and install it.
	stage, err := e.StageWorldState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer stage.Release()
	keys := []string{"import/a", "import/b", "import/c"}
	quads := []world.GraphQuad{
		world.NewGraphQuadWithKeys("import/a", "<p>", "import/b", ""),
		world.NewGraphQuadWithKeys("import/b", "<p>", "import/c", ""),
	}
	ref, err := ImportSnapshot(ctx, stage, func(yield func(string, block.Block) bool) {
		for _, key := range keys {
			if !yield(key, block_mock.NewExample(key)) {
				return
			}
		}
	}, quads)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.SetRootRef(ctx, ref); err != nil {
		t.Fatal(err)
	}

	// Expect the objects of both Worlds and the imported quads.
	_, changes := readChangeWindow(t, ctx, e, after)
	if changes.Unknown {
		t.Fatal("changes to an imported root are unknown")
	}
	want := slices.Sorted(slices.Values(append([]string{"watch/settle", "watch/start"}, keys...)))
	if got := slices.Sorted(slices.Values(changes.Keys)); !slices.Equal(got, want) {
		t.Fatalf("changed keys %q, want %q", got, want)
	}
	checkQuads(t, changes.Quads, quads...)
}

// TestWatchChangesEvicted checks that a watcher whose head the Engine no
// longer retains sees unknown changes.
func TestWatchChangesEvicted(t *testing.T) {
	// Settle on a retained head, then publish past the retention bound.
	ctx := t.Context()
	e := newRetirementTestEngine(t, ctx)
	ws := world.NewEngineWorldState(e, true)
	after := settleWatch(t, ctx, e, ws)
	for i := range changeWatchHeads {
		createKeyWatchTestObject(t, ctx, ws, "evict/"+strconv.Itoa(i))
	}

	// Expect unknown changes since the evicted head.
	if _, changes := readChangeWindow(t, ctx, e, after); !changes.Unknown {
		t.Fatalf("changes since an evicted head = %+v, want unknown", changes)
	}
}

// settleWatch starts the change feed and returns a seqno whose head the
// Engine retains. Heads published before the feed starts give unknown
// changes.
func settleWatch(t *testing.T, ctx context.Context, e *Engine, ws world.WorldState) uint64 {
	// Start the feed at a head published before it.
	t.Helper()
	after, err := e.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}
	createKeyWatchTestObject(t, ctx, ws, "watch/start")
	after, changes := readChangeWindow(t, ctx, e, after)
	if !changes.Unknown {
		t.Fatalf("changes before the feed started = %+v, want unknown", changes)
	}

	// The head the feed started at is the floor, so its window is unknown.
	createKeyWatchTestObject(t, ctx, ws, "watch/settle")
	after, changes = readChangeWindow(t, ctx, e, after)
	if !changes.Unknown {
		t.Fatalf("changes since the floor = %+v, want unknown", changes)
	}
	return after
}

// readChangeWindow reads the first change window after afterSeqno.
func readChangeWindow(t *testing.T, ctx context.Context, e *Engine, afterSeqno uint64) (uint64, *world.ChangeSet) {
	// Keep the first report and end the watch.
	t.Helper()
	var seqno uint64
	var changes *world.ChangeSet
	err := e.WatchChanges(ctx, afterSeqno, func(s uint64, cs *world.ChangeSet) error {
		seqno, changes = s, cs
		return errWindowRead
	})
	if !errors.Is(err, errWindowRead) {
		t.Fatal(err)
	}
	return seqno, changes
}

// checkQuads checks that got holds exactly the want quads in any order.
func checkQuads(t *testing.T, got []world.GraphQuad, want ...world.GraphQuad) {
	t.Helper()
	format := func(quads []world.GraphQuad) []string {
		out := make([]string, len(quads))
		for i, q := range quads {
			out[i] = q.GetSubject() + " " + q.GetPredicate() + " " + q.GetObj() + " " + q.GetLabel()
		}
		slices.Sort(out)
		return out
	}
	if g, w := format(got), format(want); !slices.Equal(g, w) {
		t.Fatalf("changed quads %q, want %q", g, w)
	}
}
