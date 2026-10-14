package testbed

import (
	"context"
	"testing"
	"time"

	"github.com/s4wave/spacewave/db/world"
	world_types "github.com/s4wave/spacewave/db/world/types"
)

// TestReadSetWaitChange checks that WaitChange skips a commit the reads did not
// touch and ends at the commit that touched them.
func TestReadSetWaitChange(t *testing.T) {
	// Read one object of a World.
	ctx := t.Context()
	tb := MustDefault(t, ctx)
	ws := world.NewEngineWorldState(tb.Engine, true)
	reads := world.NewReadSet()
	seqno := readObject(t, ctx, reads.Track(ws), "wait/read")

	// Commit an unrelated object, then the read one.
	createObject(t, ctx, ws, "wait/unrelated")
	createObject(t, ctx, ws, "wait/read")
	want, err := ws.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Expect the wait to end at the second commit.
	if got := waitChange(t, ctx, reads, ws, seqno); got != want {
		t.Fatalf("WaitChange = %d, want %d", got, want)
	}
}

// TestReadSetWaitChangeLateRead checks that a change landing before the read of
// what it changed still ends a wait already running.
func TestReadSetWaitChangeLateRead(t *testing.T) {
	// Start waiting before the view reads anything.
	ctx := t.Context()
	tb := MustDefault(t, ctx)
	ws := world.NewEngineWorldState(tb.Engine, true)
	reads := world.NewReadSet()
	seqno, err := ws.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := reads.WaitChange(ctx, ws, seqno)
		done <- err
	}()

	// Change an object, then read it.
	createObject(t, ctx, ws, "late/read")
	readObject(t, ctx, reads.Track(ws), "late/read")

	// Expect the wait to end.
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("WaitChange missed a change made before its read")
	}
}

// TestReadSetWaitChangeIteration checks that a view that only iterates a
// prefix sees an object created under it.
func TestReadSetWaitChangeIteration(t *testing.T) {
	// Open a World and take its seqno.
	ctx := t.Context()
	tb := MustDefault(t, ctx)
	ws := world.NewEngineWorldState(tb.Engine, true)
	reads := world.NewReadSet()
	seqno, err := ws.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Iterate the prefix through the tracked state.
	iter := reads.Track(ws).IterateObjects(ctx, "iter/", false)
	for iter.Next() {
	}
	if err := iter.Err(); err != nil {
		t.Fatal(err)
	}
	iter.Close()

	// Create an object under the prefix and expect the wait to end.
	createObject(t, ctx, ws, "iter/new")
	waitChange(t, ctx, reads, ws, seqno)
}

// TestReadSetWaitChangeTypeListing checks that a view listing a type skips a
// commit typing an object as another type and sees an object gaining its type.
func TestReadSetWaitChangeTypeListing(t *testing.T) {
	// List the objects of a type through the tracked state.
	ctx := t.Context()
	tb := MustDefault(t, ctx)
	ws := world.NewEngineWorldState(tb.Engine, true)
	reads := world.NewReadSet()
	seqno, err := ws.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world_types.ListObjectsWithType(ctx, reads.Track(ws), "listed"); err != nil {
		t.Fatal(err)
	}

	// Type one object as another type, then give a second object the type.
	createObject(t, ctx, ws, "typed/unrelated")
	if err := world_types.SetObjectType(ctx, ws, "typed/unrelated", "other"); err != nil {
		t.Fatal(err)
	}
	createObject(t, ctx, ws, "typed/listed")
	if err := world_types.SetObjectType(ctx, ws, "typed/listed", "listed"); err != nil {
		t.Fatal(err)
	}
	want, err := ws.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Expect the wait to end at the commit that set the type.
	if got := waitChange(t, ctx, reads, ws, seqno); got != want {
		t.Fatalf("WaitChange = %d, want %d", got, want)
	}
}

// readObject reads key through ws and returns the seqno before the read.
func readObject(t *testing.T, ctx context.Context, ws world.WorldState, key string) uint64 {
	// Take the seqno, then read the object.
	t.Helper()
	seqno, err := ws.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err)
	}
	obj, _, err := ws.GetObject(ctx, key)
	world.ReleaseObjectState(obj)
	if err != nil {
		t.Fatal(err)
	}
	return seqno
}

// createObject creates an object at key.
func createObject(t *testing.T, ctx context.Context, ws world.WorldState, key string) {
	t.Helper()
	obj, err := ws.CreateObject(ctx, key, nil)
	world.ReleaseObjectState(obj)
	if err != nil {
		t.Fatal(err)
	}
}

// waitChange waits for a change touching reads after seqno, bounded so a
// missed change fails the test.
func waitChange(t *testing.T, ctx context.Context, reads *world.ReadSet, ws world.WorldState, seqno uint64) uint64 {
	// Bound the wait so a missed change fails the test.
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	got, err := reads.WaitChange(ctx, ws, seqno)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
