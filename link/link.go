package link

import (
	"io"

	"github.com/aperturerobotics/bifrost/peer"
)

// Link represents a one-hop connection between two peers.
type Link interface {
	// GetSecureConn returns a secure ReadWriteCloser for the connection.
	// This connection is expected to be ordered and encrypted.
	GetSecureConn() io.ReadWriteCloser
	// GetInsecureConn returns an insecure ReadWriteCloser for the connection.
	// This allows messages to be sent out-of-band unencrypted.
	GetInsecureConn() io.ReadWriteCloser
	// GetRemotePeer returns the identity of the remote peer.
	GetRemotePeer() peer.ID
	// Close closes the connection.
	// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
	Close() error
}
