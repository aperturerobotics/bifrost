package identity

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/link/kcp"
	"github.com/libp2p/go-libp2p-crypto"
)

// IdentityHandshaker performs an identity handshake.
type IdentityHandshaker interface {
	// HandshakeIdentity performs a key-exchange and secret negotiation handshake.
	// Returns the generated secret and any error.
	HandshakeIdentity(ctx context.Context, lnk link.Link) ([]byte, crypto.PubKey, error)
}

// Handshaker uses an identity handshake to upgrade a link.
type Handshaker struct {
	ih IdentityHandshaker
}

// NewHandshaker builds a new handshaker.
func NewHandshaker(ih IdentityHandshaker) *Handshaker {
	return &Handshaker{ih: ih}
}

// Handshake performs the handshake.
func (h *Handshaker) Handshake(ctx context.Context, lnk link.Link) (link.Link, error) {
	secret, pubKey, err := h.ih.HandshakeIdentity(ctx, lnk)
	if err != nil {
		return nil, err
	}

	return kcp.NewLink(lnk, secret, pubKey)
}
