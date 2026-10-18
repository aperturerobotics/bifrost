//go:build !js

package sobject_world_engine_test

import (
	"context"
	"errors"
	"testing"
	"time"

	world_block "github.com/s4wave/spacewave/db/world/block"
)

// TestWorldReadSurvivesRelease checks that stopping a mounted engine waits for
// its open read, rather than discarding the snapshot while the caller reads it.
func TestWorldReadSurvivesRelease(t *testing.T) {
	// Open a read through the same World interface the Space catalog uses.
	w, so := startGatedWorld(t, t.TempDir(), nil)
	w.create(t, "catalog")
	read, err := w.eng.NewTransaction(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Discard()

	// Stop the controller as a mount release does. It must keep the block
	// engine and SharedObject alive until the transaction releases its snapshot.
	stopped := make(chan struct{})
	go func() {
		so.stop()
		close(stopped)
	}()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if err := so.ctrl.WaitWorldEngineReplaced(ctx, w.eng); err != nil {
		t.Fatalf("wait for engine withdrawal: %v", err)
	}
	select {
	case <-stopped:
		t.Fatal("engine stopped with an open read transaction")
	case <-time.After(50 * time.Millisecond):
	}
	if found, err := read.HasObject(t.Context(), "catalog"); err != nil || !found {
		t.Fatalf("read during release: found=%v err=%v", found, err)
	}

	// Discard twice: cleanup is idempotent and lets the mount finish releasing.
	read.Discard()
	read.Discard()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("engine did not stop after the read discarded")
	}
	if _, err := w.eng.NewTransaction(t.Context(), false); !errors.Is(err, world_block.ErrEngineClosed) {
		t.Fatalf("transaction after release: %v, want ErrEngineClosed", err)
	}

	// Remount the same World and read the catalog through the replacement.
	so.serve(t)
	engine, err := so.ctrl.GetWorldEngine(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	replacement, err := engine.NewTransaction(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer replacement.Discard()
	if found, err := replacement.HasObject(t.Context(), "catalog"); err != nil || !found {
		t.Fatalf("read after remount: found=%v err=%v", found, err)
	}
}
