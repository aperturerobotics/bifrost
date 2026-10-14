package world_block_test

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/db/world"
)

// TestWaitObjectChange checks that WaitObjectChange ends when a seen object is
// created or deleted, through the Engine's key watch.
func TestWaitObjectChange(t *testing.T) {
	// Create one of two objects on a World Engine.
	ctx := t.Context()
	eng, cleanup := setupWorldWriteBench(ctx, t)
	defer cleanup()
	ws := world.NewEngineWorldState(eng, true)
	obj, err := ws.CreateObject(ctx, "wait/present", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	world.ReleaseObjectState(obj)

	// Creating the missing object ends a wait on both.
	seen, err := world.GetObjectRootRefsBatch(ctx, ws, []string{"wait/present", "wait/missing"})
	if err != nil {
		t.Fatal(err.Error())
	}
	waitObjectChangeAfter(ctx, t, ws, seen, func() error {
		obj, err := ws.CreateObject(ctx, "wait/missing", nil)
		world.ReleaseObjectState(obj)
		return err
	})

	// Deleting the present object ends a wait on it alone.
	seen, err = world.GetObjectRootRefsBatch(ctx, ws, []string{"wait/present"})
	if err != nil {
		t.Fatal(err.Error())
	}
	waitObjectChangeAfter(ctx, t, ws, seen, func() error {
		_, err := ws.DeleteObject(ctx, "wait/present")
		return err
	})
}

// waitObjectChangeAfter waits for a change to seen and checks that the wait
// ends only after change runs.
func waitObjectChangeAfter(ctx context.Context, t *testing.T, ws world.WorldState, seen []*world.ObjectRootRef, change func() error) {
	// Start the wait and check that it holds.
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- world.WaitObjectChange(ctx, ws, seen...)
	}()
	select {
	case err := <-done:
		t.Fatalf("wait ended before the change: %v", err)
	default:
	}

	// Make the change and expect the wait to end.
	if err := change(); err != nil {
		t.Fatal(err.Error())
	}
	if err := <-done; err != nil {
		t.Fatal(err.Error())
	}
}
