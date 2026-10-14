package world_block

import (
	"context"
	"slices"

	"github.com/s4wave/spacewave/db/block"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/world"
)

// changeWatchHeads bounds the published heads the Engine records so a new
// change subscriber can compare the head it last saw against a later one.
const changeWatchHeads = 128

// watchedHead is a published head recorded for change watchers.
type watchedHead struct {
	// seqno is the World seqno of the head.
	seqno uint64
	// root is the root cursor of the head.
	root *bucket_lookup.Cursor
}

// changeSubscriber is one WatchChanges call on the change watch. Its fields
// are guarded by the Engine bcast lock.
type changeSubscriber struct {
	// seqno is the seqno of the newest head merged into changes.
	seqno uint64
	// changes merges the changes not yet reported, or is nil.
	changes *world.ChangeSet
	// wake receives a value when changes is set or the watch loses track.
	wake chan struct{}
}

// WatchChanges calls cb with each later seqno and the changes since the
// previous one. The changes come from the Engine's change watch, which
// compares each new head with the last compared one once for every watcher,
// so they need no changelog. Calls coalesce while cb runs. cb must not modify
// the ChangeSet, which other subscribers share.
//
// A subscriber that starts behind the change watch first compares the head at
// afterSeqno against the published head. The Engine records the last
// changeWatchHeads heads it published, starting with the head it opened at, so
// a recent afterSeqno is precise. Recording does not pin a head, so superseded
// Worlds stay collectible, and a base the volume collected reports unknown
// changes. The ChangeSet is also unknown when the head at afterSeqno is no
// longer recorded, when two recorded heads share that seqno, or when the
// change is too large to list.
func (e *Engine) WatchChanges(
	ctx context.Context,
	afterSeqno uint64,
	cb func(seqno uint64, changes *world.ChangeSet) error,
) error {
	// Subscribe to the change watch.
	sub := &changeSubscriber{wake: make(chan struct{}, 1)}
	locked := e.bcast.Lock()
	w, err := e.startChangeWatchLocked()
	if err != nil {
		locked.Unlock()
		return err
	}
	w.subs[sub] = struct{}{}
	watchSeqno := w.baseSeqno
	locked.Unlock()
	defer func() {
		locked := e.bcast.Lock()
		delete(w.subs, sub)
		w.checkIdleLocked()
		locked.Unlock()
	}()

	// Catch up from afterSeqno when the change watch starts past it.
	if afterSeqno < watchSeqno {
		seqno, changes, err := e.catchUpChanges(ctx, afterSeqno)
		if err != nil {
			return err
		}
		locked := e.bcast.Lock()
		sub.addLocked(seqno, changes)
		locked.Unlock()
	}

	// Report each change past afterSeqno.
	for {
		// Take the merged changes, or wait for the next.
		locked := e.bcast.Lock()
		if e.closed {
			locked.Unlock()
			return ErrEngineClosed
		}
		seqno, changes := sub.seqno, sub.changes
		sub.changes = nil
		watchErr := e.headWatchErr
		locked.Unlock()
		if changes == nil {
			if watchErr != nil {
				return watchErr
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-sub.wake:
			}
			continue
		}

		// Skip changes the caller already saw.
		if seqno <= afterSeqno {
			continue
		}
		if err := cb(seqno, changes); err != nil {
			return err
		}
		afterSeqno = seqno
	}
}

// catchUpChanges compares the recorded head at afterSeqno with the published
// head and returns the published head's seqno and the changes.
func (e *Engine) catchUpChanges(ctx context.Context, afterSeqno uint64) (uint64, *world.ChangeSet, error) {
	// Hold the published head and the recorded base.
	seqno, head, base, release, err := e.holdWatchedHeads(ctx, afterSeqno)
	if err != nil {
		return 0, nil, err
	}
	defer release()
	if base == nil {
		return seqno, world.NewUnknownChangeSet(), nil
	}

	// Compare the two heads outside the publication lock.
	_, beforeBcs := base.BuildTransactionWithStore(nil, e.writeBlockStore)
	_, afterBcs := head.BuildTransactionWithStore(nil, e.writeBlockStore)
	changes, err := WorldChanges(ctx, beforeBcs, afterBcs, changeDiffLimit)
	if err != nil {
		if ctx.Err() != nil {
			return 0, nil, ctx.Err()
		}
		e.le.WithError(err).Warn("world change comparison failed; reporting unknown changes")
	}
	return seqno, changes, nil
}

// holdWatchedHeads holds the published head and pins the recorded head at
// afterSeqno until release. base is nil when no single recorded head has that
// seqno or it can no longer be pinned.
func (e *Engine) holdWatchedHeads(ctx context.Context, afterSeqno uint64) (
	seqno uint64,
	head, base *bucket_lookup.Cursor,
	release func(),
	err error,
) {
	// Hold the head and clone the recorded root at the base seqno.
	locked := e.bcast.Lock()
	if e.closed {
		locked.Unlock()
		return 0, nil, nil, nil, ErrEngineClosed
	}
	seqno = e.headSeqnoLocked()
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

// addLocked merges the changes up to the head at seqno and wakes the
// subscriber. The caller must hold bcast.
func (s *changeSubscriber) addLocked(seqno uint64, changes *world.ChangeSet) {
	s.seqno = max(s.seqno, seqno)
	s.changes = mergeChangeSets(s.changes, changes)
	s.signal()
}

// signal wakes the subscriber without blocking.
func (s *changeSubscriber) signal() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// mergeChangeSets returns the union of two ChangeSets without modifying
// either. The union is unknown when either is, or when it lists more than
// changeDiffLimit keys.
func mergeChangeSets(a, b *world.ChangeSet) *world.ChangeSet {
	// Share a lone ChangeSet and keep unknown sets unknown.
	if a == nil {
		return b
	}
	if a.Unknown || b.Unknown {
		return world.NewUnknownChangeSet()
	}

	// Join the keys without duplicates, and the quads.
	keys := slices.Concat(a.Keys, b.Keys)
	slices.Sort(keys)
	keys = slices.Compact(keys)
	if len(keys) > changeDiffLimit {
		return world.NewUnknownChangeSet()
	}
	return &world.ChangeSet{Keys: keys, Quads: slices.Concat(a.Quads, b.Quads)}
}

// _ is a type assertion
var _ world.ChangeWatcher = (*Engine)(nil)
