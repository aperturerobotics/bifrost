package world_block

import (
	"context"
	"slices"

	"github.com/s4wave/spacewave/db/block"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/world"
)

// changeWatchHeads bounds the published heads the Engine records so a change
// watcher can compare a later head against the one it last saw.
const changeWatchHeads = 128

// changeWatchDiffLimit bounds the changed tree keys one head comparison lists
// per tree before it reports an unknown ChangeSet.
const changeWatchDiffLimit = 16384

// watchedHead is a published head recorded for change watchers.
type watchedHead struct {
	// seqno is the World seqno of the head.
	seqno uint64
	// root is the root cursor of the head.
	root *bucket_lookup.Cursor
}

// WatchChanges calls cb with each later seqno and the changes since the
// previous one. The changes come from comparing the World trees of the head the
// watcher last saw against the current head, so they need no changelog.
//
// The Engine records the last changeWatchHeads heads it published, starting
// with the head it opened at, so a watch from the current seqno is precise.
// Recording does not pin a head, so superseded Worlds stay collectible, and a
// base the volume collected reports unknown changes. The ChangeSet is also
// unknown when the head at the previous seqno is no longer
// recorded, when two recorded heads share that seqno, or when the change is
// too large to list. Calls coalesce while cb runs.
func (e *Engine) WatchChanges(
	ctx context.Context,
	afterSeqno uint64,
	cb func(seqno uint64, changes *world.ChangeSet) error,
) error {
	for {
		// Wait for the head to pass the base seqno.
		if _, err := e.WaitSeqno(ctx, afterSeqno+1); err != nil {
			return err
		}

		// Hold the current head and pin the recorded head at the base seqno.
		seqno, head, base, release, err := e.holdWatchedHeads(ctx, afterSeqno)
		if err != nil {
			return err
		}
		if seqno <= afterSeqno {
			release()
			continue
		}

		// Compare the two heads outside the publication lock.
		changes := world.NewUnknownChangeSet()
		if base != nil {
			changes, err = e.compareHeads(ctx, base, head)
		}
		release()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			e.le.WithError(err).Warn("world change comparison failed; reporting unknown changes")
		}

		// Report the changes and advance the base.
		if err := cb(seqno, changes); err != nil {
			return err
		}
		afterSeqno = seqno
	}
}

// holdWatchedHeads holds the current head and pins the recorded head at
// afterSeqno until release. base is nil when no single recorded head has that
// seqno or it can no longer be pinned.
func (e *Engine) holdWatchedHeads(ctx context.Context, afterSeqno uint64) (
	seqno uint64,
	head, base *bucket_lookup.Cursor,
	release func(),
	err error,
) {
	// Read the current head's seqno under the publication lock.
	locked := e.bcast.Lock()
	if e.closed {
		locked.Unlock()
		return 0, nil, nil, nil, ErrEngineClosed
	}
	if err := e.initializeHeadReadTx(ctx); err != nil {
		locked.Unlock()
		return 0, nil, nil, nil, err
	}
	seqno, err = e.head.readTx.state.GetSeqno(ctx)
	if err != nil {
		locked.Unlock()
		return 0, nil, nil, nil, err
	}

	// Hold the head and clone the recorded root at the base seqno.
	head, releaseHead := e.holdHeadLocked()
	if found := e.watchedHeadLocked(afterSeqno); found != nil {
		base = found.Clone()
	}
	locked.Unlock()
	if base == nil {
		return seqno, head, nil, releaseHead, nil
	}

	// Pin the base outside the lock without waiting for a peer. A base the
	// volume already collected cannot be compared.
	pinCtx, _ := block.WithoutPeerWait(ctx)
	unpinBase, err := block.PinRoot(pinCtx, e.writeBlockStore, base.GetRef().GetRootRef())
	if err != nil {
		base.Release()
		return seqno, head, nil, releaseHead, nil
	}
	return seqno, head, base, func() {
		base.Release()
		unpinBase()
		releaseHead()
	}, nil
}

// watchedHeadLocked returns the one recorded root at seqno, or nil when the
// seqno is below the floor, unrecorded, or shared by different roots. The
// caller must hold bcast.
func (e *Engine) watchedHeadLocked(seqno uint64) *bucket_lookup.Cursor {
	if seqno < e.watchedFloor {
		return nil
	}
	var found *bucket_lookup.Cursor
	for _, h := range e.watchedHeads {
		if h.seqno != seqno {
			continue
		}
		if found != nil && !found.GetRef().EqualsRef(h.root.GetRef()) {
			return nil
		}
		found = h.root
	}
	return found
}

// recordInitialHead records the head the Engine opens at. NewEngine calls it
// before the Engine is shared.
func (e *Engine) recordInitialHead(ctx context.Context) error {
	seqno, err := e.head.readTx.state.GetSeqno(ctx)
	if err != nil {
		return err
	}
	e.recordWatchedHeadLocked(e.head.root, seqno)
	return nil
}

// recordWatchedHeadLocked records a newly published head for change watchers.
// It returns the evicted oldest head for release after the lock. Eviction
// raises the floor above the evicted seqno, so no watcher compares against
// another head that shared it. The caller must hold bcast.
func (e *Engine) recordWatchedHeadLocked(root *bucket_lookup.Cursor, seqno uint64) *watchedHead {
	// Evict the oldest head at capacity.
	var evicted *watchedHead
	if len(e.watchedHeads) == changeWatchHeads {
		evicted = e.watchedHeads[0]
		e.watchedFloor = max(e.watchedFloor, evicted.seqno+1)
		e.watchedHeads = slices.Delete(e.watchedHeads, 0, 1)
	}
	e.watchedHeads = append(e.watchedHeads, &watchedHead{seqno: seqno, root: root.Clone()})
	return evicted
}

// compareHeads lists the changes between two pinned World roots.
func (e *Engine) compareHeads(ctx context.Context, before, after *bucket_lookup.Cursor) (*world.ChangeSet, error) {
	_, beforeBcs := before.BuildTransactionWithStore(nil, e.writeBlockStore)
	_, afterBcs := after.BuildTransactionWithStore(nil, e.writeBlockStore)
	return WorldChanges(ctx, beforeBcs, afterBcs, changeWatchDiffLimit)
}

// _ is a type assertion
var _ world.ChangeWatcher = (*Engine)(nil)
