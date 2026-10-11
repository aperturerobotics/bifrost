//go:build !js

package sobject_world_engine_test

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/world"
)

// TestWorldCommitFinishesAfterCancel cancels the commit context as soon as the
// SharedObject holds the operation. The commit has started, so it must finish
// and report that the operation is in the operation set and the World.
func TestWorldCommitFinishesAfterCancel(t *testing.T) {
	// Mount the SharedObject the engine writes to.
	w := startLocalWorld(t, t.TempDir(), nil)
	w.create(t, "cancel/0")
	so, soRef, err := sobject.ExMountSharedObject(t.Context(), w.bus, w.soRef, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer soRef.Release()

	// Count the operations it holds before the write.
	state, relState, err := so.AccessSharedObjectState(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer relState()
	before, err := state.GetValue().GetOperationSet(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	// Cancel once the operation set grows, which is after the engine queued the
	// operation and before it finishes placing it.
	tx := w.newWrite(t, "cancel/1")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	canceled := make(chan error, 1)
	go func() {
		_, err := state.WaitValueWithValidator(t.Context(), func(snap sobject.SharedObjectStateSnapshot) (bool, error) {
			set, err := snap.GetOperationSet(t.Context())
			return err != nil || set.Len() > before.Len(), err
		}, nil)
		cancel()
		canceled <- err
	}()
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit after admission: %v", err)
	}
	if err := <-canceled; err != nil {
		t.Fatal(err)
	}

	// The World after the commit holds the object.
	read, err := w.eng.NewTransaction(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Discard()
	obj, found, err := read.GetObject(t.Context(), "cancel/1")
	world.ReleaseObjectState(obj)
	if err != nil || !found {
		t.Fatalf("object after commit: found %v, %v", found, err)
	}
}
