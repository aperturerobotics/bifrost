package pairing

import (
	"context"
	"io"
	"time"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/provider"
	"github.com/s4wave/spacewave/core/transport"
	"github.com/s4wave/spacewave/net/link"
	link_solicit "github.com/s4wave/spacewave/net/link/solicit"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/s4wave/spacewave/net/protocol"
	"github.com/s4wave/spacewave/net/stream"
	stream_packet "github.com/s4wave/spacewave/net/stream/packet"
)

// ProtocolID identifies account enrollment on an authenticated connection.
const ProtocolID = protocol.ID("alpha/account-pairing/1")

const confirmationTimeout = 120 * time.Second

func (e *Engine) runSolicit(ctx context.Context, active *attempt, transport *transport.SessionTransport) {
	childBus := transport.GetChildBus()
	streams := make(chan link_solicit.SolicitMountedStream, 1)
	_, ref, err := childBus.AddDirective(
		link_solicit.NewSolicitProtocol(ProtocolID, []byte("pairing-confirm"), "", 0),
		directive.NewTypedCallbackHandler(func(value directive.TypedAttachedValue[link_solicit.SolicitMountedStream]) {
			select {
			case streams <- value.GetValue():
			default:
			}
		}, nil, nil, nil),
	)
	if err != nil {
		e.fail(active, StatusFailed, err)
		return
	}
	defer ref.Release()

	var mounted link_solicit.SolicitMountedStream
	select {
	case <-ctx.Done():
		return
	case mounted = <-streams:
	}
	accepted, taken, err := mounted.AcceptMountedStream()
	if err != nil || taken {
		if err == nil {
			err = errors.New("pairing stream was already accepted")
		}
		e.fail(active, StatusFailed, err)
		return
	}
	strm := accepted.GetStream()
	defer strm.Close()
	remote := accepted.GetPeerID()
	_, release, err := link.EstablishLinkWithPeerEx(ctx, childBus, e.peerID, remote, false)
	if err != nil {
		e.fail(active, StatusFailed, err)
		return
	}
	defer release()
	e.runStream(ctx, active, strm, remote, nil)
}

func (e *Engine) runDirect(ctx context.Context, active *attempt, lnk link.Link) {
	var strm stream.Stream
	var err error
	if active.offering {
		strm, _, err = lnk.AcceptStream()
	} else {
		strm, err = lnk.OpenStream(stream.OpenOpts{})
		if err == nil {
			_, err = strm.Write([]byte{0})
		}
	}
	if err != nil {
		e.fail(active, StatusFailed, err)
		return
	}
	defer strm.Close()
	if active.offering {
		var preamble [1]byte
		if _, err := io.ReadFull(strm, preamble[:]); err != nil {
			e.fail(active, StatusFailed, err)
			return
		}
		if preamble[0] != 0 {
			e.fail(active, StatusFailed, errors.New("invalid direct pairing preamble"))
			return
		}
	}
	e.runStream(ctx, active, strm, lnk.GetRemotePeer(), lnk)
	snapshot, _ := e.Snapshot()
	if snapshot.Status != StatusBothConfirmed {
		_ = lnk.Close()
	}
}

func (e *Engine) prepare(ctx context.Context, active *attempt, stream *stream_packet.Session, remote peer.ID) (*AccountOffer, *Identity, *Receiver, func(), error) {
	if active.offering {
		offer, err := e.adapter.OfferPairingAccount(ctx, e.key)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if err := stream.SendMsg(&Frame{Body: &Frame_Account{Account: offer}}); err != nil {
			return nil, nil, nil, nil, err
		}
		frame, err := ReceiveFrame(stream)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		identity := frame.GetIdentity()
		return offer, identity, nil, func() {}, ValidateIdentity(offer, identity, e.peerID, remote)
	}

	frame, err := ReceiveFrame(stream)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	offer := frame.GetAccount()
	if offer.GetAccountId() == "" || offer.GetOperationId() == "" {
		return nil, nil, nil, nil, errors.New("pairing source did not offer an account")
	}
	p, releaseProvider, err := provider.ExLookupProvider(ctx, e.b, SessionProviderID(offer), false, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if p == nil {
		releaseProvider.Release()
		return nil, nil, nil, nil, errors.New("the offered account provider is not configured")
	}
	account, releaseAccount, err := p.AccessProviderAccount(ctx, offer.GetAccountId(), nil)
	if err != nil {
		releaseProvider.Release()
		return nil, nil, nil, nil, err
	}
	release := func() { releaseAccount(); releaseProvider.Release() }
	adapter, ok := account.(AccountAdapter)
	if !ok {
		release()
		return nil, nil, nil, nil, errors.New("the offered account provider does not support pairing")
	}
	receiver, err := adapter.PreparePairingReceiver(ctx, offer, remote, e.peerID)
	if err != nil {
		release()
		return nil, nil, nil, nil, err
	}
	releaseAll := func() { receiver.Release(); release() }
	if err := ValidateIdentity(offer, receiver.Identity, remote, e.peerID); err != nil {
		releaseAll()
		return nil, nil, nil, nil, err
	}
	if err := stream.SendMsg(&Frame{Body: &Frame_Identity{Identity: receiver.Identity}}); err != nil {
		releaseAll()
		return nil, nil, nil, nil, err
	}
	return offer, receiver.Identity, receiver, releaseAll, nil
}

func (e *Engine) runStream(ctx context.Context, active *attempt, strm io.ReadWriteCloser, remote peer.ID, directLink link.Link) {
	defer strm.Close()
	stopClose := context.AfterFunc(ctx, func() { _ = strm.Close() })
	defer stopClose()
	sess := stream_packet.NewSession(strm, 16<<20)
	prepareCtx, cancelPrepare := context.WithTimeout(ctx, confirmationTimeout)
	stopPrepare := context.AfterFunc(prepareCtx, func() { _ = strm.Close() })
	offer, identity, receiver, release, err := e.prepare(prepareCtx, active, sess, remote)
	stopPrepare()
	cancelPrepare()
	if err != nil {
		e.fail(active, StatusFailed, errors.Wrap(err, "prepare account enrollment"))
		return
	}
	if !e.retain(active, release) {
		return
	}
	succeeded := false
	defer func() {
		if !succeeded {
			e.releaseResources(active)
		}
	}()
	sourcePeer, receivingPeer := e.peerID, remote
	if !active.offering {
		sourcePeer, receivingPeer = remote, e.peerID
	}
	proof, err := ApprovalContext(offer, identity, sourcePeer, receivingPeer)
	if err != nil {
		e.fail(active, StatusFailed, err)
		return
	}
	remoteKey, err := remote.ExtractPublicKey()
	if err != nil {
		e.fail(active, StatusFailed, err)
		return
	}
	emoji, err := DeriveSASEmoji(e.key, remoteKey, e.peerID, remote)
	if err != nil {
		e.fail(active, StatusFailed, err)
		return
	}
	e.update(active, func(a *attempt) {
		a.snapshot.RemotePeerID = remote
		a.snapshot.AccountID = offer.GetAccountId()
		a.snapshot.AccountName = offer.GetDisplayName()
		a.snapshot.ProviderID = SessionProviderID(offer)
		a.snapshot.Emoji = emoji
		a.snapshot.Status = StatusVerifyingEmoji
	})
	setStatus := func(status Status) { e.update(active, func(a *attempt) { a.snapshot.Status = status }) }
	approvalCtx, cancelApproval := context.WithTimeout(ctx, confirmationTimeout)
	status, err := exchangeApproval(approvalCtx, sess, active.confirm, proof, setStatus)
	cancelApproval()
	if err != nil {
		e.fail(active, status, err)
		return
	}
	setStatus(StatusEnrolling)
	if active.offering {
		err = e.adapter.EnrollPairingReceiver(ctx, sess, offer, identity, e.key, sourcePeer, receivingPeer)
		if err == nil {
			var frame *Frame
			frame, err = ReceiveFrame(sess)
			if err == nil && !frame.GetComplete() {
				err = errors.New("receiving client did not acknowledge account enrollment")
			}
		}
	} else {
		err = receiver.Receive(ctx, sess)
		if err == nil {
			err = sess.SendMsg(&Frame{Body: &Frame_Complete{Complete: true}})
		}
	}
	if err != nil {
		e.fail(active, StatusFailed, err)
		return
	}
	if err := ctx.Err(); err != nil {
		e.fail(active, StatusFailed, err)
		return
	}
	if directLink != nil {
		var owner *transport.SessionTransport
		if active.offering {
			owner, err = e.transport(ctx, Relay{})
		} else if receiver.Transport != nil {
			owner, err = receiver.Transport(ctx)
		} else {
			err = errors.New("receiving provider did not supply a Session transport")
		}
		if err == nil {
			err = owner.AdoptLink(ctx, directLink)
		}
		if err != nil {
			e.fail(active, StatusFailed, errors.Wrap(err, "retain direct pairing connection"))
			return
		}
	}
	e.update(active, func(a *attempt) {
		a.result = receivingSessionRef(receiver)
		a.snapshot.Status = StatusBothConfirmed
		a.snapshot.ErrMsg = ""
	})
	succeeded = true
}

// exchangeApproval reads the remote decision while the local user decides.
// A remote rejection ends the flow immediately, including before local approval.
func exchangeApproval(ctx context.Context, stream *stream_packet.Session, confirmCh <-chan bool, proof string, setStatus func(Status)) (Status, error) {
	type decision struct {
		message Approval
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
		var message Approval
		err := stream.RecvMsg(&message)
		remote <- decision{message: message, err: err}
	}()

	// Keep the receiving goroutine active before either side sends to a duplex stream.
	var localApproved, remoteApproved bool
	for !localApproved || !remoteApproved {
		select {
		case <-ctx.Done():
			return StatusConfirmationTimeout, errors.New("pairing confirmation timed out")
		case approved := <-confirmCh:
			confirmCh = nil
			if err := stream.SendMsg(&Approval{Confirmed: approved, Rejected: !approved, OperationContext: proof}); err != nil {
				return StatusFailed, err
			}
			if !approved {
				return StatusPairingRejected, errors.New("pairing rejected locally")
			}
			localApproved = true
			setStatus(StatusWaitingForRemote)
		case result := <-remote:
			remote = nil
			if result.err != nil {
				return StatusFailed, result.err
			}
			if result.message.GetOperationContext() != proof {
				return StatusPairingRejected, errors.New("remote approval selected a different account enrollment")
			}
			if result.message.GetRejected() || !result.message.GetConfirmed() {
				return StatusPairingRejected, errors.New("remote client rejected the pairing")
			}
			remoteApproved = true
		}
	}
	return StatusEnrolling, nil
}
