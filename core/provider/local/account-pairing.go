package provider_local

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
	"github.com/aperturerobotics/util/routine"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/session"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/link"
	"github.com/s4wave/spacewave/net/peer"
)

// ErrNoSessionTransport is returned when a pairing operation requires
// a session transport but none is running.
var ErrNoSessionTransport = errors.New("no session transport running")

// PairingStatus describes the current phase of a device pairing flow.
// Values match the PairingStatus proto enum in sdk/session/session.proto.
type PairingStatus int32

const (
	PairingStatusIdle                PairingStatus = 0
	PairingStatusCodeGenerated       PairingStatus = 1
	PairingStatusWaitingForPeer      PairingStatus = 2
	PairingStatusPeerConnected       PairingStatus = 3
	PairingStatusVerifyingEmoji      PairingStatus = 4
	PairingStatusVerified            PairingStatus = 5
	PairingStatusFailed              PairingStatus = 6
	PairingStatusSignalingFailed     PairingStatus = 7
	PairingStatusConnectionTimeout   PairingStatus = 8
	PairingStatusWaitingForRemote    PairingStatus = 9
	PairingStatusBothConfirmed       PairingStatus = 10
	PairingStatusPairingRejected     PairingStatus = 11
	PairingStatusConfirmationTimeout PairingStatus = 12
	PairingStatusEnrolling           PairingStatus = 13
)

// PairingSnapshot is a point-in-time snapshot of pairing state.
type PairingSnapshot struct {
	Status       PairingStatus
	Code         string
	RemotePeerID peer.ID
	Emoji        []string
	ErrMsg       string
	AccountID    string
	AccountName  string
	Receiving    bool
}

// pairingState tracks an active pairing flow on the ProviderAccount.
type pairingState struct {
	status       PairingStatus
	code         string
	remotePeerID peer.ID
	sessionKey   crypto.PrivKey
	linkDiRef    directive.Reference
	linkCh       chan link.MountedLink
	emoji        []string
	errMsg       string
	// exchangeRc manages the confirmation exchange goroutine lifecycle.
	exchangeRc *routine.RoutineContainer
	// localConfirmed tracks whether the local user confirmed the SAS match.
	localConfirmed bool
	// confirmCh receives the local user's confirmation decision (true = confirmed, false = rejected).
	confirmCh chan bool
	// direct is true when the active pairing flow uses a direct manual-signal link.
	direct bool
	// offering distinguishes the account source from the receiving client.
	offering bool
	// accountID is the account selected by the authenticated exchange.
	accountID string
	// accountName is the selected account's user-facing name.
	accountName string
	// enrolledSession is set only after the receiving account attachment is durable.
	enrolledSession *session.SessionRef
}

// setPairingContext updates the lifecycle context used for pairing routines.
func (a *ProviderAccount) setPairingContext(ctx context.Context) {
	a.pairingBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		a.pairingCtx = ctx
	})
}

// GetPairingContext returns the lifecycle context used for pairing routines.
func (a *ProviderAccount) GetPairingContext() context.Context {
	var ctx context.Context
	a.pairingBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		ctx = a.pairingCtx
	})
	return ctx
}

// SetPairingCode stores the pairing code and sets status to CODE_GENERATED.
// Starts the confirmation exchange goroutine which waits for a remote peer
// to connect via the solicit protocol (generator side).
func (a *ProviderAccount) SetPairingCode(code string, sessionKey crypto.PrivKey) {
	a.ClearPairingState()
	a.pairingBcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		a.pairing = &pairingState{}
		a.pairing.code = code
		a.pairing.offering = true
		a.pairing.sessionKey = sessionKey
		a.pairing.direct = false
		a.pairing.status = PairingStatusCodeGenerated
		a.startExchangeRoutine()
		bcast()
	})
}

// SetPairingRemotePeer stores the remote peer ID, adds an
// EstablishLinkWithPeer directive on the session transport's child bus,
// and sets status to WAITING_FOR_PEER. sessionKey is stored for SAS
// emoji computation when the link establishes.
func (a *ProviderAccount) SetPairingRemotePeer(ctx context.Context, remotePeerID peer.ID, sessionKey crypto.PrivKey) error {
	// A new operation cancels the prior exchange and its transport demand.
	a.ClearPairingState()

	st := a.GetSessionTransport()
	if st == nil {
		return ErrNoSessionTransport
	}
	if err := st.AwaitReady(ctx); err != nil {
		return err
	}
	childBus := st.GetChildBus()
	if childBus == nil {
		return ErrNoSessionTransport
	}

	// Publish the operation before AddDirective can synchronously deliver a link.
	linkCh := make(chan link.MountedLink, 1)
	active := &pairingState{remotePeerID: remotePeerID, sessionKey: sessionKey, linkCh: linkCh, status: PairingStatusWaitingForPeer}
	a.pairingBcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		a.pairing = active
		bcast()
	})
	handler := directive.NewTypedCallbackHandler(
		func(v directive.TypedAttachedValue[link.MountedLink]) {
			select {
			case linkCh <- v.GetValue():
			default:
			}
			a.onPairingLinkEstablished(active)
		},
		nil, nil, nil,
	)

	// AddDirective outside lock (may block).
	_, diRef, err := childBus.AddDirective(
		link.NewEstablishLinkWithPeer(st.GetPeerID(), remotePeerID),
		handler,
	)
	if err != nil {
		return err
	}

	// Retain the directive only if this operation is still current.
	var retained bool
	a.pairingBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if a.pairing == active {
			active.linkDiRef = diRef
			retained = true
		}
	})
	if !retained {
		diRef.Release()
	}
	return nil
}

// onPairingLinkEstablished is called when the bifrost link with the
// remote peer establishes during pairing. Sets PEER_CONNECTED and
// starts the confirmation exchange which handles emoji computation.
func (a *ProviderAccount) onPairingLinkEstablished(active *pairingState) {
	a.pairingBcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		if a.pairing != active || active.status != PairingStatusWaitingForPeer {
			return
		}
		a.pairing.status = PairingStatusPeerConnected
		a.startExchangeRoutine()
		bcast()
	})
}

// startExchangeRoutine creates and starts the confirmation exchange
// routine container. Must be called inside pairingBcast.HoldLock.
func (a *ProviderAccount) startExchangeRoutine() {
	if a.pairing == nil {
		return
	}
	parentCtx := a.pairingCtx
	if parentCtx == nil {
		a.le.Warn("pairing confirm exchange has no lifecycle context")
		return
	}
	if a.pairing.exchangeRc != nil {
		a.pairing.exchangeRc.ClearContext()
	}
	rc := routine.NewRoutineContainer(
		routine.WithExitCb(func(err error) {
			if err != nil {
				a.le.WithError(err).Warn("pairing confirm exchange exited with error")
			}
		}),
	)
	rc.SetRoutine(func(ctx context.Context) error {
		a.runPairingConfirmExchange(ctx)
		return nil
	})
	a.pairing.exchangeRc = rc
	rc.SetContext(parentCtx, false)
}

// OnDirectPairingConnected is called when a no-cloud WebRTC link establishes
// via ManualSignalTransport. Sets PEER_CONNECTED and starts the confirmation
// exchange directly on the link (bypassing SolicitProtocol).
func (a *ProviderAccount) OnDirectPairingConnected(
	remotePeerID peer.ID,
	sessionKey crypto.PrivKey,
	localPeerID peer.ID,
	lnk link.Link,
	isOfferer bool,
) {
	a.ClearPairingState()
	a.pairingBcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		parentCtx := a.pairingCtx
		a.pairing = &pairingState{}
		a.pairing.remotePeerID = remotePeerID
		a.pairing.sessionKey = sessionKey
		a.pairing.direct = true
		a.pairing.offering = isOfferer
		a.pairing.status = PairingStatusPeerConnected

		if a.pairing.exchangeRc != nil {
			a.pairing.exchangeRc.ClearContext()
		}
		rc := routine.NewRoutineContainer(
			routine.WithExitCb(func(err error) {
				if err != nil {
					a.le.WithError(err).Warn("direct pairing confirm exchange exited with error")
				}
			}),
		)
		rc.SetRoutine(func(ctx context.Context) error {
			a.runDirectConfirmExchange(ctx, lnk, localPeerID, isOfferer)
			return nil
		})
		a.pairing.exchangeRc = rc
		// note: bcast must run on both paths, so the branch cannot return.
		if parentCtx == nil {
			a.le.Warn("direct pairing confirm exchange has no lifecycle context")
		} else {
			rc.SetContext(parentCtx, false)
		}
		bcast()
	})
}

// ConfirmSASMatch sends the local user's SAS confirmation decision.
// Called from the UI when the user clicks "Yes, they match" or "No, abort".
func (a *ProviderAccount) ConfirmSASMatch(confirmed bool) {
	a.pairingBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		if a.pairing == nil || a.pairing.confirmCh == nil {
			return
		}
		a.pairing.localConfirmed = confirmed
		select {
		case a.pairing.confirmCh <- confirmed:
		default:
		}
	})
}

// SetPairingFailed marks the pairing as failed with an error message.
func (a *ProviderAccount) SetPairingFailed(msg string) {
	a.setPairingError(PairingStatusFailed, msg)
}

// SetPairingSignalingFailed marks the pairing as failed due to signaling.
func (a *ProviderAccount) SetPairingSignalingFailed(msg string) {
	a.pairingBcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		if a.pairing == nil || a.pairing.direct {
			return
		}
		a.pairing.status = PairingStatusSignalingFailed
		a.pairing.errMsg = msg
		bcast()
	})
}

// setPairingError sets the pairing to an error status with a message.
func (a *ProviderAccount) setPairingError(status PairingStatus, msg string) {
	a.pairingBcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		if a.pairing == nil {
			return
		}
		a.pairing.status = status
		a.pairing.errMsg = msg
		bcast()
	})
}

// GetPairingBroadcast returns the pairing state broadcast.
func (a *ProviderAccount) GetPairingBroadcast() *broadcast.Broadcast {
	return &a.pairingBcast
}

// GetPairingSnapshot returns a snapshot of the current pairing status.
// Must be called inside pairingBcast.HoldLock or from a single goroutine.
func (a *ProviderAccount) GetPairingSnapshot() PairingSnapshot {
	if a.pairing == nil {
		return PairingSnapshot{Status: PairingStatusIdle}
	}
	return PairingSnapshot{
		Status:       a.pairing.status,
		Code:         a.pairing.code,
		RemotePeerID: a.pairing.remotePeerID,
		Emoji:        a.pairing.emoji,
		ErrMsg:       a.pairing.errMsg,
		AccountID:    a.pairing.accountID,
		AccountName:  a.pairing.accountName,
		Receiving:    !a.pairing.offering,
	}
}

// GetPairingRemotePeerID returns the remote peer ID being paired, or empty.
// Must be called inside pairingBcast.HoldLock or from a single goroutine.
func (a *ProviderAccount) GetPairingRemotePeerID() peer.ID {
	if a.pairing == nil {
		return ""
	}
	return a.pairing.remotePeerID
}

// GetPairingLinkCh returns the channel that receives a link when the
// remote peer connects, or nil if no pairing is active.
func (a *ProviderAccount) GetPairingLinkCh() <-chan link.MountedLink {
	if a.pairing == nil {
		return nil
	}
	return a.pairing.linkCh
}

// ClearPairingState releases any active pairing directive and clears state.
func (a *ProviderAccount) ClearPairingState() {
	// Snapshot refs to release outside lock.
	var diRef directive.Reference
	var exchangeRc *routine.RoutineContainer
	a.pairingBcast.HoldLock(func(bcast func(), _ func() <-chan struct{}) {
		if a.pairing == nil {
			return
		}
		diRef = a.pairing.linkDiRef
		exchangeRc = a.pairing.exchangeRc
		a.pairing = nil
		bcast()
	})
	if exchangeRc != nil {
		exchangeRc.ClearContext()
	}
	if diRef != nil {
		diRef.Release()
	}
}
