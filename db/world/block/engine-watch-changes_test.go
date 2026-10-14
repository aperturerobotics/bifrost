package world_block

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"testing"
	"time"

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

	// Watch from the current head.
	after := headSeqno(t, ctx, e)

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
	// Watch from the head the Engine opens at.
	ctx := t.Context()
	e := newRetirementTestEngine(t, ctx)
	after := headSeqno(t, ctx, e)

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
	if got := slices.Sorted(slices.Values(changes.Keys)); !slices.Equal(got, keys) {
		t.Fatalf("changed keys %q, want %q", got, keys)
	}
	checkQuads(t, changes.Quads, quads...)
}

// TestWatchChangesEvicted checks that a watcher whose head the Engine no
// longer retains sees unknown changes.
func TestWatchChangesEvicted(t *testing.T) {
	// Watch from the current head, then publish past the retention bound.
	ctx := t.Context()
	e := newRetirementTestEngine(t, ctx)
	ws := world.NewEngineWorldState(e, true)
	after := headSeqno(t, ctx, e)
	for i := range changeWatchHeads {
		createKeyWatchTestObject(t, ctx, ws, "evict/"+strconv.Itoa(i))
	}

	// Expect unknown changes since the evicted head.
	if _, changes := readChangeWindow(t, ctx, e, after); !changes.Unknown {
		t.Fatalf("changes since an evicted head = %+v, want unknown", changes)
	}
}

// TestWatchChangesCollectedBase checks that a watcher whose head the volume
// collected sees unknown changes rather than a partial list.
func TestWatchChangesCollectedBase(t *testing.T) {
	// Commit an object and watch from that head.
	f := newSessionFixture(t)
	ctx := t.Context()
	w := sessionWriter(t, f)
	sessionObject(t, w, "collected/first")
	if err := w.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	after := headSeqno(t, ctx, f.engine)
	base := f.engine.GetRootRef().GetRootRef()

	// Supersede the head and collect it.
	w = sessionWriter(t, f)
	sessionObject(t, w, "collected/second")
	if err := w.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	f.collect(t)
	if found, err := f.volume.GetBlockExists(ctx, base); err != nil || found {
		t.Fatalf("superseded head retained: %v %v", found, err)
	}

	// Expect unknown changes since the collected head.
	if _, changes := readChangeWindow(t, ctx, f.engine, after); !changes.Unknown {
		t.Fatalf("changes since a collected head = %+v, want unknown", changes)
	}
}

// peerRootStore pins one root through a peer. A pin of that root that may wait
// for a peer blocks until its context ends, and one that may not reports the
// root unavailable, as a store reading through to an absent peer does. Other
// pins go to the store.
type peerRootStore struct {
	block.StoreOps
	block.RootRetainer
	// peer is the root only a peer holds.
	peer *block.BlockRef
}

// SupportsRootRetention reports that the store pins roots.
func (s *peerRootStore) SupportsRootRetention() bool { return true }

// PinRoot pins ref, asking the absent peer for the peer root.
func (s *peerRootStore) PinRoot(ctx context.Context, ref *block.BlockRef) (func(), error) {
	if !ref.EqualsRef(s.peer) {
		return block.PinRoot(ctx, s.StoreOps, ref)
	}
	if block.GetPeerWait(ctx) != nil {
		return nil, block.ErrUnavailable
	}
	<-ctx.Done()
	return nil, context.Cause(ctx)
}

// TestWatchChangesPeerBase checks that a watcher whose base only an absent
// peer could serve sees unknown changes. Pinning the base must not wait for
// the peer, since it holds the publication lock every reader and writer needs.
func TestWatchChangesPeerBase(t *testing.T) {
	// Commit an object and watch from that head.
	f := newSessionFixture(t)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	w := sessionWriter(t, f)
	sessionObject(t, w, "peer/first")
	if err := w.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	after := headSeqno(t, ctx, f.engine)
	base := f.engine.GetRootRef().GetRootRef()

	// Supersede the head and leave it only on an absent peer.
	w = sessionWriter(t, f)
	sessionObject(t, w, "peer/second")
	if err := w.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	e := f.engine
	locked := e.bcast.Lock()
	e.writeBlockStore = &peerRootStore{StoreOps: e.writeBlockStore, peer: base}
	locked.Unlock()

	// Expect unknown changes since the base before the deadline.
	if _, changes := readChangeWindow(t, ctx, e, after); !changes.Unknown {
		t.Fatalf("changes since a peer-only base = %+v, want unknown", changes)
	}
	if err := ctx.Err(); err != nil {
		t.Fatalf("watch waited for the absent peer: %v", err)
	}
}

// headSeqno returns the seqno of the published head.
func headSeqno(t *testing.T, ctx context.Context, e *Engine) uint64 {
	t.Helper()
	seqno, err := e.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return seqno
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
