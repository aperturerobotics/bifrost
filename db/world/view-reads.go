package world

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// viewReads collects the reads and next deadline of one view evaluation.
// All work using its context must finish before the evaluation returns.
type viewReads struct {
	// reads unions the reads of every transaction opened during the evaluation.
	reads *ReadSet
	// mtx guards wake.
	mtx sync.Mutex
	// wake is the earliest requested rerun time, or zero.
	wake time.Time
}

// viewReadsKey identifies the evaluation carried by a context.
type viewReadsKey struct{}

// getViewReads returns the evaluation carried by ctx, or nil outside a view.
func getViewReads(ctx context.Context) *viewReads {
	reads, _ := ctx.Value(viewReadsKey{}).(*viewReads)
	return reads
}

// RerunAt asks the running WatchView or SendViews evaluation to run again at
// deadline even without a World change. The earliest nonzero deadline wins.
// It does nothing outside an evaluation. Call it before the evaluation returns.
func RerunAt(ctx context.Context, deadline time.Time) {
	// Ignore absent evaluations and absent deadlines.
	reads := getViewReads(ctx)
	if reads == nil || deadline.IsZero() {
		return
	}

	// Keep the earliest deadline across concurrent reads.
	reads.mtx.Lock()
	if reads.wake.IsZero() || deadline.Before(reads.wake) {
		reads.wake = deadline
	}
	reads.mtx.Unlock()
}

// getDeadline returns the earliest rerun time requested by the evaluation.
func (r *viewReads) getDeadline() time.Time {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.wake
}

// wait ends when the evaluation's reads change or its deadline arrives.
func (r *viewReads) wait(ctx context.Context, source WorldWaitSeqno, seqno uint64) error {
	// A deadline belongs only to the wait, not the evaluation's reads.
	wait := ctx
	if deadline := r.getDeadline(); !deadline.IsZero() {
		var cancel context.CancelFunc
		wait, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	// Distinguish our rerun deadline from caller cancellation and watch errors.
	_, err := r.reads.WaitChange(wait, source, seqno)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.DeadlineExceeded) && wait.Err() == context.DeadlineExceeded {
		return nil
	}
	return err
}
