//go:build !js

package sobject_world_engine_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/s4wave/spacewave/db/world"
)

// TestLocalWorldWatchChanges checks that a SharedObject World reports the keys
// each commit changed, so watchers wake only for the reads a change touches.
func TestLocalWorldWatchChanges(t *testing.T) {
	// Start a World and note its revision.
	ctx := t.Context()
	w := startLocalWorld(t, t.TempDir(), nil)
	w.create(t, "watch/base")
	after, err := w.eng.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Commit one object and read the changes after the noted revision.
	w.create(t, "watch/changed")
	errRead := errors.New("read one change set")
	var changes *world.ChangeSet
	err = world.WatchChanges(ctx, w.eng, after, func(_ uint64, cs *world.ChangeSet) error {
		changes = cs
		return errRead
	})
	if !errors.Is(err, errRead) {
		t.Fatal(err)
	}
	if changes.Unknown || !slices.Contains(changes.Keys, "watch/changed") {
		t.Fatalf("changes after one commit: %+v, want the key watch/changed", changes)
	}
}
