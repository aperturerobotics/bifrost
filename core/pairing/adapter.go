package pairing

import (
	"context"

	"github.com/s4wave/spacewave/core/session"
	"github.com/s4wave/spacewave/core/transport"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/peer"
	stream_packet "github.com/s4wave/spacewave/net/stream/packet"
)

// AccountAdapter implements account enrollment through the responsible provider.
// Preparation may reserve local receiving keys; it cannot grant account access.
type AccountAdapter interface {
	OfferPairingAccount(context.Context, crypto.PrivKey) (*AccountOffer, error)
	PreparePairingReceiver(context.Context, *AccountOffer, peer.ID, peer.ID) (*Receiver, error)
	EnrollPairingReceiver(context.Context, *stream_packet.Session, *AccountOffer, *Identity, crypto.PrivKey, peer.ID, peer.ID) error
}

// Receiver retains the approved receiving identity through durable attachment.
// Receive consumes provider checkpoints through a final Complete frame. Release
// drops preparation handles while preserving keys and any completed attachment.
type Receiver struct {
	// Transport returns the enrolled Session connection owner after Receive succeeds.
	Transport func(context.Context) (*transport.SessionTransport, error)
	Identity  *Identity
	Receive   func(context.Context, *stream_packet.Session) error
	Release   func()
}

// Session supports one pairing operation within its mounted lifetime.
type Session interface {
	GetPairingEngine() (*Engine, error)
	GetPairingTransport(context.Context, Relay) (*transport.SessionTransport, error)
}

// SessionProviderID returns the configured provider selected by an offer.
// Empty identifies local offers from the first account-enrollment protocol.
func SessionProviderID(offer *AccountOffer) string {
	if id := offer.GetProviderId(); id != "" {
		return id
	}
	return "local"
}

// receivingSessionRef names the Session attached by a successful receiver.
func receivingSessionRef(receiver *Receiver) *session.SessionRef {
	if receiver == nil {
		return nil
	}
	return receiver.Identity.GetSessionRef()
}
