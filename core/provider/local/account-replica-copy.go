package provider_local

import (
	"context"
	"time"

	"github.com/aperturerobotics/util/routine"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/sobject"
	sobject_world_engine "github.com/s4wave/spacewave/core/sobject/world/engine"
	space_world_optypes "github.com/s4wave/spacewave/core/space/world/optypes"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
	dex_solicit "github.com/s4wave/spacewave/db/dex/solicit"
	"github.com/s4wave/spacewave/db/kvtx"
)

// runAccountReplicaCopy hydrates the published graph of every accepted state
// through the Session's existing DEX read-through store: the checkpoint's World
// and the payloads of the operations above it, which together hold every block
// the World after those operations reads. Nothing is replayed; the Worlds the
// operations derive are rebuilt from these blocks by whoever replays them.
// Cached blocks survive cancellation and restart; a persisted completion record
// is valid only for its exact immutable head.
//
// A block that no connected peer holds, or holds only without its refs, is not
// a failure: the copy logs it once and waits for a new state, or for a peer
// session to start on exchange that may hold the block. exchange is nil when
// the bucket has no DEX controller. Any other failure is published as the
// object's copy error before the routine retries.
func (a *ProviderAccount) runAccountReplicaCopy(
	ctx context.Context,
	so sobject.SharedObject,
	state *p2pSyncState,
	exchange *dex_solicit.Controller,
) (rerr error) {
	// Publish a failure, such as exhausted storage, where the sync status
	// reads it.
	defer func() {
		if rerr != nil && ctx.Err() == nil {
			state.publishCopyError(so.GetSharedObjectID(), rerr)
		}
	}()

	// Open the copy's local progress store and the object state stream.
	local, release, err := so.AccessLocalStateStore(ctx, "account-replica-copy", nil)
	if err != nil {
		return err
	}
	defer release()
	states, releaseStates, err := so.AccessSharedObjectState(ctx, nil)
	if err != nil {
		return err
	}
	defer releaseStates()

	// Replacing the head cancels and joins the previous copy before starting
	// the next one. A missing old block cannot pin replication to an old head.
	copier := routine.NewRoutineContainerWithLogger(a.le.WithField("routine", "copy-world-head"), routine.WithRetry(providerBackoff))
	copier.SetContext(ctx, false)
	defer func() {
		if exited, _ := copier.SetRoutine(nil); exited != nil {
			<-exited
		}
	}()

	// Copy each new published graph as its state arrives.
	var snapshot sobject.SharedObjectStateSnapshot
	var previous *block.BlockRef
	for {
		snapshot, err = states.WaitValueChange(ctx, snapshot, nil)
		if err != nil {
			return err
		}
		if snapshot == nil {
			continue
		}
		span, root, err := sobject_world_engine.PublishedGraph(ctx, snapshot, space_world_optypes.LookupWorldOp)
		if err != nil {
			return err
		}
		if root.GetEmpty() || root.EqualsRef(previous) {
			continue
		}
		previous = root

		// Store the span block on each attempt, so a sweep between attempts
		// cannot leave the copy without its root.
		head := &bucket.ObjectRef{RootRef: root}
		copier.SetRoutine(func(ctx context.Context) error {
			if span != nil {
				if _, err := so.GetBlockStore().PutBlockBatch(ctx, []*block.PutBatchEntry{span}); err != nil {
					return err
				}
			}
			return a.copyAccountWorldHead(ctx, so, state, local, head, exchange)
		})
	}
}

// copyAccountWorldHead copies head until it completes. A block that no
// reachable source holds, or holds only without its refs, does not appear by
// retrying the same head: the copy logs it once and retries when a peer
// session starts on exchange. Without exchange it waits for the next head,
// which cancels this routine.
func (a *ProviderAccount) copyAccountWorldHead(
	ctx context.Context,
	so sobject.SharedObject,
	state *p2pSyncState,
	local kvtx.Store,
	head *bucket.ObjectRef,
	exchange *dex_solicit.Controller,
) error {
	for logged := false; ; logged = true {
		// Count peer sessions first so one starting during the copy still
		// wakes the wait.
		var starts uint64
		if exchange != nil {
			starts = exchange.GetSessionStarts()
		}
		err := a.copyAndPersistAccountWorld(ctx, so, state, local, head)
		if !errors.Is(err, block.ErrNotFound) && !errors.Is(err, block.ErrRefsUnknown) {
			return err
		}

		// Wait for a source that may hold the block with its refs.
		if !logged {
			a.le.WithError(err).Info("waiting for a peer that holds a World block with its refs")
		}
		if exchange == nil {
			<-ctx.Done()
			return context.Cause(ctx)
		}
		if err := exchange.WaitSessionStart(ctx, starts); err != nil {
			return err
		}
	}
}

// copyAndPersistAccountWorld records completion only after the block fence.
func (a *ProviderAccount) copyAndPersistAccountWorld(ctx context.Context, so sobject.SharedObject, state *p2pSyncState, local kvtx.Store, head *bucket.ObjectRef) error {
	// Read the persisted progress record for this head.
	progress := &AccountReplicaCopyState{}
	err := kvtx.RunTransaction(ctx, false, func(ctx context.Context) (kvtx.Tx, error) { return local.NewTransaction(ctx, false) }, func(ctx context.Context, tx kvtx.Tx) error {
		data, found, err := tx.Get(ctx, []byte("account-replica-copy/progress"))
		if err == nil && found {
			err = progress.UnmarshalVT(data)
		}
		return err
	})
	if err != nil {
		return err
	}

	// Retain the head and republish when this head already completed.
	if progress.GetComplete() && progress.GetHead().EqualVT(head) {
		if err := block.SetRetainedRoot(ctx, so.GetBlockStore(), "account-world", head.GetRootRef()); err != nil {
			return err
		}
		state.publishCopyProgress(progress)
		return nil
	}

	// Copy the World and retain its root on success.
	progress = &AccountReplicaCopyState{ObjectId: so.GetSharedObjectID(), Head: head.CloneVT()}
	state.publishCopyProgress(progress)
	err = a.copyAccountWorld(ctx, so, progress, func() { state.publishCopyProgress(progress) })
	if err == nil {
		err = block.SetRetainedRoot(ctx, so.GetBlockStore(), "account-world", head.GetRootRef())
	}

	// Record the outcome in the progress state.
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		progress.Error = err.Error()
	} else {
		progress.Complete = true
	}

	// The block fence precedes this transaction, so a crash cannot retain
	// a completion claim whose data was still buffered.
	persistErr := kvtx.RunTransaction(ctx, true, func(ctx context.Context) (kvtx.Tx, error) { return local.NewTransaction(ctx, true) }, func(ctx context.Context, tx kvtx.Tx) error {
		data, err := progress.MarshalVT()
		if err != nil {
			return err
		}
		return tx.Set(ctx, []byte("account-replica-copy/progress"), data)
	})
	if persistErr != nil {
		progress.Complete = false
		progress.Error = persistErr.Error()
	}

	// Publish the final progress and report the copy or persistence error.
	state.publishCopyProgress(progress)
	if err != nil {
		return err
	}
	return persistErr
}

// copyAccountWorld shares durable graph retention with World publication.
// Its private proof store belongs to the serialized account replica copy routine.
func (a *ProviderAccount) copyAccountWorld(ctx context.Context, so sobject.SharedObject, progress *AccountReplicaCopyState, changed func()) error {
	// Open the shared retention store for the copy.
	local, release, err := so.AccessLocalStateStore(ctx, "account-replica-retention", nil)
	if err != nil {
		return err
	}
	defer release()

	// Retain the World head, publishing progress while blocks copy.
	lastUpdate := time.Now()
	return sobject_world_engine.RetainWorld(ctx, so, progress.GetHead(), local, func(_ *block.BlockRef, data []byte) {
		progress.Blocks++
		progress.Bytes += uint64(len(data))
		if time.Since(lastUpdate) >= 250*time.Millisecond {
			changed()
			lastUpdate = time.Now()
		}
	}, nil)
}

// publishCopyProgress snapshots mutable worker progress under the state lock.
func (s *p2pSyncState) publishCopyProgress(progress *AccountReplicaCopyState) {
	s.bcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		if s.copyProgress == nil {
			s.copyProgress = make(map[string]*AccountReplicaCopyState)
		}
		s.copyProgress[progress.GetObjectId()] = progress.CloneVT()
		bcast()
	})
}

// publishCopyError records err on the object's latest copy progress. A
// complete copy stays complete: its blocks are durable.
func (s *p2pSyncState) publishCopyError(objectID string, err error) {
	s.bcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		// Mark the latest progress failed and wake the status watchers.
		progress := s.copyProgress[objectID].CloneVT()
		if progress == nil {
			progress = &AccountReplicaCopyState{ObjectId: objectID}
		}
		progress.Error = err.Error()
		if s.copyProgress == nil {
			s.copyProgress = make(map[string]*AccountReplicaCopyState)
		}
		s.copyProgress[objectID] = progress
		bcast()
	})
}

// GetAccountCopyProgress returns isolated progress snapshots and the next change.
func (a *ProviderAccount) GetAccountCopyProgress() ([]*AccountReplicaCopyState, <-chan struct{}) {
	var out []*AccountReplicaCopyState
	var wait <-chan struct{}
	a.p2pSyncBcast.HoldLock(func(_ func(), getWait func() <-chan struct{}) {
		wait = getWait()
		if state := a.p2pSync; state != nil {
			state.bcast.HoldLock(func(_ func(), getWait func() <-chan struct{}) {
				wait = getWait()
				for _, progress := range state.copyProgress {
					out = append(out, progress.CloneVT())
				}
			})
		}
	})
	return out, wait
}
