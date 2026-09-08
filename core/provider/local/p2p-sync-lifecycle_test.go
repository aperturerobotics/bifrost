package provider_local

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aperturerobotics/util/routine"
)

// TestRetrySharedObjectSyncRestartsOnlySelectedSO verifies that retry replaces
// one SO routine serially without disturbing its generation or sibling SOs.
func TestRetrySharedObjectSyncRestartsOnlySelectedSO(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	newSyncRoutine := func(started chan<- struct{}, running *atomic.Int32) *routine.RoutineContainer {
		rc := routine.NewRoutineContainer()
		rc.SetRoutine(func(ctx context.Context) error {
			if active := running.Add(1); active != 1 {
				t.Errorf("SO sync overlapped %d executions", active)
			}
			started <- struct{}{}
			<-ctx.Done()
			running.Add(-1)
			return ctx.Err()
		})
		rc.SetContext(ctx, false)
		t.Cleanup(func() {
			exitedCh, _ := rc.SetRoutine(nil)
			if exitedCh != nil {
				<-exitedCh
			}
		})
		return rc
	}

	targetStarted := make(chan struct{}, 2)
	otherStarted := make(chan struct{}, 2)
	var targetRunning atomic.Int32
	var otherRunning atomic.Int32
	targetRoutine := newSyncRoutine(targetStarted, &targetRunning)
	otherRoutine := newSyncRoutine(otherStarted, &otherRunning)
	state := &p2pSyncState{
		ctx:     ctx,
		started: true,
		soSync: map[string]*routine.RoutineContainer{
			"target": targetRoutine,
			"other":  otherRoutine,
		},
	}
	account := &ProviderAccount{}
	account.p2pSyncBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		account.p2pSync = state
	})

	<-targetStarted
	<-otherStarted
	if !account.RetrySharedObjectSync("target") {
		t.Fatal("target SO sync was not restarted")
	}
	<-targetStarted

	select {
	case <-otherStarted:
		t.Fatal("retry restarted an unrelated SO sync")
	default:
	}
	if targetRunning.Load() != 1 || otherRunning.Load() != 1 {
		t.Fatalf("unexpected active routines: target=%d other=%d", targetRunning.Load(), otherRunning.Load())
	}
	if account.p2pSync != state {
		t.Fatal("retry replaced the P2P transport generation")
	}
}

// TestRetrySharedObjectSyncRejectsUnknownSOWithoutRestart verifies that list
// reconciliation can still own initial startup when no routine exists yet.
func TestRetrySharedObjectSyncRejectsUnknownSOWithoutRestart(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	started := make(chan struct{}, 2)
	rc := routine.NewRoutineContainer()
	rc.SetRoutine(func(ctx context.Context) error {
		started <- struct{}{}
		<-ctx.Done()
		return ctx.Err()
	})
	rc.SetContext(ctx, false)
	t.Cleanup(func() {
		exitedCh, _ := rc.SetRoutine(nil)
		if exitedCh != nil {
			<-exitedCh
		}
	})
	state := &p2pSyncState{
		ctx:     ctx,
		started: true,
		soSync:  map[string]*routine.RoutineContainer{"known": rc},
	}
	account := &ProviderAccount{p2pSync: state}

	<-started
	if account.RetrySharedObjectSync("missing") {
		t.Fatal("unknown SO sync was restarted")
	}
	select {
	case <-started:
		t.Fatal("failed retry restarted an existing SO sync")
	default:
	}
}

// TestRetireP2PSyncStateWaitsForSOSyncRoutine verifies that retirement joins
// SO routines before completing generation cleanup.
func TestRetireP2PSyncStateWaitsForSOSyncRoutine(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	routineStarted := make(chan struct{})
	routineCanceled := make(chan struct{})
	allowRoutineExit := make(chan struct{})
	var routineExited atomic.Bool
	rc := routine.NewRoutineContainer()
	rc.SetRoutine(func(ctx context.Context) error {
		close(routineStarted)
		<-ctx.Done()
		close(routineCanceled)
		<-allowRoutineExit
		routineExited.Store(true)
		return ctx.Err()
	})
	rc.SetContext(ctx, false)
	state := &p2pSyncState{
		ctx:           ctx,
		cancel:        cancel,
		started:       true,
		startComplete: true,
		startupExited: true,
		soSync:        map[string]*routine.RoutineContainer{"target": rc},
	}
	account := &ProviderAccount{p2pSync: state}

	<-routineStarted
	retired := make(chan struct{})
	go func() {
		account.retireP2PSyncState(nil)
		close(retired)
	}()
	<-routineCanceled
	select {
	case <-retired:
		t.Fatal("retirement completed before SO sync exited")
	default:
	}
	close(allowRoutineExit)
	<-retired
	if !routineExited.Load() {
		t.Fatal("retirement released resources before SO sync exited")
	}
}

func TestP2PSyncStateFinishStartPublishesStableState(t *testing.T) {
	ctx := t.Context()
	state := &p2pSyncState{ctx: ctx}

	restart := state.finishStart(nil)
	if restart {
		t.Fatal("stable startup requested a restart")
	}

	state.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if !state.startComplete || !state.started || state.startErr != nil {
			t.Fatalf("unexpected completed state: %+v", state)
		}
	})

	restart = state.finishStart(context.Canceled)
	if restart {
		t.Fatal("completed startup requested a restart")
	}
}

func TestP2PSyncStateRestartStaysIncompleteUntilStable(t *testing.T) {
	ctx := t.Context()
	state := &p2pSyncState{ctx: ctx, restartPending: true}

	restart := state.finishStart(nil)
	if !restart {
		t.Fatal("pending startup did not request a restart")
	}
	state.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if state.startComplete {
			t.Fatal("restart decision published startup completion")
		}
		if state.restartPending {
			t.Fatal("restart decision remained pending")
		}
	})

	restart = state.finishStart(nil)
	if restart {
		t.Fatal("stable startup requested an extra restart")
	}
}

func TestRetireP2PSyncStateWaitsForOwnedWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state := &p2pSyncState{
		ctx:           ctx,
		cancel:        cancel,
		startupExited: false,
		workers:       1,
		relFns:        []func(){func() {}},
	}
	account := &ProviderAccount{}
	account.p2pSyncBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		account.p2pSync = state
	})

	retired := make(chan struct{})
	go func() {
		account.retireP2PSyncState(nil)
		close(retired)
	}()

	for {
		var (
			stopping bool
			waitCh   <-chan struct{}
		)
		state.bcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			stopping = state.stopping
			if !stopping {
				waitCh = getWaitCh()
			}
		})
		if stopping {
			break
		}
		<-waitCh
	}

	select {
	case <-retired:
		t.Fatal("retirement completed while owned work remained")
	default:
	}

	state.bcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		state.startupExited = true
		state.workers = 0
		bcast()
	})
	<-retired

	state.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if !state.cleanupDone {
			t.Fatal("retirement did not complete cleanup")
		}
	})
}

func TestRetireP2PSyncStateReleasesLowerChain(t *testing.T) {
	makeState := func(lower *p2pSyncState) *p2pSyncState {
		ctx, cancel := context.WithCancel(context.Background())
		return &p2pSyncState{
			ctx:             ctx,
			cancel:          cancel,
			owners:          1,
			startComplete:   true,
			started:         true,
			startupExited:   true,
			lowerSource:     lower,
			lowerSourceHeld: lower != nil,
		}
	}

	first := makeState(nil)
	second := makeState(first)
	third := makeState(second)
	account := &ProviderAccount{}
	account.p2pSyncBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		account.p2pSync = third
	})

	account.retireP2PSyncState(nil)

	for name, state := range map[string]*p2pSyncState{
		"first":  first,
		"second": second,
		"third":  third,
	} {
		state.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
			if !state.stopping || !state.cleanupDone {
				t.Fatalf("%s state was not fully retired", name)
			}
			if state.lowerSource != nil || state.lowerSourceHeld {
				t.Fatalf("%s state retained its lower source", name)
			}
		})
		if state.ctx.Err() == nil {
			t.Fatalf("%s state context was not canceled", name)
		}
	}
}

func TestFailedP2PSyncStartDoesNotRestoreStoppedPrevious(t *testing.T) {
	makeState := func(stopping bool) *p2pSyncState {
		ctx, cancel := context.WithCancel(context.Background())
		return &p2pSyncState{
			ctx:           ctx,
			cancel:        cancel,
			startComplete: true,
			started:       !stopping,
			startupExited: true,
			stopping:      stopping,
		}
	}

	previous := makeState(true)
	failed := makeState(true)
	account := &ProviderAccount{}
	account.p2pSyncBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		account.p2pSync = failed
	})

	account.restoreP2PSyncAfterFailedStart(failed, previous, false)

	account.p2pSyncBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if account.p2pSync != nil {
			t.Fatal("failed replacement restored a stopped previous state")
		}
	})
	previous.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if !previous.cleanupDone {
			t.Fatal("stopped previous state was not retired")
		}
	})
}

func TestFailedP2PSyncStartDoesNotRestoreErroredPrevious(t *testing.T) {
	previousCtx, previousCancel := context.WithCancel(context.Background())
	previousCancel()
	previous := &p2pSyncState{
		ctx:           previousCtx,
		cancel:        previousCancel,
		startComplete: true,
		started:       true,
		startupExited: true,
	}
	failedCtx, failedCancel := context.WithCancel(context.Background())
	defer failedCancel()
	failed := &p2pSyncState{
		ctx:           failedCtx,
		cancel:        failedCancel,
		startComplete: true,
		startupExited: true,
		stopping:      true,
	}
	account := &ProviderAccount{}
	account.p2pSyncBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		account.p2pSync = failed
	})

	account.restoreP2PSyncAfterFailedStart(failed, previous, true)

	account.p2pSyncBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if account.p2pSync != nil {
			t.Fatal("failed replacement restored a previous state with an errored context")
		}
	})
	previous.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if !previous.cleanupDone {
			t.Fatal("previous state with an errored context was not retired")
		}
	})
}

func TestFailedP2PSyncStartDoesNotRestoreRetiredRetainedPredecessor(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	makeState := func(owners int, complete bool) *p2pSyncState {
		stateCtx, stateCancel := context.WithCancel(ctx)
		return &p2pSyncState{
			ctx:           stateCtx,
			cancel:        stateCancel,
			owners:        owners,
			startComplete: complete,
			started:       complete,
			startupExited: true,
		}
	}

	oldest := makeState(1, true)
	middle := makeState(1, false)
	newest := makeState(1, false)
	account := &ProviderAccount{}
	account.p2pSyncBcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		account.p2pSync = oldest
		if !account.retainP2PSyncLowerSourceLocked(oldest) {
			t.Fatal("middle start did not retain oldest state")
		}
		middle.lowerSource = oldest
		middle.lowerSourceHeld = true
		account.p2pSync = middle
		if !account.retainP2PSyncLowerSourceLocked(middle) {
			t.Fatal("newest start did not retain middle state")
		}
		newest.lowerSource = middle
		newest.lowerSourceHeld = true
		account.p2pSync = newest
		bcast()
	})

	// The oldest caller gives up while the middle and newest starts still
	// retain the replacement chain.
	account.releaseP2PSyncState(oldest)

	middleRetired := make(chan struct{})
	go func() {
		account.retireP2PSyncState(middle)
		close(middleRetired)
	}()
	select {
	case <-middleRetired:
	case <-ctx.Done():
		t.Fatal("middle state retirement did not complete")
	}
	select {
	case <-oldest.ctx.Done():
	case <-ctx.Done():
		t.Fatal("middle retirement did not stop oldest state")
	}
	oldest.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if !oldest.stopping {
			t.Fatal("retired middle state left oldest state active")
		}
	})

	if restart := newest.finishStart(errors.New("controlled startup failure")); restart {
		t.Fatal("failed newest start requested a restart")
	}
	newest.markStartupExited()
	newestRetired := make(chan struct{})
	go func() {
		account.restoreP2PSyncAfterFailedStart(newest, middle, true)
		account.retireP2PSyncState(newest)
		close(newestRetired)
	}()
	select {
	case <-newestRetired:
	case <-ctx.Done():
		t.Fatal("newest state retirement did not complete")
	}

	account.p2pSyncBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if account.p2pSync != nil {
			t.Fatal("failed newest start restored a retired predecessor")
		}
	})
}
