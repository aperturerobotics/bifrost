//go:build !js

package sobject_world_engine_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/s4wave/spacewave/db/world"
)

// errWindowRead ends WatchChanges after one report.
var errWindowRead = errors.New("read one change window")

// TestLocalWorldWatchChanges checks that a SharedObject World reports the keys
// each commit changed, so watchers wake only for the reads a change touches.
func TestLocalWorldWatchChanges(t *testing.T) {
	// Start a World and settle the feed on a retained head. Heads published
	// before the feed starts report unknown changes.
	ctx := t.Context()
	w := startLocalWorld(t, t.TempDir(), nil)
	after, err := w.eng.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"watch/start", "watch/settle"} {
		w.create(t, key)
		after, _ = readChangeWindow(t, w.eng, after)
	}

	// Commit one object and expect exactly its key.
	w.create(t, "watch/changed")
	_, changes := readChangeWindow(t, w.eng, after)
	if changes.Unknown || !slices.Equal(changes.Keys, []string{"watch/changed"}) {
		t.Fatalf("changes after one commit: %+v, want the key watch/changed", changes)
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
