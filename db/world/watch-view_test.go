package world_test

import (
	"context"
	"errors"
	"testing"
	"time"

	block_mock "github.com/s4wave/spacewave/db/block/mock"
	"github.com/s4wave/spacewave/db/world"
	world_testbed "github.com/s4wave/spacewave/db/world/testbed"
)

// TestWatchViewSecondTransaction checks a lost wakeup when only a transaction
// opened by a captured client reads the changed key, before the watch starts.
func TestWatchViewSecondTransaction(t *testing.T) {
	// Give the view and its captured transaction client the same tracking engine.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	tb := world_testbed.MustDefault(t, ctx)
	engine := world.NewReadTrackingEngine(tb.Engine)

	// Coordinate the write with the evaluation before its change wait begins.
	started := make(chan struct{})
	committed := make(chan struct{})
	done := make(chan error, 1)
	stop := errors.New("view observed second transaction change")
	runs := 0

	// Read the first key through the supplied state and the second through a client.
	go func() {
		done <- world.WatchView(ctx, engine, func(ctx context.Context, ws world.WorldState) error {
			// Keep the first dependency unchanged throughout the test.
			if _, err := ws.HasObject(ctx, "view/first"); err != nil {
				return err
			}

			// Open and discard a distinct snapshot through the captured client.
			found, err := readViewObject(ctx, engine, "view/second")
			if err != nil {
				return err
			}
			runs++
			if runs == 2 {
				if !found {
					return errors.New("rerun did not read the second transaction's change")
				}
				return stop
			}

			// Let the write land after both reads but before WaitChange starts.
			close(started)
			select {
			case <-committed:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()

	// Change only the second transaction's key while the first evaluation runs.
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("view did not finish its two initial reads")
	}
	createViewObject(t, ctx, engine, "view/second")
	close(committed)

	// Exactly one rerun must observe the change without another commit.
	select {
	case err := <-done:
		if !errors.Is(err, stop) {
			t.Fatalf("WatchView = %v, want %v", err, stop)
		}
		if runs != 2 {
			t.Fatalf("view ran %d times, want 2", runs)
		}
	case <-ctx.Done():
		t.Fatal("view missed the second transaction's change")
	}
}

// TestWatchViewRerunAt checks the earliest deadline without any World commits.
func TestWatchViewRerunAt(t *testing.T) {
	// Bound a quiet World view and its requested deadline.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	tb := world_testbed.MustDefault(t, ctx)
	deadline := time.Now().Add(20 * time.Millisecond)
	stop := errors.New("deadline reached")
	runs := 0

	// Request deadlines in both orders; neither a later nor zero time replaces it.
	err := world.WatchView(ctx, tb.WorldState, func(ctx context.Context, _ world.WorldState) error {
		// Finish the second evaluation only after the requested time.
		runs++
		if runs == 2 {
			if time.Now().Before(deadline) {
				return errors.New("view reran before its deadline")
			}
			return stop
		}

		// Register the earliest deadline among later and absent requests.
		world.RerunAt(ctx, deadline.Add(time.Hour))
		world.RerunAt(ctx, deadline)
		world.RerunAt(ctx, deadline.Add(2*time.Hour))
		world.RerunAt(ctx, time.Time{})
		return nil
	})

	// A timer, not a commit or caller cancellation, must have rerun the view.
	if !errors.Is(err, stop) {
		t.Fatalf("WatchView = %v, want %v", err, stop)
	}
	if runs != 2 {
		t.Fatalf("view ran %d times, want 2", runs)
	}
}

// TestSendViewsSuppressesEqual checks first-nil delivery and generated equality.
func TestSendViewsSuppressesEqual(t *testing.T) {
	// Prepare immutable projections, including equal but distinct messages.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	tb := world_testbed.MustDefault(t, ctx)
	changed := block_mock.NewExample("changed")
	projections := []*block_mock.Example{nil, nil, changed, changed.CloneVT()}
	stop := errors.New("all projections evaluated")
	var sent []*block_mock.Example
	runs := 0

	// Force fresh evaluations without changing the World.
	err := world.SendViews(ctx, tb.Engine, func(ctx context.Context, _ world.WorldState) (*block_mock.Example, error) {
		// Stop after evaluating each repeated projection.
		if runs == len(projections) {
			return nil, stop
		}

		// Schedule an immediate rerun while returning the next immutable value.
		next := projections[runs]
		runs++
		world.RerunAt(ctx, time.Now())
		return next, nil
	}, func(next *block_mock.Example) error {
		sent = append(sent, next)
		return nil
	})

	// Send nil once and the changed content once, not their equal repetitions.
	if !errors.Is(err, stop) {
		t.Fatalf("SendViews = %v, want %v", err, stop)
	}
	if len(sent) != 2 {
		t.Fatalf("sent %d projections, want 2", len(sent))
	}
	if sent[0] != nil {
		t.Fatalf("first projection = %v, want nil", sent[0])
	}
	if !sent[1].EqualVT(changed) {
		t.Fatalf("second projection = %v, want %v", sent[1], changed)
	}
}

// TestSendViewsErrors checks that reader and sender failures end the watch.
func TestSendViewsErrors(t *testing.T) {
	// Share a quiet World across the two immediate failure paths.
	tb := world_testbed.MustDefault(t, t.Context())
	stop := errors.New("projection failed")

	// Preserve a reader's error without sending its partial result.
	err := world.SendViews(t.Context(), tb.Engine, func(context.Context, world.WorldState) (*block_mock.Example, error) {
		return nil, stop
	}, func(*block_mock.Example) error {
		t.Error("sent a failed projection")
		return nil
	})
	if !errors.Is(err, stop) {
		t.Fatalf("reader failure = %v, want %v", err, stop)
	}

	// Preserve a sender's error without evaluating again.
	err = world.SendViews(t.Context(), tb.Engine, func(context.Context, world.WorldState) (*block_mock.Example, error) {
		return nil, nil
	}, func(*block_mock.Example) error {
		return stop
	})
	if !errors.Is(err, stop) {
		t.Fatalf("sender failure = %v, want %v", err, stop)
	}
}

// TestWatchViewCancellation checks that caller expiry is not a rerun deadline.
func TestWatchViewCancellation(t *testing.T) {
	// Cancel the caller after the first evaluation asks for a future rerun.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	tb := world_testbed.MustDefault(t, t.Context())
	runs := 0
	err := world.WatchView(ctx, tb.Engine, func(ctx context.Context, _ world.WorldState) error {
		runs++
		world.RerunAt(ctx, time.Now().Add(time.Hour))
		cancel()
		return nil
	})

	// End at cancellation without another evaluation.
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WatchView = %v, want context.Canceled", err)
	}
	if runs != 1 {
		t.Fatalf("view ran %d times, want 1", runs)
	}
}

// readViewObject reads one key in its own production Engine transaction.
func readViewObject(ctx context.Context, engine world.Engine, key string) (bool, error) {
	// Discard this snapshot before any client opens the next one.
	tx, err := engine.NewTransaction(ctx, false)
	if err != nil {
		return false, err
	}
	defer tx.Discard()

	// Return existence without retaining an object handle.
	return tx.HasObject(ctx, key)
}

// createViewObject publishes one key through the production Engine interface.
func createViewObject(t *testing.T, ctx context.Context, engine world.Engine, key string) {
	// Keep the write transaction and its object handle scoped to this helper.
	t.Helper()
	tx, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()

	// Publish the new object before notifying the view.
	obj, err := tx.CreateObject(ctx, key, nil)
	world.ReleaseObjectState(obj)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
