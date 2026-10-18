//go:build !js && !wasip1

package world_block

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/testbed"
	"github.com/s4wave/spacewave/db/world"
	world_mock "github.com/s4wave/spacewave/db/world/mock"
	"github.com/sirupsen/logrus"
)

// TestEngineSyncThroughDeferredDurability checks the per-seqno fence of the
// single-writer deferred mode: it fences only a head the durable head does not
// cover, and a covered seqno returns without a fence.
func TestEngineSyncThroughDeferredDurability(t *testing.T) {
	// Open a deferred-durability engine that counts durable head writes.
	ctx := t.Context()
	le := logrus.NewEntry(logrus.New())
	tb, err := testbed.NewTestbed(ctx, le)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tb.Release)

	// Build an empty world cursor to back the engine.
	base, err := tb.BuildEmptyCursor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(base.Release)

	// Count the durable head writes the engine makes.
	var heads int
	var durableHead *bucket.ObjectRef
	eng, err := NewEngine(ctx, le, base, world_mock.LookupMockOp,
		func(_ context.Context, _, next *bucket.ObjectRef) error {
			heads++
			durableHead = next.Clone()
			return nil
		}, false, WithDeferredDurability())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := eng.Close(); err != nil {
			t.Error(err)
		}
	})
	commit := func(key string) uint64 {
		t.Helper()

		// Create one object in a write transaction.
		w, err := eng.NewBlockEngineTransaction(ctx, true)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Discard()
		obj, err := w.CreateObject(ctx, key, &bucket.ObjectRef{BucketId: "test"})
		world.ReleaseObjectState(obj)
		if err != nil {
			t.Fatal(err)
		}

		// Commit it and read the seqno it published.
		if err := w.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		seqno, err := eng.GetSeqno(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return seqno
	}

	// A deferred commit leaves the durable head behind; SyncThrough catches it up.
	first := commit("first")
	if heads != 0 {
		t.Fatalf("deferred commit wrote %d durable heads", heads)
	}
	durable, err := eng.SyncThrough(ctx, first)
	if err != nil || durable != first {
		t.Fatalf("SyncThrough(%d) = %d, %v; want %d", first, durable, err, first)
	}
	if heads != 1 || !durableHead.EqualsRef(eng.GetRootRef()) {
		t.Fatalf("durable head writes = %d; want one at the published head", heads)
	}

	// A covered seqno returns without a fence, even with a later unsynced commit.
	second := commit("second")
	durable, err = eng.SyncThrough(ctx, first)
	if err != nil || durable != first || heads != 1 {
		t.Fatalf("covered SyncThrough = %d, %v, %d head writes; want %d, nil, 1", durable, err, heads, first)
	}
	durable, err = eng.SyncThrough(ctx, 0)
	if err != nil || durable != first || heads != 1 {
		t.Fatalf("SyncThrough(0) = %d, %v, %d head writes; want %d, nil, 1", durable, err, heads, first)
	}

	// An uncovered seqno fences the later commit.
	durable, err = eng.SyncThrough(ctx, second)
	if err != nil || durable != second || heads != 2 {
		t.Fatalf("SyncThrough(%d) = %d, %v, %d head writes; want %d, nil, 2", second, durable, err, heads, second)
	}

	// A seqno no commit has reached waits for that commit.
	waitCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if _, err := eng.SyncThrough(waitCtx, second+1); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("SyncThrough past the head = %v; want deadline exceeded", err)
	}
}

// TestEngineSyncThroughSkipsLaterAdmissions checks that a fence for an earlier
// commit does not join a later writer's admitted but undurable publication,
// which Sync does.
func TestEngineSyncThroughSkipsLaterAdmissions(t *testing.T) {
	// Make the first writer's commit durable.
	f := newSessionFixture(t)
	first := sessionWriter(t, f)
	sessionObject(t, first, "one")
	if err := sessionWait(t, sessionSubmit(t, first)); err != nil {
		t.Fatal(err)
	}
	firstSeqno, err := f.engine.GetSeqno(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	// Admit a second writer's publication and hold it undurable.
	g := f.publisher.arm(t, nil)
	second := sessionWriter(t, f)
	sessionObject(t, second, "two")
	r2 := sessionSubmit(t, second)
	sessionEntered(t, g)

	// Sync joins the second writer.
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	_, err = f.engine.Sync(ctx)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Sync did not join the undurable publication: %v", err)
	}

	// SyncThrough for the first writer does not.
	ctx, cancel = context.WithTimeout(t.Context(), 2*time.Second)
	durable, err := f.engine.SyncThrough(ctx, firstSeqno)
	cancel()
	if err != nil || durable != firstSeqno {
		t.Fatalf("SyncThrough(%d) = %d, %v; want %d, nil", firstSeqno, durable, err, firstSeqno)
	}

	// The durable head covers the first commit.
	head, err := f.load(t.Context())
	if err != nil || !head.EqualsRef(f.engine.GetRootRef()) {
		t.Fatalf("durable head %v, %v does not cover the first commit", head, err)
	}

	// A seqno the held publication will take waits for it.
	ctx, cancel = context.WithTimeout(t.Context(), 20*time.Millisecond)
	_, err = f.engine.SyncThrough(ctx, firstSeqno+1)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("SyncThrough past the held publication: %v", err)
	}

	// Releasing the publication lets its own seqno fence.
	g.open()
	if err := sessionWait(t, r2); err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	if durable, err := f.engine.SyncThrough(ctx, firstSeqno+1); err != nil || durable != firstSeqno+1 {
		t.Fatalf("SyncThrough(%d) = %d, %v; want %d, nil", firstSeqno+1, durable, err, firstSeqno+1)
	}
}
