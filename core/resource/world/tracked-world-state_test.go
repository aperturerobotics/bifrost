//go:build !js

package resource_world_test

import (
	"context"
	"testing"
	"time"

	resource_world "github.com/s4wave/spacewave/core/resource/world"
	"github.com/s4wave/spacewave/db/world"
)

// commitDuringCheckWorldState commits one change to objectKey from inside the
// tracked-revision check, after the check has already read the revisions.
//
// That places the commit in the window between the revision check and the
// seqno the watch waits on, which is the ordering a loaded runner produces on
// its own when the watch routine is descheduled between those two reads.
type commitDuringCheckWorldState struct {
	world.WorldState

	engine    world.Engine
	objectKey string
	committed bool
}

// GetObjectRootRefsBatch reads the revisions, then commits the change once.
func (w *commitDuringCheckWorldState) GetObjectRootRefsBatch(ctx context.Context, keys []string) ([]*world.ObjectRootRef, error) {
	refs, err := world.GetObjectRootRefsBatch(ctx, w.WorldState, keys)
	if err != nil {
		return nil, err
	}
	if !w.committed {
		w.committed = true
		if err := incrementObjectRev(ctx, w.engine, w.objectKey); err != nil {
			return nil, err
		}
	}
	return refs, nil
}

// incrementObjectRev commits one revision of an existing object.
func incrementObjectRev(ctx context.Context, engine world.Engine, key string) error {
	// Open a write transaction for the change.
	tx, err := engine.NewTransaction(ctx, true)
	if err != nil {
		return err
	}
	defer tx.Discard()

	// Increment the object's revision and commit.
	obj, found, err := tx.GetObject(ctx, key)
	defer world.ReleaseObjectState(obj)
	if err != nil {
		return err
	}
	if !found {
		return world.ErrObjectNotFound
	}
	if _, err := obj.IncrementRev(ctx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// createObject commits a new empty object.
func createObject(ctx context.Context, t *testing.T, engine world.Engine, key string) {
	// Create the object in its own write transaction.
	t.Helper()
	tx, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatalf("NewTransaction: %v", err)
	}
	defer tx.Discard()
	obj, err := tx.CreateObject(ctx, key, nil)
	world.ReleaseObjectState(obj)
	if err != nil {
		t.Fatalf("CreateObject %s: %v", key, err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("Commit %s: %v", key, err)
	}
}

// openReadTx opens a read transaction and returns it with its seqno.
func openReadTx(ctx context.Context, t *testing.T, engine world.Engine) (world.Tx, uint64) {
	// Open the transaction and release it with the test.
	t.Helper()
	tx, err := engine.NewTransaction(ctx, false)
	if err != nil {
		t.Fatalf("NewTransaction read: %v", err)
	}
	t.Cleanup(tx.Discard)
	seqno, err := tx.GetSeqno(ctx)
	if err != nil {
		t.Fatalf("GetSeqno read: %v", err)
	}
	return tx, seqno
}

// waitForChanges waits for the tracked state to report a change.
func waitForChanges(ctx context.Context, t *testing.T, tracked *resource_world.TrackedWorldState) {
	// Bound the wait so a missed change fails the test.
	t.Helper()
	waitCtx, waitCancel := context.WithTimeout(ctx, 10*time.Second)
	defer waitCancel()
	if err := tracked.WaitForChanges(waitCtx); err != nil {
		t.Fatalf("WaitForChanges: %v", err)
	}
}

// TestTrackedWorldStateCommitDuringRevisionCheck checks that a commit landing
// while the watch compares revisions still wakes it.
func TestTrackedWorldStateCommitDuringRevisionCheck(t *testing.T) {
	// Start a World testbed.
	ctx := t.Context()
	tb, tbCleanup := setupWorldTestbed(ctx, t)
	defer tbCleanup()

	// Read the object, then commit an unrelated object so the watch compares
	// revisions instead of trusting the read's own seqno.
	const objectKey = "tracked-watch/commit-during-check"
	createObject(ctx, t, tb.Engine, objectKey)
	readTx, seqno := openReadTx(ctx, t, tb.Engine)
	createObject(ctx, t, tb.Engine, "tracked-watch/unrelated")

	// Track the object through a live World that commits during the check.
	watchWs := &commitDuringCheckWorldState{
		WorldState: world.NewEngineWorldState(tb.Engine, false),
		engine:     tb.Engine,
		objectKey:  objectKey,
	}
	trackedWs := resource_world.NewTrackedWorldState(readTx, watchWs, seqno, ctx)
	defer trackedWs.Close()
	objectState, _, err := trackedWs.GetObject(ctx, objectKey)
	world.ReleaseObjectState(objectState)
	if err != nil {
		t.Fatalf("GetObject tracked: %v", err)
	}

	// The commit during the check wakes the watch.
	waitForChanges(ctx, t, trackedWs)
}

// TestTrackedWorldStateIterationSeesNewObject checks that a view that only
// iterates a prefix sees an object created under it.
func TestTrackedWorldStateIterationSeesNewObject(t *testing.T) {
	// Start a World testbed.
	ctx := t.Context()
	tb, tbCleanup := setupWorldTestbed(ctx, t)
	defer tbCleanup()

	// Iterate the prefix through the tracked state.
	readTx, seqno := openReadTx(ctx, t, tb.Engine)
	trackedWs := resource_world.NewTrackedWorldState(readTx, world.NewEngineWorldState(tb.Engine, false), seqno, ctx)
	defer trackedWs.Close()
	iter := trackedWs.IterateObjects(ctx, "tracked-iter/", false)
	for iter.Next() {
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("IterateObjects: %v", err)
	}
	iter.Close()

	// Create an object under the prefix and wait for the change.
	createObject(ctx, t, tb.Engine, "tracked-iter/new")
	waitForChanges(ctx, t, trackedWs)
}
