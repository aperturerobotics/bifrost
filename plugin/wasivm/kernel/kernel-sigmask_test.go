package kernel

import (
	"context"
	"testing"
	"time"
)

// newSigTestProcess returns a registered process with no instance.
func newSigTestProcess() *Process {
	k := &Kernel{procs: newProcessTable()}
	p := newProcess(k, 1, nil)
	k.procs.set(1, p)
	return p
}

// maskSet returns a pointer to the signal set of sigs, for setSigMask.
func maskSet(sigs ...uint32) *uint64 {
	var set uint64
	for _, sig := range sigs {
		set |= sigBit(sig)
	}
	return &set
}

// TestSetSigMask checks each rt_sigprocmask operation, that SIGKILL and
// SIGSTOP stay unblocked and that a nil set only reads the mask.
func TestSetSigMask(t *testing.T) {
	p := newSigTestProcess()
	steps := []struct {
		how     uint32
		set     *uint64
		wantOld uint64
		want    uint64
	}{
		{sigBlock, maskSet(SIGINT, SIGTERM), 0, *maskSet(SIGINT, SIGTERM)},
		{sigUnblock, maskSet(SIGTERM), *maskSet(SIGINT, SIGTERM), sigBit(SIGINT)},
		{sigSetmask, maskSet(SIGQUIT), sigBit(SIGINT), sigBit(SIGQUIT)},
		{sigBlock, nil, sigBit(SIGQUIT), sigBit(SIGQUIT)},
		{sigSetmask, maskSet(SIGKILL, SIGSTOP, SIGINT), sigBit(SIGQUIT), sigBit(SIGINT)},
	}
	for i, step := range steps {
		if old := p.setSigMask(step.how, step.set); old != step.wantOld {
			t.Errorf("step %d: old mask %#x, want %#x", i, old, step.wantOld)
		}
		if p.sigMask != step.want {
			t.Errorf("step %d: mask %#x, want %#x", i, p.sigMask, step.want)
		}
	}
}

// TestBlockedSignalHeldPending checks that a blocked signal waits in the
// pending set and takes its default action once unblocked.
func TestBlockedSignalHeldPending(t *testing.T) {
	// Deliver SIGINT while it is blocked.
	p := newSigTestProcess()
	canceled := false
	p.cancel = func() { canceled = true }
	p.setSigMask(sigSetmask, maskSet(SIGINT))
	p.deliverSignal(SIGINT)
	if canceled || p.sigPending != sigBit(SIGINT) {
		t.Fatalf("blocked SIGINT: canceled %v, pending %#x", canceled, p.sigPending)
	}

	// Unblock it.
	p.setSigMask(sigSetmask, maskSet())
	if !canceled || p.sigPending != 0 {
		t.Fatalf("unblocked SIGINT: canceled %v, pending %#x", canceled, p.sigPending)
	}
}

// TestWaitForSignal checks that rt_sigtimedwait takes a signal delivered
// before or during the wait off the pending set.
func TestWaitForSignal(t *testing.T) {
	// Wait for a signal while another goroutine delivers it.
	p := newSigTestProcess()
	p.setSigMask(sigSetmask, maskSet(SIGCHLD, SIGTERM))
	done := make(chan uint64, 1)
	go func() {
		done <- p.waitForSignal(context.Background(), sigBit(SIGCHLD), -1)
	}()
	p.deliverSignal(SIGCHLD)

	// It returns the signal and leaves nothing pending.
	if got := <-done; got != uint64(SIGCHLD) {
		t.Fatalf("waitForSignal returned %d, want %d", got, SIGCHLD)
	}
	if p.sigPending != 0 {
		t.Fatalf("pending %#x after the wait took SIGCHLD", p.sigPending)
	}
}

// TestWaitForSignalTimeout checks that rt_sigtimedwait fails with EAGAIN at
// its timeout, and at once for a zero timeout.
func TestWaitForSignalTimeout(t *testing.T) {
	p := newSigTestProcess()
	for _, timeout := range []time.Duration{0, time.Millisecond} {
		if got := p.waitForSignal(context.Background(), sigBit(SIGCHLD), timeout); got != negErrno(EAGAIN) {
			t.Errorf("timeout %v: got %#x, want -EAGAIN", timeout, got)
		}
	}
}

// TestSuspendWakesOnUnblockedSignal checks that rt_sigsuspend returns EINTR
// once its mask lets a signal through, then restores the previous mask.
func TestSuspendWakesOnUnblockedSignal(t *testing.T) {
	// Suspend with an empty mask while SIGINT, which has a handler, arrives.
	p := newSigTestProcess()
	p.sigActions[SIGINT] = sigAction{handler: 0x1000}
	p.setSigMask(sigSetmask, maskSet(SIGINT))
	done := make(chan uint64, 1)
	go func() {
		done <- p.suspendWithMask(context.Background(), 0)
	}()
	p.deliverSignal(SIGINT)

	// It returns EINTR with SIGINT caught and blocked again.
	if got := <-done; got != negErrno(EINTR) {
		t.Fatalf("suspend returned %#x, want -EINTR", got)
	}
	if p.sigMask != sigBit(SIGINT) || p.sigCaught != sigBit(SIGINT) {
		t.Fatalf("after suspend: mask %#x, caught %#x", p.sigMask, p.sigCaught)
	}
}

// TestSuspendKeepsBlockedSignalPending checks that a signal rt_sigsuspend's
// mask blocks stays pending, so only the context ends the suspend.
func TestSuspendKeepsBlockedSignalPending(t *testing.T) {
	// Deliver SIGINT before a suspend that keeps it blocked.
	p := newSigTestProcess()
	p.sigActions[SIGINT] = sigAction{handler: 0x1000}
	p.setSigMask(sigSetmask, maskSet(SIGINT))
	p.deliverSignal(SIGINT)

	// The suspend waits until its context ends, leaving SIGINT pending.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := p.suspendWithMask(ctx, sigBit(SIGINT)); got != negErrno(EINTR) {
		t.Fatalf("suspend returned %#x, want -EINTR", got)
	}
	if p.sigPending != sigBit(SIGINT) || p.sigCaught != 0 {
		t.Fatalf("after suspend: pending %#x, caught %#x", p.sigPending, p.sigCaught)
	}
}
