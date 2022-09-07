package identity

import (
	"context"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// Result is the outcome of the handshake.
type Result struct {
	// Secret is the negotiated secret.
	Secret [32]byte
	// Peer is the public key of the remote peer.
	Peer crypto.PubKey
	// ExtraData is the extra data the remote peer sent.
	ExtraData []byte
}

// Handshaker performs an identity handshake.
type Handshaker interface {
	// Execute executes the handshake with a context.
	// Returning an error cancels the attempt.
	Execute(ctx context.Context) (*Result, error)
	// Handle handles an incoming packet.
	// The buffer will be re-used after the func returns.
	// Returns if we expect more handshake packets.
	Handle(data []byte) bool
	// Close cleans up any resources allocated by the handshake.
	Close()
}
