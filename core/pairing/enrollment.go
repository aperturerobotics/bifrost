package pairing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/provider"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/net/peer"
	stream_packet "github.com/s4wave/spacewave/net/stream/packet"
)

// Enrollment retains the selected account and receiving capability until the
// pairing engine releases its preparation. Offering identifies the account
// source, independently of which client created the pairing code.
type Enrollment struct {
	// Choice is the immutable account relationship approved by both clients.
	Choice *AccountChoice
	// Offer binds the selected account to the complete account choice.
	Offer *AccountOffer
	// Identity proves possession of both receiving keys.
	Identity *Identity
	// Receiver installs the approved account on the receiving client.
	Receiver *Receiver
	// Account is the mounted destination account on the receiving client.
	Account provider.ProviderAccount
	// Offering indicates that this client authorizes the selected account.
	Offering bool
	// Release drops preparation handles without undoing durable enrollment.
	Release func()
}

// prepare exchanges account identities, fixes the selected outcome, and reserves
// the receiving keys before either client can approve account access.
func (e *Engine) prepare(ctx context.Context, active *attempt, stream *stream_packet.Session, remote peer.ID) (*Enrollment, error) {
	// Describe the selected local account; Home has only a temporary transport.
	var local *AccountOffer
	if active.offerCurrent {
		var err error
		local, err = e.adapter.OfferPairingAccount(ctx, e.key)
		if err != nil {
			return nil, err
		}
		if local.MachineName == "" {
			local.MachineName, _ = os.Hostname()
			if local.MachineName == "" || local.MachineName == "js" {
				local.MachineName = "Browser"
			}
		}
		if objects, ok := e.adapter.(sobject.SharedObjectProvider); ok {
			list, release, err := objects.AccessSharedObjectList(ctx, nil)
			if err != nil {
				return nil, err
			}
			inventory, err := list.WaitValue(ctx, nil)
			release()
			if err != nil {
				return nil, err
			}
			for _, entry := range inventory.GetSharedObjects() {
				if !entry.GetMeta().GetAccountPrivate() && entry.GetMeta().GetBodyType() == "space" {
					local.SpaceCount++
				}
			}
		}
	}

	// Order writes by connection role so an unbuffered duplex stream cannot deadlock.
	var remoteOffer *AccountOffer
	if active.offering {
		if err := stream.SendMsg(&Frame{Body: &Frame_Account{Account: local}}); err != nil {
			return nil, err
		}
		frame, err := ReceiveFrame(stream)
		if err != nil {
			return nil, err
		}
		remoteOffer = frame.GetAccount()
	} else {
		frame, err := ReceiveFrame(stream)
		if err != nil {
			return nil, err
		}
		remoteOffer = frame.GetAccount()
		if err := stream.SendMsg(&Frame{Body: &Frame_Account{Account: local}}); err != nil {
			return nil, err
		}
	}

	// Both clients retain the same ordering of account identities in their choice.
	choice := &AccountChoice{OfferedAccount: local, ReceivingAccount: remoteOffer}
	if !active.offering {
		choice.OfferedAccount, choice.ReceivingAccount = remoteOffer, local
	}
	if choice.GetReceivingAccount().GetAccountId() == "" {
		choice.ReceivingAccount = nil
	}
	if err := choice.ValidateAccounts(); err != nil {
		return nil, err
	}
	if err := e.chooseAccount(ctx, active, stream, choice); err != nil {
		return nil, err
	}

	// Bind both identities and the outcome into the selected account's key proofs.
	offer := choice.GetOfferedAccount().CloneVT()
	offering := active.offering
	if choice.GetOutcome() == AccountOutcome_AccountOutcome_SIGN_IN_RECEIVING || choice.GetOutcome() == AccountOutcome_AccountOutcome_MERGE_INTO_RECEIVING {
		offer = choice.GetReceivingAccount().CloneVT()
		offering = !offering
	}
	data, err := choice.MarshalVT()
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)
	offer.SelectionContext = hex.EncodeToString(digest[:])
	if offering {
		frame, err := ReceiveFrame(stream)
		if err != nil {
			return nil, err
		}
		identity := frame.GetIdentity()
		if err := ValidateIdentity(offer, identity, e.peerID, remote); err != nil {
			return nil, err
		}
		return &Enrollment{Choice: choice, Offer: offer, Identity: identity, Offering: true, Release: func() {}}, nil
	}

	// The configured provider owns receiving keys, storage, and durable attachment.
	p, releaseProvider, err := provider.ExLookupProvider(ctx, e.b, SessionProviderID(offer), false, nil)
	if err != nil {
		return nil, err
	}
	if p == nil {
		releaseProvider.Release()
		return nil, errors.New("the selected account provider is not configured")
	}
	account, releaseAccount, err := p.AccessProviderAccount(ctx, offer.GetAccountId(), nil)
	if err != nil {
		releaseProvider.Release()
		return nil, err
	}
	release := func() { releaseAccount(); releaseProvider.Release() }
	adapter, ok := account.(AccountAdapter)
	if !ok {
		release()
		return nil, errors.New("the selected account provider does not support pairing")
	}
	receiver, err := adapter.PreparePairingReceiver(ctx, offer, remote, e.peerID)
	if err != nil {
		release()
		return nil, err
	}
	releaseAll := func() { receiver.Release(); release() }
	if err := ValidateIdentity(offer, receiver.Identity, remote, e.peerID); err != nil {
		releaseAll()
		return nil, err
	}
	if err := stream.SendMsg(&Frame{Body: &Frame_Identity{Identity: receiver.Identity}}); err != nil {
		releaseAll()
		return nil, err
	}
	return &Enrollment{Choice: choice, Offer: offer, Identity: receiver.Identity, Receiver: receiver, Account: account, Release: releaseAll}, nil
}

// chooseAccount lets the code-entering client propose one outcome. The other
// client sees that exact proposal before approving and can reject it. A fixed
// proposal cannot change after either client's approval begins.
func (e *Engine) chooseAccount(ctx context.Context, active *attempt, stream *stream_packet.Session, choice *AccountChoice) error {
	// Home and an already-shared account have only one useful sign-in direction.
	choice.Outcome = AccountOutcome_AccountOutcome_SIGN_IN_OFFERED
	if receiving := choice.GetReceivingAccount(); receiving != nil && !SameAccount(choice.GetOfferedAccount(), receiving) {
		choice.Outcome = AccountOutcome_AccountOutcome_UNSPECIFIED
		e.update(active, func(a *attempt) {
			a.snapshot.Choice = choice.CloneVT()
			a.snapshot.Status = StatusSelectingAccount
		})
	}

	// Receive and validate the entire proposal against the authenticated exchange.
	if active.offering {
		frame, err := ReceiveFrame(stream)
		if err != nil {
			return err
		}
		selected := frame.GetChoice()
		choice.Outcome = selected.GetOutcome()
		if !choice.EqualVT(selected) {
			return errors.New("pairing choice does not match the exchanged accounts")
		}
		if err := choice.Validate(); err != nil {
			return err
		}
		e.update(active, func(a *attempt) { a.snapshot.Choice = choice.CloneVT() })
		return nil
	}

	// Wait on the engine's choice submission while retaining the same operation.
	if choice.GetOutcome() == AccountOutcome_AccountOutcome_UNSPECIFIED {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case choice.Outcome = <-active.choose:
		}
	}
	if err := choice.Validate(); err != nil {
		return err
	}
	e.update(active, func(a *attempt) { a.snapshot.Choice = choice.CloneVT() })
	return stream.SendMsg(&Frame{Body: &Frame_Choice{Choice: choice}})
}
