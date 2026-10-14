package world

import "context"

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

// WatchView runs view against ws, recording what it reads, and runs it again
// each time a World change touches those reads. A view that sends a result
// should compare it with the last one sent, since an unknown or coarse change
// reruns the view without changing what it reads.
// Returns when ctx is canceled or view returns an error.
func WatchView(ctx context.Context, ws WorldState, view func(ctx context.Context, ws WorldState) error) error {
	for {
		// Run the view against the head, recording its reads.
		seqno, err := ws.GetSeqno(ctx)
		if err != nil {
			return err
		}
		reads := NewReadSet()
		if err := view(ctx, reads.Track(ws)); err != nil {
			return err
		}

		// Wait for a change to what the view read.
		if _, err := reads.WaitChange(ctx, ws, seqno); err != nil {
			return err
		}
	}
}
