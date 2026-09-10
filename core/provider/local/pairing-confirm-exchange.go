package provider_local

import (
	"context"
	"io"
	"time"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/link"
	link_solicit "github.com/s4wave/spacewave/net/link/solicit"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/s4wave/spacewave/net/protocol"
	"github.com/s4wave/spacewave/net/stream"
	stream_packet "github.com/s4wave/spacewave/net/stream/packet"
	"github.com/sirupsen/logrus"
)

// ConfirmProtocolID is the protocol ID for pairing confirmation exchange.
const ConfirmProtocolID = protocol.ID("alpha/account-pairing/1")

// confirmationTimeout is the maximum time to wait for remote confirmation.
const confirmationTimeout = 120 * time.Second

// directConfirmPreamble wakes the accepting side before packetized confirmation starts.
const directConfirmPreamble = byte(0)

// runPairingConfirmExchange runs the mutual SAS confirmation exchange over
// a solicit stream on the session transport's child bus. Both the generator
// and joiner call this: the generator starts it immediately (solicit waits
// for the remote peer to connect), the joiner starts it after the bifrost
// link establishes.
//
// The flow:
// 1. Emit SolicitProtocol directive to match with the remote peer
// 2. On stream match: extract remote peer ID, compute SAS emoji
// 3. Set VERIFYING_EMOJI status with emoji data
// 4. Wait for local user confirmation via confirmCh
// 5. Send local confirmation to remote
// 6. Wait for remote's confirmation
// 7. Update pairing status based on both results
func (a *ProviderAccount) runPairingConfirmExchange(ctx context.Context) {
	st := a.GetSessionTransport()
	if st == nil {
		return
	}
	childBus := st.GetChildBus()
	if childBus == nil {
		return
	}

	localPeerID := st.GetPeerID()
	le := a.le.WithField("phase", "confirm-exchange")

	// Use a solicit protocol to establish a bilateral stream.
	dir := link_solicit.NewSolicitProtocol(
		ConfirmProtocolID,
		[]byte("pairing-confirm"),
		"",
		0,
	)

	streamCh := make(chan link_solicit.SolicitMountedStream, 1)
	_, solicitRef, err := childBus.AddDirective(
		dir,
		directive.NewTypedCallbackHandler(
			func(v directive.TypedAttachedValue[link_solicit.SolicitMountedStream]) {
				select {
				case streamCh <- v.GetValue():
				default:
				}
			},
			nil, nil, nil,
		),
	)
	if err != nil {
		le.WithError(err).Warn("failed to add confirm solicit directive")
		return
	}
	defer solicitRef.Release()

	// Wait for a bilateral stream match or context cancel.
	var sms link_solicit.SolicitMountedStream
	select {
	case <-ctx.Done():
		return
	case sms = <-streamCh:
	}

	ms, taken, err := sms.AcceptMountedStream()
	if err != nil || taken {
		return
	}

	strm := ms.GetStream()
	defer strm.Close()

	// Extract remote peer from the matched stream.
	remotePeerID := ms.GetPeerID()
	if len(remotePeerID) == 0 {
		le.Warn("matched stream has no remote peer ID")
		return
	}

	// The offering client learns its peer from solicitation. Retain that link
	// through approval and enrollment so the transport's idle lease cannot close it.
	_, releaseLink, err := link.EstablishLinkWithPeerEx(ctx, childBus, localPeerID, remotePeerID, false)
	if err != nil {
		a.SetPairingFailed("failed to retain the pairing connection")
		return
	}
	defer releaseLink()

	le = le.WithField("remote-peer", remotePeerID.String()[:8])
	le.Debug("pairing confirm stream accepted")

	a.runConfirmExchangeOnStream(ctx, strm, remotePeerID, localPeerID, le)
}

// runDirectConfirmExchange opens (or accepts) a stream on the direct bifrost
// link and runs the mutual SAS confirmation exchange over it.
func (a *ProviderAccount) runDirectConfirmExchange(ctx context.Context, lnk link.Link, localPeerID peer.ID, isOfferer bool) {
	le := a.le.WithField("phase", "direct-confirm-exchange")

	// Offerer (QUIC server) accepts streams; answerer (QUIC client) opens.
	var strm stream.Stream
	var err error
	if isOfferer {
		strm, _, err = lnk.AcceptStream()
	} else {
		strm, err = lnk.OpenStream(stream.OpenOpts{})
		if err == nil {
			_, err = strm.Write([]byte{directConfirmPreamble})
		}
	}
	if err != nil {
		le.WithError(err).Warn("failed to open/accept confirm stream")
		a.SetPairingFailed("failed to establish confirmation channel")
		return
	}
	defer strm.Close()
	if isOfferer {
		var buf [1]byte
		if _, err := io.ReadFull(strm, buf[:]); err != nil {
			le.WithError(err).Warn("failed to read direct confirm preamble")
			a.SetPairingFailed("failed to establish confirmation channel")
			return
		}
	}

	remotePeerID := lnk.GetRemotePeer()
	le = le.WithField("remote-peer", remotePeerID.String()[:8])
	le.Debug("direct pairing confirm stream opened")

	a.runConfirmExchangeOnStream(ctx, strm, remotePeerID, localPeerID, le)
}

// runConfirmExchangeOnStream runs the core SAS emoji computation and bilateral
// confirmation exchange over an established stream. Used by both the
// SolicitProtocol path (cloud relay) and the direct link path (no-cloud).
func (a *ProviderAccount) runConfirmExchangeOnStream(
	ctx context.Context,
	strm io.ReadWriteCloser,
	remotePeerID peer.ID,
	localPeerID peer.ID,
	le *logrus.Entry,
) {
	// Capture one exchange identity so an old connection cannot complete a new attempt.
	var active *pairingState
	var sessionKey crypto.PrivKey
	var offering bool
	a.pairingBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		active = a.pairing
		if active != nil {
			sessionKey = active.sessionKey
			offering = active.offering
		}
	})
	if sessionKey == nil || localPeerID == "" {
		return
	}
	fail := func(status PairingStatus, err error) {
		a.pairingBcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
			if a.pairing == active {
				active.status = status
				active.errMsg = err.Error()
				broadcast()
			}
		})
	}

	// Cancellation closes the stream and releases pending protocol reads and writes.
	exchangeCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stopClose := context.AfterFunc(exchangeCtx, func() { _ = strm.Close() })
	defer stopClose()
	sess := stream_packet.NewSession(strm, 16<<20)
	prepareCtx, cancelPrepare := context.WithTimeout(exchangeCtx, confirmationTimeout)
	stopPrepare := context.AfterFunc(prepareCtx, func() { _ = strm.Close() })
	enrollment, err := a.preparePairingEnrollment(prepareCtx, sess, offering, localPeerID, remotePeerID)
	stopPrepare()
	cancelPrepare()
	if err != nil {
		fail(PairingStatusFailed, errors.Wrap(err, "prepare account enrollment"))
		return
	}
	defer enrollment.release()

	// Present the authenticated account and SAS before any grant or binding changes.
	remotePub, err := remotePeerID.ExtractPublicKey()
	if err != nil {
		fail(PairingStatusFailed, err)
		return
	}
	emoji, err := DeriveSASEmoji(sessionKey, remotePub, localPeerID, remotePeerID)
	if err != nil {
		fail(PairingStatusFailed, err)
		return
	}
	confirmCh := make(chan bool, 1)
	var current bool
	a.pairingBcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		if a.pairing == active {
			active.remotePeerID = remotePeerID
			active.accountID = enrollment.offer.GetAccountId()
			active.accountName = enrollment.offer.GetDisplayName()
			active.emoji = emoji
			active.confirmCh = confirmCh
			active.enrolledSession = nil
			active.status = PairingStatusVerifyingEmoji
			current = true
			broadcast()
		}
	})
	if !current {
		return
	}

	// Both approvals cover the account and both receiving identity proofs.
	approvalCtx, cancelApproval := context.WithTimeout(exchangeCtx, confirmationTimeout)
	setStatus := func(status PairingStatus) {
		a.pairingBcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
			if a.pairing == active {
				active.status = status
				broadcast()
			}
		})
	}
	status, err := exchangePairingApproval(approvalCtx, sess, confirmCh, enrollment.proof, setStatus)
	cancelApproval()
	if err != nil {
		fail(status, err)
		return
	}
	setStatus(PairingStatusEnrolling)
	if err := a.completePairingEnrollment(exchangeCtx, sess, enrollment); err != nil {
		le.WithError(err).Warn("account enrollment failed")
		fail(PairingStatusFailed, err)
		return
	}

	// Completion exposes only an account attachment acknowledged by the receiver.
	a.pairingBcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		if a.pairing == active {
			if enrollment.session != nil {
				active.enrolledSession = enrollment.session.GetSessionRef()
			}
			active.status = PairingStatusBothConfirmed
			active.errMsg = ""
			broadcast()
		}
	})
}

// exchangePairingApproval reads the remote decision while the local user decides.
// A remote rejection ends the flow immediately, including before local approval.
func exchangePairingApproval(ctx context.Context, stream *stream_packet.Session, confirmCh <-chan bool, proof string, setStatus func(PairingStatus)) (PairingStatus, error) {
	type decision struct {
		message PairingConfirmMessage
		err     error
	}
	remote := make(chan decision, 1)
	stop := context.AfterFunc(ctx, func() { _ = stream.Close() })
	defer func() {
		stop()
		if ctx.Err() != nil {
			_ = stream.Close()
		}
	}()
	go func() {
		var message PairingConfirmMessage
		err := stream.RecvMsg(&message)
		remote <- decision{message: message, err: err}
	}()

	// Keep the receiving goroutine active before either side sends to a duplex stream.
	var localApproved, remoteApproved bool
	for !localApproved || !remoteApproved {
		select {
		case <-ctx.Done():
			return PairingStatusConfirmationTimeout, errors.New("pairing confirmation timed out")
		case approved := <-confirmCh:
			confirmCh = nil
			if err := stream.SendMsg(&PairingConfirmMessage{Confirmed: approved, Rejected: !approved, OperationContext: proof}); err != nil {
				return PairingStatusFailed, err
			}
			if !approved {
				return PairingStatusPairingRejected, errors.New("pairing rejected locally")
			}
			localApproved = true
			setStatus(PairingStatusWaitingForRemote)
		case result := <-remote:
			remote = nil
			if result.err != nil {
				return PairingStatusFailed, result.err
			}
			if result.message.GetOperationContext() != proof {
				return PairingStatusPairingRejected, errors.New("remote approval selected a different account enrollment")
			}
			if result.message.GetRejected() || !result.message.GetConfirmed() {
				return PairingStatusPairingRejected, errors.New("remote client rejected the pairing")
			}
			remoteApproved = true
		}
	}
	return PairingStatusEnrolling, nil
}
