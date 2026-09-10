package pairing

import (
	"context"
	"io"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/util/broadcast"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/session"
	"github.com/s4wave/spacewave/core/transport"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/link"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// Status matches the Session resource's pairing status enum.
type Status int32

const (
	StatusIdle Status = iota
	StatusCodeGenerated
	StatusWaitingForPeer
	StatusPeerConnected
	StatusVerifyingEmoji
	StatusVerified
	StatusFailed
	StatusSignalingFailed
	StatusConnectionTimeout
	StatusWaitingForRemote
	StatusBothConfirmed
	StatusPairingRejected
	StatusConfirmationTimeout
	StatusEnrolling
)

// Snapshot describes one Session's current pairing operation.
type Snapshot struct {
	Status       Status
	Code         string
	RemotePeerID peer.ID
	Emoji        []string
	ErrMsg       string
	AccountID    string
	AccountName  string
	ProviderID   string
	Receiving    bool
}

// Engine owns approval and enrollment for one mounted Session. Account adapters
// own provider authorization and persistence; its transport remains owned by the
// Session's ordinary transport lifecycle.
type Engine struct {
	ctx       context.Context
	le        *logrus.Entry
	b         bus.Bus
	key       crypto.PrivKey
	peerID    peer.ID
	adapter   AccountAdapter
	transport func(context.Context, Relay) (*transport.SessionTransport, error)
	bcast     broadcast.Broadcast
	active    *attempt
}

type attempt struct {
	snapshot Snapshot
	offering bool
	confirm  chan bool
	cancel   context.CancelFunc
	result   *session.SessionRef
	release  func()
}

// NewEngine binds a Session's key, provider adapter, and transport owner.
func NewEngine(ctx context.Context, le *logrus.Entry, b bus.Bus, key crypto.PrivKey, adapter AccountAdapter, getTransport func(context.Context, Relay) (*transport.SessionTransport, error)) (*Engine, error) {
	if ctx == nil || key == nil {
		return nil, errors.New("pairing requires an unlocked Session lifecycle")
	}
	peerID, err := peer.IDFromPrivateKey(key)
	if err != nil {
		return nil, err
	}
	e := &Engine{ctx: ctx, le: le, b: b, key: key, peerID: peerID, adapter: adapter, transport: getTransport}
	context.AfterFunc(ctx, e.Clear)
	return e, nil
}

// Context returns the mounted Session lifetime for direct signaling operations.
func (e *Engine) Context() context.Context { return e.ctx }

// begin cancels the prior attempt before publishing another operation identity.
func (e *Engine) begin(offering bool, code string, remote peer.ID, status Status) (context.Context, *attempt) {
	ctx, cancel := context.WithCancel(e.ctx)
	active := &attempt{offering: offering, cancel: cancel, confirm: make(chan bool, 1), snapshot: Snapshot{Status: status, Code: code, RemotePeerID: remote, Receiving: !offering}}
	var previous *attempt
	e.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		previous, e.active = e.active, active
		broadcast()
	})
	releaseAttempt(previous)
	return ctx, active
}

// Clear releases the active exchange and its temporary transport demand.
func (e *Engine) Clear() {
	var active *attempt
	e.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		active, e.active = e.active, nil
		broadcast()
	})
	releaseAttempt(active)
}

func releaseAttempt(active *attempt) {
	if active == nil {
		return
	}
	if active.cancel != nil {
		active.cancel()
	}
	if active.release != nil {
		active.release()
		active.release = nil
	}
}

func (e *Engine) retain(active *attempt, release func()) bool {
	var retained bool
	e.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if e.active == active {
			active.release = release
			retained = true
		}
	})
	if !retained {
		release()
	}
	return retained
}

func (e *Engine) releaseResources(active *attempt) {
	var release func()
	e.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if e.active == active {
			release, active.release = active.release, nil
		}
	})
	if release != nil {
		release()
	}
}

func (e *Engine) update(active *attempt, update func(*attempt)) {
	e.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		if e.active == active {
			update(active)
			broadcast()
		}
	})
}

func (e *Engine) fail(active *attempt, status Status, err error) {
	e.update(active, func(a *attempt) {
		a.snapshot.Status = status
		a.snapshot.ErrMsg = err.Error()
	})
}

// SetFailed publishes a direct signaling failure, including before a link exists.
func (e *Engine) SetFailed(message string) {
	var previous *attempt
	e.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		previous = e.active
		e.active = &attempt{snapshot: Snapshot{Status: StatusFailed, ErrMsg: message}}
		broadcast()
	})
	releaseAttempt(previous)
}

// ConfirmSAS submits the current screen's bilateral approval decision.
func (e *Engine) ConfirmSAS(confirmed bool) {
	e.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if e.active == nil || e.active.snapshot.Status != StatusVerifyingEmoji {
			return
		}
		select {
		case e.active.confirm <- confirmed:
		default:
		}
	})
}

// Watch observes coherent snapshots until cancellation or a callback error.
func (e *Engine) Watch(ctx context.Context, fn func(Snapshot) error) error {
	for {
		snapshot, wait := e.Snapshot()
		if err := fn(snapshot); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-wait:
		}
	}
}

// Snapshot returns coherent state and a channel that closes on the next change.
func (e *Engine) Snapshot() (Snapshot, <-chan struct{}) {
	var snapshot Snapshot
	var wait <-chan struct{}
	e.bcast.HoldLock(func(_ func(), getWait func() <-chan struct{}) {
		wait = getWait()
		if e.active != nil {
			snapshot = e.active.snapshot
			snapshot.Emoji = append([]string(nil), snapshot.Emoji...)
		}
	})
	return snapshot, wait
}

var (
	ErrExchangeMissing      = errors.New("pairing exchange is missing")
	ErrExchangeUnconfirmed  = errors.New("account enrollment is not complete")
	ErrExchangePeerMismatch = errors.New("pairing exchange peer does not match")
)

// Result returns the durable receiving attachment. The offering side returns
// nil. Repeated reads cannot grant more access or enroll another Session.
func (e *Engine) Result(remote peer.ID) (*session.SessionRef, error) {
	var result *session.SessionRef
	var err error
	e.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		switch {
		case e.active == nil:
			err = ErrExchangeMissing
		case e.active.snapshot.RemotePeerID != remote:
			err = ErrExchangePeerMismatch
		case e.active.snapshot.Status != StatusBothConfirmed || (!e.active.offering && e.active.result == nil):
			err = ErrExchangeUnconfirmed
		default:
			if e.active.result != nil {
				result = e.active.result.CloneVT()
			}
		}
	})
	return result, err
}

// StartOnStream runs enrollment on an already authenticated duplex connection.
// The caller retains the stream's transport through completion.
func (e *Engine) StartOnStream(ctx context.Context, stream io.ReadWriteCloser, remote peer.ID, offering bool) {
	attemptCtx, active := e.begin(offering, "", remote, StatusPeerConnected)
	stop := context.AfterFunc(ctx, active.cancel)
	defer stop()
	e.runStream(attemptCtx, active, stream, remote, nil)
}

// StartDirect uses the same approval and enrollment on a manually signaled link.
func (e *Engine) StartDirect(lnk link.Link, offering bool) {
	ctx, active := e.begin(offering, "", lnk.GetRemotePeer(), StatusPeerConnected)
	go e.runDirect(ctx, active, lnk)
}
