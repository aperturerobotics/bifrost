package provider_local

import (
	"context"

	"github.com/aperturerobotics/util/ccontainer"
	"github.com/aperturerobotics/util/keyed"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/s4wave/spacewave/db/object"
	trace "github.com/s4wave/spacewave/db/traceutil"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/net/peer"
)

// NewObjectStoreSOStateFuncs constructs a SOHostState backed by an object store and in-memory locks.
//
// Assumes no other writers will access the object store.
// rctx is the context to use for looking up states from the object store.
// objStoreID names objStore in its volume, so an ordinary local operation can
// publish its state write with the World commit that queued it; empty commits
// every state write through objStore.
// localPeer binds private read checkpoints to the local participant; empty disables retention.
// watchFn watches the durable state; syncFuncs.Admitted watches the admitted
// state, which also holds the operations admitted to pending publications.
func NewObjectStoreSOStateFuncs(rctx context.Context, objStore object.ObjectStore, objStoreID string, localPeer peer.ID) (
	watchFn sobject.SOStateWatchFunc,
	lockFn sobject.SOStateLockFunc,
	syncFuncs *sobject.SOHostSyncFuncs,
) {
	// Load one state per retained object without reading its complete history,
	// and settle its admissions while it is retained.
	soRc := keyed.NewKeyedRefCount(
		func(sharedObjectID string) (keyed.Routine, *soStateEntry) {
			ent := newSOStateEntry(SobjectObjectStoreHostStateKey(sharedObjectID))
			return func(ctx context.Context) error {
				return ent.execute(ctx, objStore, sharedObjectID)
			}, ent
		},
		keyed.WithExitCb(func(sharedObjectID string, _ keyed.Routine, ent *soStateEntry, err error) {
			if err != nil {
				ent.stateProm.SetResult(nil, err)
			}
		}),
	)
	soRc.SetContext(rctx, true)

	// Bind a watched state of the entry to the caller's reference lifetime.
	watch := func(ctx context.Context, sharedObjectID string, pick func(*soStateEntry) *ccontainer.CContainer[*sobject.SOState]) (ccontainer.Watchable[*sobject.SOState], func(), error) {
		ref, ent, _ := soRc.AddKeyRef(sharedObjectID)
		_, err := ent.stateProm.Await(ctx)
		if err != nil {
			ref.Release()
			return nil, nil, err
		}
		return pick(ent), ref.Release, nil
	}
	watchFn = func(ctx context.Context, sharedObjectID string, _ func()) (ccontainer.Watchable[*sobject.SOState], func(), error) {
		return watch(ctx, sharedObjectID, func(ent *soStateEntry) *ccontainer.CContainer[*sobject.SOState] {
			return ent.stateCtr
		})
	}

	// Retain the state entry until the locked write has completed. An
	// ordinary local operation carrying a publication builds on the admitted
	// state and holds the lock until its publication is admitted; every other
	// write waits for the pending receipts and builds on the durable state.
	newLock := func(ctx context.Context, sharedObjectID string, checkpoint, peerImport bool) (sobject.SOStateLock, error) {
		// Retain the loaded state entry until the write lock is built.
		ref, ent, _ := soRc.AddKeyRef(sharedObjectID)
		_, err := ent.stateProm.Await(ctx)
		if err != nil {
			ref.Release()
			return nil, err
		}

		// Serialize writes after the initial state load succeeds.
		relLock, err := ent.writeMtx.Lock(ctx)
		if err != nil {
			ref.Release()
			return nil, err
		}

		// Read the state the write builds on.
		admission := sobject.GetPublishState(ctx) != nil && !checkpoint && !peerImport && objStoreID != ""
		base := ent.admitted()
		if !admission {
			base, err = ent.settled(ctx)
			if err != nil {
				relLock()
				ref.Release()
				return nil, err
			}
		}

		// Commit state and supplied configuration history before updating watchers.
		initialState := base.Clone()
		return sobject.NewSOStateLock(
			initialState,
			func(ctx context.Context, state *sobject.SOState, changes ...*sobject.SOConfigChange) error {
				// Serialize an independent state snapshot for replayable transaction writes.
				ctx, task := trace.NewTask(ctx, "alpha/local-so/write-so-state")
				defer task.End()
				state = state.Clone()
				data, err := state.MarshalVT()
				if err != nil {
					return err
				}

				// Admit an ordinary local operation's state to the World commit
				// that queued it. Its state write sets only the state key, and
				// its receipt settles the state into the durable state.
				if publish := sobject.GetPublishState(ctx); publish != nil {
					var head *block.AtomicHeadUpdate
					ordinary := !checkpoint && !peerImport && len(changes) == 0 && state.GetConfig().EqualVT(initialState.GetConfig())
					if ordinary && objStoreID != "" {
						head = newSOStateHead(objStoreID, ent.key, initialState, data)
					}
					receipt, err := publish(ctx, head)
					if err != nil {
						return err
					}
					if receipt != nil {
						admRef, _, _ := soRc.AddKeyRef(sharedObjectID)
						ent.admit(state, receipt, admRef.Release)
						return nil
					}
				}

				// Otherwise write after every pending receipt, on the state the
				// lock loaded. A failed receipt removed that state.
				durable, err := ent.settled(ctx)
				if err != nil {
					return err
				}
				if durable != base {
					return errors.Wrap(block.ErrPublicationDependency, "shared object state write built on a failed publication")
				}

				// Retain signed changes and state under the same commit boundary.
				err = kvtx.RunTransaction(ctx, true,
					func(ctx context.Context) (kvtx.Tx, error) {
						return objStore.NewTransaction(ctx, true)
					},
					func(ctx context.Context, tx kvtx.Tx) error {
						// Record the config history and the read checkpoint with the state.
						if checkpoint {
							if err := WriteSOConfigLineage(ctx, tx, sharedObjectID, state.GetConfig(), changes); err != nil {
								return err
							}
						}
						if peerImport {
							if err := WriteSOConfigCheckpoint(ctx, tx, sharedObjectID, initialState.GetConfig(), false); err != nil {
								return err
							}
						}
						if !checkpoint {
							if err := WriteSOConfigHistory(ctx, tx, sharedObjectID, initialState.GetConfig(), state.GetConfig(), changes); err != nil {
								return err
							}
						}
						if err := writeReadCheckpoint(ctx, tx, sharedObjectID, localPeer, initialState, state); err != nil {
							return err
						}
						return tx.Set(ctx, ent.key, data)
					},
				)
				if err != nil {
					return err
				}

				// Make the committed snapshot visible only after the transaction succeeds.
				ent.commit(state)
				return nil
			},
			func() {
				relLock()
				ref.Release()
			},
		), nil
	}

	// All three operations share the same retained state and serialization lock.
	lockFn = func(ctx context.Context, id string) (sobject.SOStateLock, error) {
		return newLock(ctx, id, false, false)
	}
	syncFuncs = NewSOHostSyncFuncs(objStore)
	syncFuncs.Lock = func(ctx context.Context, id string) (sobject.SOStateLock, error) {
		return newLock(ctx, id, false, true)
	}
	syncFuncs.CheckpointLock = func(ctx context.Context, id string) (sobject.SOStateLock, error) {
		return newLock(ctx, id, true, false)
	}
	syncFuncs.Admitted = func(ctx context.Context, sharedObjectID string, _ func()) (ccontainer.Watchable[*sobject.SOState], func(), error) {
		return watch(ctx, sharedObjectID, func(ent *soStateEntry) *ccontainer.CContainer[*sobject.SOState] {
			return ent.admittedCtr
		})
	}
	return watchFn, lockFn, syncFuncs
}

// newSOStateHead returns the head update that replaces the state at key in
// the object store objStoreID with data, after checking that the committed
// state is still previous.
func newSOStateHead(objStoreID string, key []byte, previous *sobject.SOState, data []byte) *block.AtomicHeadUpdate {
	return &block.AtomicHeadUpdate{
		ObjectStoreID: objStoreID,
		Key:           key,
		Replace: func(_ context.Context, current []byte, found bool) ([]byte, error) {
			// Require the state the write lock loaded.
			if !found {
				return nil, world.ErrObjectNotFound
			}
			committed := &sobject.SOState{}
			if err := committed.UnmarshalVT(current); err != nil {
				return nil, err
			}
			if !committed.EqualVT(previous) {
				return nil, errors.New("shared object state changed outside its write lock")
			}
			return data, nil
		},
	}
}
