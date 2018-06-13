package handshake

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
)

// Handshaker upgrades a connection with a particular attribute.
type Handshaker interface {
	// Handshake performs the handshake.
	Handshake(context.Context, link.Link) (link link.Link, err error)
}

// HandshakeIdentity performs a key-exchange and secret negotiation handshake.
// Returns the generated secret and any error.
// HandshakeIdentity(ctx context.Context) ([]byte, error)
