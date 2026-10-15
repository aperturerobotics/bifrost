package provider_local

import (
	"context"

	"github.com/aperturerobotics/util/ccontainer"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/block"
	block_transform "github.com/s4wave/spacewave/db/block/transform"
	trace "github.com/s4wave/spacewave/db/traceutil"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// LocalSOHost publishes the local participant's view of a shared object and
// adds the local participant's operations to it.
type LocalSOHost struct {
	// le is the logger.
	le *logrus.Entry
	// privKey signs local operations.
	privKey crypto.PrivKey
	// peerID identifies the local participant.
	peerID peer.ID
	// sharedObjectID is the shared object id.
	sharedObjectID string
	// sfs builds the block transforms of key epochs.
	sfs *block_transform.StepFactorySet
	// soHost owns the shared object state.
	soHost *sobject.SOHost
	// stateSnapCtr holds the snapshot of the latest durable host state.
	stateSnapCtr *ccontainer.CContainer[sobject.SharedObjectStateSnapshot]
	// admittedSnapCtr holds the snapshot of the latest admitted host state.
	admittedSnapCtr *ccontainer.CContainer[sobject.SharedObjectStateSnapshot]
	// publishedConfigCtr holds the config of the latest published snapshot.
	publishedConfigCtr *ccontainer.CContainer[*sobject.SharedObjectConfig]
	// lifetime ends when the mount that runs Execute stops. A commit that
	// started ends with it.
	lifetime context.Context
}

// NewLocalSOHost constructs a new LocalSOHost. lifetime is the context its
// owner runs Execute under.
func NewLocalSOHost(
	lifetime context.Context,
	le *logrus.Entry,
	privKey crypto.PrivKey,
	soHost *sobject.SOHost,
	sharedObjectID string,
	sfs *block_transform.StepFactorySet,
) (*LocalSOHost, error) {
	peerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	return &LocalSOHost{
		le:                 le,
		privKey:            privKey,
		peerID:             peerID,
		sharedObjectID:     sharedObjectID,
		sfs:                sfs,
		soHost:             soHost,
		stateSnapCtr:       ccontainer.NewCContainer[sobject.SharedObjectStateSnapshot](nil),
		admittedSnapCtr:    ccontainer.NewCContainer[sobject.SharedObjectStateSnapshot](nil),
		publishedConfigCtr: ccontainer.NewCContainer[*sobject.SharedObjectConfig](nil),
		lifetime:           lifetime,
	}, nil
}

// Execute publishes a snapshot of every durable and every admitted host
// state until ctx is canceled.
func (l *LocalSOHost) Execute(ctx context.Context) error {
	// Watch the durable and the admitted host state.
	stateCtr, relStateCtr, err := l.soHost.GetSOStateCtr(ctx, nil)
	if err != nil {
		return err
	}
	defer relStateCtr()
	admittedCtr, relAdmittedCtr, err := l.soHost.GetAdmittedSOStateCtr(ctx, nil)
	if err != nil {
		return err
	}
	defer relAdmittedCtr()

	// Publish a snapshot of each, and the config of each durable state.
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return l.publishSnapshots(ctx, stateCtr, l.stateSnapCtr, func(state *sobject.SOState) {
			l.publishedConfigCtr.SetValue(state.GetConfig().CloneVT())
		})
	})
	g.Go(func() error {
		return l.publishSnapshots(ctx, admittedCtr, l.admittedSnapCtr)
	})
	return g.Wait()
}

// buildSnapshot returns the local participant's snapshot of state.
func (l *LocalSOHost) buildSnapshot(state *sobject.SOState) *sobject.SOStateParticipantHandle {
	return sobject.NewSOStateParticipantHandle(
		l.le,
		l.sfs,
		l.sharedObjectID,
		state,
		l.privKey,
		l.peerID,
	).WithConfigHistory(l.soHost.ReadConfigEntry).WithOperationVerifier(l.soHost.GetOperationVerifier())
}

// waitPublishedConfig waits until body readers can observe target or a verified descendant.
func (l *LocalSOHost) waitPublishedConfig(ctx context.Context, target *sobject.SharedObjectConfig) error {
	if target == nil {
		return errors.New("published SharedObject configuration is unavailable")
	}

	// Accept the target configuration or a descendant with verified history.
	_, err := l.publishedConfigCtr.WaitValueWithValidator(ctx, func(current *sobject.SharedObjectConfig) (bool, error) {
		// Wait for the target's sequence number, then accept the target itself.
		if current == nil || current.GetConfigChainSeqno() < target.GetConfigChainSeqno() {
			return false, nil
		}
		if current.EqualVT(target) {
			return true, nil
		}
		changes, err := l.soHost.ReadConfigHistory(ctx, target.GetConfigChainHash(), current.GetConfigChainHash())
		if err != nil {
			return false, err
		}
		if err := sobject.VerifyConfigChainSuffix(l.sharedObjectID, target, current, changes); err != nil {
			return false, err
		}
		return true, nil
	}, nil)
	return err
}

// AccessSharedObjectState adds a reference to the durable state and returns
// the state container.
func (l *LocalSOHost) AccessSharedObjectState(ctx context.Context, released func()) (ccontainer.Watchable[sobject.SharedObjectStateSnapshot], func(), error) {
	return l.stateSnapCtr, func() {}, nil
}

// AccessAdmittedState adds a reference to the admitted state and returns the
// state container.
func (l *LocalSOHost) AccessAdmittedState(ctx context.Context, released func()) (ccontainer.Watchable[sobject.SharedObjectStateSnapshot], func(), error) {
	return l.admittedSnapCtr, func() {}, nil
}

// publishSnapshots sets snaps to the snapshot of each state of states until
// ctx is canceled, after calling each, when set, with the state.
func (l *LocalSOHost) publishSnapshots(
	ctx context.Context,
	states ccontainer.Watchable[*sobject.SOState],
	snaps *ccontainer.CContainer[sobject.SharedObjectStateSnapshot],
	each ...func(*sobject.SOState),
) error {
	var state *sobject.SOState
	for {
		var err error
		state, err = states.WaitValueChange(ctx, state, nil)
		if err != nil {
			return err
		}
		snaps.SetValue(l.buildSnapshot(state))
		for _, fn := range each {
			fn(state)
		}
	}
}

// detach returns ctx's values without its cancellation, ended when the host's
// lifetime ends. The caller must call cancel.
func (l *LocalSOHost) detach(ctx context.Context) (context.Context, context.CancelFunc) {
	detached, cancel := context.WithCancel(context.WithoutCancel(ctx))
	stop := context.AfterFunc(l.lifetime, cancel)
	return detached, func() {
		stop()
		cancel()
	}
}

// QueueOperation signs op as the local participant and adds it to the
// operation set. When ctx carries sobject.WithPublishState, an ordinary state
// write is admitted to that publication, and QueueOperation returns the local
// operation id once the admitted snapshot holds the operation, or the local
// operation id with the publication's error once its receipt fails. Otherwise
// it returns once the durable and the admitted snapshots hold it: the host
// publishes them apart, and the World engine replays the admitted one.
func (l *LocalSOHost) QueueOperation(ctx context.Context, op []byte) (string, error) {
	// Capture the receipt of an admitted write: its failure removes the
	// operation from the admitted state.
	ctx, task := trace.NewTask(ctx, "alpha/local-so/queue-operation")
	defer task.End()
	var receipt *block.PublicationReceipt
	snaps := []*ccontainer.CContainer[sobject.SharedObjectStateSnapshot]{l.stateSnapCtr, l.admittedSnapCtr}
	if publish := sobject.GetPublishState(ctx); publish != nil {
		snaps = snaps[1:]
		ctx = sobject.WithPublishState(ctx, func(ctx context.Context, head *block.AtomicHeadUpdate) (*block.PublicationReceipt, error) {
			var err error
			receipt, err = publish(ctx, head)
			return receipt, err
		})
	}

	// Sign the operation and add it to the host state.
	localID, nonce, err := l.soHost.AddLocalOperation(ctx, l.le, l.sfs, l.privKey, op)
	if err != nil {
		return "", err
	}

	// Wait for the snapshots that hold the operation, or a later checkpoint.
	// The state write admitted or wrote the operation, so the wait ignores
	// ctx's cancellation and the call reports the operation as added. It
	// ends with the host's lifetime, and the outcome is then unknown, or with
	// a failed receipt.
	ctx, cancel := l.detach(ctx)
	defer cancel()
	if receipt != nil {
		go func() {
			select {
			case <-receipt.Done():
				if receipt.Err() != nil {
					cancel()
				}
			case <-ctx.Done():
			}
		}()
	}
	holds := func(snap sobject.SharedObjectStateSnapshot) (bool, error) {
		// A snapshot holds the operation, or its checkpoint covers it.
		if snap == nil {
			return false, nil
		}
		set, err := snap.GetOperationSet(ctx)
		if err != nil {
			return false, err
		}
		peerID := l.peerID.String()
		return set.Find(peerID, localID) != nil || set.Covers(peerID, nonce), nil
	}
	for _, ctr := range snaps {
		if _, err = ctr.WaitValueWithValidator(ctx, holds, nil); err != nil {
			break
		}
	}
	if err == nil {
		return localID, nil
	}

	// Report the publication's failure over the canceled wait.
	if receipt != nil {
		select {
		case <-receipt.Done():
			if rerr := receipt.Err(); rerr != nil {
				return localID, rerr
			}
		default:
		}
	}
	return "", err
}
