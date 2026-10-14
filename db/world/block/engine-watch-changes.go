package world_block

import (
	"context"
	"slices"

	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/world"
)

// changeWatchHeads bounds the published heads the Engine retains so a change
// watcher can compare a later head against the one it last saw.
const changeWatchHeads = 128

// changeWatchDiffLimit bounds the changed tree keys one head comparison lists
// per tree before it reports an unknown ChangeSet.
const changeWatchDiffLimit = 16384

// watchedHead is a published head retained for change watchers.
type watchedHead struct {
	// seqno is the World seqno of the head.
	seqno uint64
	// root is the held root cursor.
	root *bucket_lookup.Cursor
	// release drops the root and its pin.
	release func()
}

// WatchChanges calls cb with each later seqno and the changes since the
// previous one. The changes come from comparing the World trees of the head the
// watcher last saw against the current head, so they need no changelog.
//
// The Engine retains the last changeWatchHeads heads it publishes after the
// first watcher starts. The ChangeSet is unknown when the head at the previous
// seqno is no longer retained or was published before retention started, when
// two retained heads share that seqno, or when the change is too large to
// list. Calls coalesce while cb runs.
func (e *Engine) WatchChanges(
	ctx context.Context,
	afterSeqno uint64,
	cb func(seqno uint64, changes *world.ChangeSet) error,
) error {
	if err := e.startWatchedHeads(ctx); err != nil {
		return err
	}
	for {
		// Wait for the head to pass the base seqno.
		if _, err := e.WaitSeqno(ctx, afterSeqno+1); err != nil {
			return err
		}

		// Hold the current head and the retained head at the base seqno.
		head, base, err := e.holdWatchedHeads(ctx, afterSeqno)
		if err != nil {
			return err
		}
		if head.seqno <= afterSeqno {
			head.release()
			continue
		}

		// Compare the two heads outside the publication lock.
		changes := world.NewUnknownChangeSet()
		if base != nil {
			changes, err = e.compareHeads(ctx, base.root, head.root)
			base.release()
		}
		head.release()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			e.le.WithError(err).Warn("world change comparison failed; reporting unknown changes")
		}

		// Report the changes and advance the base.
		if err := cb(head.seqno, changes); err != nil {
			return err
		}
		afterSeqno = head.seqno
	}
}

// startWatchedHeads starts retaining published heads for change watchers.
// The current head was published before retention, and a watcher at its seqno
// may have seen another root with that seqno, so its seqno becomes the floor.
func (e *Engine) startWatchedHeads(ctx context.Context) error {
	// Leave retention running when it already started.
	locked := e.bcast.Lock()
	defer locked.Unlock()
	if e.closed {
		return ErrEngineClosed
	}
	if e.watchedHeads != nil {
		return nil
	}

	// Start with an empty history above the current head.
	seqno, err := e.headSeqnoLocked(ctx)
	if err != nil {
		return err
	}
	e.watchedHeads = make([]*watchedHead, 0, changeWatchHeads)
	e.watchedFloor = seqno
	return nil
}

// holdWatchedHeads holds the current head and the retained head at
// afterSeqno. base is nil when no single retained head has that seqno.
func (e *Engine) holdWatchedHeads(ctx context.Context, afterSeqno uint64) (head, base *watchedHead, err error) {
	// Hold the current head under the publication lock.
	locked := e.bcast.Lock()
	defer locked.Unlock()
	if e.closed {
		return nil, nil, ErrEngineClosed
	}
	seqno, err := e.headSeqnoLocked(ctx)
	if err != nil {
		return nil, nil, err
	}
	head, err = e.holdWatchedHead(ctx, e.head.root, seqno)
	if err != nil {
		return nil, nil, err
	}

	// Find the one retained root at the base seqno.
	if afterSeqno <= e.watchedFloor {
		return head, nil, nil
	}
	var found *watchedHead
	for _, h := range e.watchedHeads {
		if h.seqno != afterSeqno {
			continue
		}
		if found != nil && !found.root.GetRef().EqualsRef(h.root.GetRef()) {
			return head, nil, nil
		}
		found = h
	}
	if found == nil {
		return head, nil, nil
	}
	base, err = e.holdWatchedHead(ctx, found.root, found.seqno)
	if err != nil {
		head.release()
		return nil, nil, err
	}
	return head, base, nil
}

// recordWatchedHeadLocked retains a newly published head for change watchers
// once retention has started. It returns the evicted oldest head for release
// after the lock. A head that cannot be retained raises the floor instead, so
// no watcher compares across it. The caller must hold bcast.
func (e *Engine) recordWatchedHeadLocked(ctx context.Context, root *bucket_lookup.Cursor, seqno uint64) *watchedHead {
	// Skip retention until the first watcher starts it.
	if e.watchedHeads == nil {
		return nil
	}

	// Retain the head, or fence comparisons across it.
	h, err := e.holdWatchedHead(ctx, root, seqno)
	if err != nil {
		e.le.WithError(err).Warn("world change watch cannot retain head; reporting unknown changes")
		e.watchedFloor = max(e.watchedFloor, seqno)
		return nil
	}

	// Evict the oldest head at capacity.
	var evicted *watchedHead
	if len(e.watchedHeads) == changeWatchHeads {
		evicted = e.watchedHeads[0]
		e.watchedFloor = max(e.watchedFloor, evicted.seqno)
		e.watchedHeads = slices.Delete(e.watchedHeads, 0, 1)
	}
	e.watchedHeads = append(e.watchedHeads, h)
	return evicted
}

// headSeqnoLocked returns the seqno of the published head. The caller must
// hold bcast.
func (e *Engine) headSeqnoLocked(ctx context.Context) (uint64, error) {
	if err := e.initializeHeadReadTx(ctx); err != nil {
		return 0, err
	}
	return e.head.readTx.state.GetSeqno(ctx)
}

// holdWatchedHead holds root with its seqno.
func (e *Engine) holdWatchedHead(ctx context.Context, root *bucket_lookup.Cursor, seqno uint64) (*watchedHead, error) {
	held, release, err := e.holdRoot(ctx, root)
	if err != nil {
		return nil, err
	}
	return &watchedHead{seqno: seqno, root: held, release: release}, nil
}

// compareHeads lists the changes between two held World roots.
func (e *Engine) compareHeads(ctx context.Context, before, after *bucket_lookup.Cursor) (*world.ChangeSet, error) {
	_, beforeBcs := before.BuildTransactionWithStore(nil, e.writeBlockStore)
	_, afterBcs := after.BuildTransactionWithStore(nil, e.writeBlockStore)
	return WorldChanges(ctx, beforeBcs, afterBcs, changeWatchDiffLimit)
}

// _ is a type assertion
var _ world.ChangeWatcher = (*Engine)(nil)
