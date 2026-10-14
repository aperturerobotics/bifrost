package testbed

import (
	"errors"
	"slices"
	"strconv"
	"testing"

	"github.com/s4wave/spacewave/db/world"
)

// errWindowRead ends WatchChanges after one report.
var errWindowRead = errors.New("read one change window")

// TestWatchChanges checks that a World without a changelog reports the
// objects and quads of a batch commit through WatchChanges.
func TestWatchChanges(t *testing.T) {
	// Start a World and watch from its current head.
	ctx := t.Context()
	tb := MustDefault(t, ctx)
	after, err := tb.Engine.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Open a write transaction.
	tx, err := tb.Engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()

	// Create objects linked from the first one and commit them together.
	keys := make([]string, 12)
	for i := range keys {
		keys[i] = "obj-" + strconv.Itoa(i)
		obj, err := tx.CreateObject(ctx, keys[i], nil)
		world.ReleaseObjectState(obj)
		if err != nil {
			t.Fatal(err)
		}
		if i != 0 {
			if err := tx.SetGraphQuad(ctx, world.NewGraphQuadWithKeys(keys[0], "<link>", keys[i], "")); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// The feed reports every key and quad.
	_, changes := readChangeWindow(t, tb.Engine, after)
	if changes.Unknown || !slices.Equal(slices.Sorted(slices.Values(changes.Keys)), slices.Sorted(slices.Values(keys))) || len(changes.Quads) != len(keys)-1 {
		t.Fatalf("changes = %+v", changes)
	}
}

// readChangeWindow reads the first change window of eng after afterSeqno.
func readChangeWindow(t *testing.T, eng world.Engine, afterSeqno uint64) (uint64, *world.ChangeSet) {
	// Keep the first report and end the watch.
	t.Helper()
	var seqno uint64
	var changes *world.ChangeSet
	err := world.WatchChanges(t.Context(), eng, afterSeqno, func(s uint64, cs *world.ChangeSet) error {
		seqno, changes = s, cs
		return errWindowRead
	})
	if !errors.Is(err, errWindowRead) {
		t.Fatal(err)
	}
	return seqno, changes
}
