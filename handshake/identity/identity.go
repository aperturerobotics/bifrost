package identity

import (
	"context"
	"github.com/libp2p/go-libp2p-crypto"
)

// Result is the outcome of the handshake.
type Result struct {
	// Secret is the negotiated secret.
	Secret [32]byte
	// Peer is the public key of the remote peer.
	Peer crypto.PubKey
}

// Handshaker performs an identity handshake.
type Handshaker interface {
	// Execute executes the handshake with a context.
	// Initiator indicates the handshaker is the initiator of the handshake.
	// Returning an error cancels the attempt.
	Execute(ctx context.Context, initiator bool) (*Result, error)
	// Handle handles an incoming packet.
	// The buffer will be re-used after the func returns.
	// Returns if we expect more handshake packets.
	Handle(data []byte) bool
	// Close cleans up any resources allocated by the handshake.
	Close()
}
