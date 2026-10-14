package unixfs_world

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/s4wave/spacewave/db/unixfs"
	"github.com/s4wave/spacewave/db/world"
)

// failingWorldState fails every object lookup with err.
type failingWorldState struct {
	world.WorldState
	err error
}

// GetObject reports the lookup failure.
func (w failingWorldState) GetObject(ctx context.Context, key string) (world.ObjectState, bool, error) {
	return nil, false, w.err
}

// TestGetProxyCursorReportsLookupError checks that a failed object lookup
// returns its own error, so a canceled lookup is not taken for a missing
// object.
func TestGetProxyCursorReportsLookupError(t *testing.T) {
	// Look up an object through a World that is canceled.
	ws := failingWorldState{err: context.Canceled}
	cursor := NewFSCursor(nil, ws, "test/lookup-error", FSType_FSType_FS_NODE, nil, false)
	t.Cleanup(cursor.Release)

	// The cursor reports the cancellation.
	if _, err := cursor.GetProxyCursor(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestNewFSCursorWithWriterConfirmsObservedRevision checks that a successful
// writer operation remains fenced on the cursor's observed object revision.
func TestNewFSCursorWithWriterConfirmsObservedRevision(t *testing.T) {
	// Construct a writer-backed FSCursor and verify its confirmation callback.
	cursor, writer := NewFSCursorWithWriterContext(
		context.Background(),
		nil,
		nil,
		"test/revision-confirmation",
		FSType_FSType_FS_NODE,
		"",
	)
	t.Cleanup(cursor.Release)
	confirm := writer.confirmFn.Load()
	if confirm == nil {
		t.Fatal("expected writer revision confirmation")
	}

	// A target newer than the observed revision must wait for the context.
	cursor.bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
		cursor.prevObjRev = 4
	})
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (*confirm)(canceled, 5); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected confirmation to wait, got %v", err)
	}

	// An observed target must return even when the context is already canceled.
	cursor.bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
		cursor.prevObjRev = 5
	})
	if err := (*confirm)(canceled, 5); err != nil {
		t.Fatalf("expected observed revision confirmation, got %v", err)
	}
}

// blockingWorldState blocks every object lookup until unblock closes, then
// reports it missing. It closes entered when the first lookup starts.
type blockingWorldState struct {
	world.WorldState
	entered chan struct{}
	unblock chan struct{}
}

// GetObject signals entered and waits for unblock.
func (w *blockingWorldState) GetObject(ctx context.Context, key string) (world.ObjectState, bool, error) {
	close(w.entered)
	<-w.unblock
	return nil, false, nil
}

// TestGetProxyCursorReadsOutsideLock checks that a change callback can be
// added while GetProxyCursor waits on a storage read.
func TestGetProxyCursorReadsOutsideLock(t *testing.T) {
	// Start a proxy resolution that blocks in its object lookup.
	ws := &blockingWorldState{entered: make(chan struct{}), unblock: make(chan struct{})}
	cursor := NewFSCursor(nil, ws, "test/blocking-lookup", FSType_FSType_FS_NODE, nil, false)
	t.Cleanup(cursor.Release)
	t.Cleanup(func() { close(ws.unblock) })
	go func() {
		_, _ = cursor.GetProxyCursor(context.Background())
	}()
	<-ws.entered

	// Add a change callback while the lookup is blocked.
	added := make(chan struct{})
	go func() {
		cursor.AddChangeCb(func(*unixfs.FSCursorChange) bool { return false })
		close(added)
	}()
	select {
	case <-added:
	case <-time.After(5 * time.Second):
		t.Fatal("AddChangeCb waited on the GetProxyCursor storage read")
	}
}
