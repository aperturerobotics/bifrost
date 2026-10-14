package sobject_world_engine

import (
	"context"
	"errors"

	"github.com/aperturerobotics/util/ccontainer"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/block"
	trace "github.com/s4wave/spacewave/db/traceutil"
)

// executeWatchSOState watches and processes shared object state changes and
// settles the receipts of admitted writes as they resolve.
func (c *Controller) executeWatchSOState(
	ctx context.Context,
	soStateCtr ccontainer.Watchable[sobject.SharedObjectStateSnapshot],
	soEngine *soEngine,
) error {
	var snap sobject.SharedObjectStateSnapshot
	var missing *block.BlockRef
	var settled <-chan struct{}
	for {
		// Wait for the state to change, the block replay stopped at to arrive,
		// or the oldest pending receipt to resolve.
		_, err := c.waitReplayInput(ctx, soEngine.so, soStateCtr, snap, missing, settled)
		if err != nil {
			return err
		}

		// Lock the writeMtx, so that we wait until any write txn is done processing first.
		lockCtx, lockTask := trace.NewTask(ctx, "alpha/watch-state/lock-write-mtx")
		unlockWriteMtx, err := c.writeMtx.Lock(lockCtx)
		lockTask.End()
		if err != nil {
			return err
		}

		// Separate lock acquisition from hold time so traces show contention vs work.
		holdCtx, holdTask := trace.NewTask(ctx, "alpha/watch-state/hold-write-mtx")

		// Get the latest snap in case it changed in the meantime, settle the
		// resolved receipts, and skip a snapshot of a failed publication.
		snap = soStateCtr.GetValue()
		soEngine.settle()
		stale, err := soEngine.stale(holdCtx, snap)

		// Watch the state once (sync any changes to soEngine and update local state).
		if err == nil && !stale {
			err = c.executeWatchSOStateOnce(holdCtx, snap, soEngine)
		}
		missing = soEngine.replay.missing
		settled = soEngine.settled()

		// Be sure to unlock the writeMtx right away.
		holdTask.End()
		unlockWriteMtx()

		// Return the error, if any.
		if err != nil {
			return err
		}
	}
}

// waitReplayInput waits for the state to change from snap and returns the new
// state. When missing is set, it also returns the current state once a peer
// serves that block, since replay stopped at it. When settled is set, it also
// returns the current state once settled closes.
func (c *Controller) waitReplayInput(
	ctx context.Context,
	so sobject.SharedObject,
	soStateCtr ccontainer.Watchable[sobject.SharedObjectStateSnapshot],
	snap sobject.SharedObjectStateSnapshot,
	missing *block.BlockRef,
	settled <-chan struct{},
) (sobject.SharedObjectStateSnapshot, error) {
	// Without another input only a state change resumes replay.
	if missing == nil && settled == nil {
		return soStateCtr.WaitValueChange(ctx, snap, nil)
	}

	// End the state wait once settled closes.
	waitCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if settled != nil {
		go func() {
			select {
			case <-settled:
				cancel()
			case <-waitCtx.Done():
			}
		}()
	}

	// Fetch the block with a read that waits for a peer, ending the state
	// wait once it arrives.
	fetched := make(chan struct{})
	go func() {
		defer close(fetched)
		if missing == nil {
			return
		}
		_, found, err := so.GetBlockStore().GetBlock(waitCtx, missing)
		if err != nil && waitCtx.Err() == nil {
			c.le.WithError(err).Warn("fetch the block replay stopped at")
		}
		if found {
			cancel()
		}
	}()

	// Wait for a state change, then stop the fetch.
	next, err := soStateCtr.WaitValueChange(waitCtx, snap, nil)
	cancel()
	<-fetched

	// Another input ends the wait with the current state.
	if err != nil && ctx.Err() == nil && errors.Is(err, context.Canceled) {
		return soStateCtr.GetValue(), nil
	}
	return next, err
}

// executeWatchSOStateOnce replays one shared object state into the World. A
// block that is not available stops replay at its operation without ending the
// watch: the World before that operation stays served, and the next state
// change, write, or arrival of the block resumes replay there.
func (c *Controller) executeWatchSOStateOnce(
	ctx context.Context,
	snap sobject.SharedObjectStateSnapshot,
	soEngine *soEngine,
) error {
	// Trace the pass.
	ctx, task := trace.NewTask(ctx, "alpha/watch-state/process-snapshot")
	defer task.End()

	// Replay, waiting out a block that is not available.
	_, err := soEngine.advance(ctx, snap, nil)
	if block.IsNotAvailable(err) {
		c.le.WithError(err).Warn("replay waits for a block that is not available")
		return nil
	}
	return err
}

// isReadAccessLoss reports whether err means this participant cannot replay
// the state, which a later grant or readmission can restore.
func isReadAccessLoss(err error) bool {
	return errors.Is(err, sobject.ErrCannotDecode) ||
		errors.Is(err, sobject.ErrNotParticipant) ||
		errors.Is(err, sobject.ErrKeyEpochUnavailable) ||
		errors.Is(err, sobject.ErrConfigHistoryUnavailable)
}

// waitReadableSnapshot waits for a snapshot this participant can replay: it
// decodes the checkpoint and holds the config and key of every operation.
// Other snapshot errors end the wait so the caller surfaces them.
func waitReadableSnapshot(ctx context.Context, soStateCtr ccontainer.Watchable[sobject.SharedObjectStateSnapshot]) error {
	_, err := soStateCtr.WaitValueWithValidator(ctx, func(snap sobject.SharedObjectStateSnapshot) (bool, error) {
		if snap == nil {
			return false, nil
		}
		return !isReadAccessLoss(checkReplayable(ctx, snap)), nil
	}, nil)
	return err
}

// checkReplayable returns the first error replaying snap would meet reading
// its checkpoint, configs and keys.
func checkReplayable(ctx context.Context, snap sobject.SharedObjectStateSnapshot) error {
	// The participant must decode the checkpoint.
	if _, err := snap.GetParticipantConfig(ctx); err != nil {
		return err
	}
	if _, err := snap.GetCheckpoint(ctx); err != nil {
		return err
	}

	// Every operation needs its config and its epoch key.
	set, err := snap.GetOperationSet(ctx)
	if err != nil {
		return err
	}
	for _, h := range set.Order() {
		inner := set.Get(h)
		if _, err := snap.GetConfigByHash(ctx, inner.GetConfigHash()); err != nil {
			return err
		}
		if _, err := snap.DecodeOperation(ctx, inner); err != nil && isReadAccessLoss(err) {
			return err
		}
	}
	return nil
}
