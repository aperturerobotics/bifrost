package world

import (
	"context"

	"github.com/pkg/errors"
)

// ChangeWatcher is an Engine that reports what changed between seqnos.
type ChangeWatcher interface {
	// WatchChanges calls cb with each later seqno and the changes since the
	// previous one, starting after afterSeqno. Calls coalesce while cb runs.
	// Returns when ctx is canceled or cb returns an error.
	WatchChanges(ctx context.Context, afterSeqno uint64, cb func(seqno uint64, changes *ChangeSet) error) error
}

// WatchChanges calls cb with each seqno of w after afterSeqno and the changes
// since the previous call. w is an Engine or a WorldState; one that is not a
// ChangeWatcher reports every advance as an unknown ChangeSet.
// Returns when ctx is canceled or cb returns an error.
func WatchChanges(
	ctx context.Context,
	w WorldWaitSeqno,
	afterSeqno uint64,
	cb func(seqno uint64, changes *ChangeSet) error,
) error {
	// Prefer the World's own change feed.
	if watcher, ok := w.(ChangeWatcher); ok {
		return watcher.WatchChanges(ctx, afterSeqno, cb)
	}

	// Report each seqno advance as unknown.
	for {
		seqno, err := w.WaitSeqno(ctx, afterSeqno+1)
		if err != nil {
			return err
		}
		if err := cb(seqno, NewUnknownChangeSet()); err != nil {
			return err
		}
		afterSeqno = seqno
	}
}

// WatchView runs view immediately and again when a change touches its reads or
// its earliest RerunAt deadline arrives. source must be an Engine or WorldState;
// an Engine supplies a live, read-only WorldState. Every evaluation has a fresh
// read set carried by its context. Transactions opened through a
// ReadTrackingEngine with that context join the same set, including transactions
// opened by clients captured by view. All transactions must read source's World,
// and all evaluation work must finish before view returns. Unknown changes can
// rerun an unchanged projection; use SendViews to suppress equal results.
// Returns when ctx ends or view or the watch fails.
func WatchView(ctx context.Context, source WorldWaitSeqno, view func(ctx context.Context, ws WorldState) error) error {
	// Preserve an existing state's access policy, or build live engine reads.
	ws, ok := source.(WorldState)
	if !ok {
		engine, ok := source.(Engine)
		if !ok {
			return errors.New("view source must be an Engine or WorldState")
		}
		ws = NewEngineWorldState(engine, false)
	}

	// Record each evaluation's starting seqno before reading its dependencies.
	for {
		// Do not start another evaluation after cancellation.
		if err := ctx.Err(); err != nil {
			return err
		}

		// Run the view against the head, recording its reads.
		seqno, err := source.GetSeqno(ctx)
		if err != nil {
			return err
		}
		reads := &viewReads{reads: NewReadSet()}
		if err := view(context.WithValue(ctx, viewReadsKey{}, reads), reads.reads.Track(ws)); err != nil {
			return err
		}

		// Wait for a relevant change, bounded by the evaluation's deadline.
		if err := reads.wait(ctx, source, seqno); err != nil {
			return err
		}
	}
}
