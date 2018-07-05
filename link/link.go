package link

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/stream"
)

// Link represents a one-hop connection between two peers.
type Link interface {
	// GetUUID returns the host-unique ID.
	// This should be repeatable between re-constructions of the same link.
	GetUUID() uint64
	// GetTransportUUID returns the unique ID of the transport.
	GetTransportUUID() uint64
	// OpenStream opens a stream on the link, with the given parameters.
	OpenStream(opts stream.OpenOpts) (stream.Stream, error)
	// GetRemotePeer returns the identity of the remote peer.
	GetRemotePeer() peer.ID
	// Close closes the link.
	// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
	Close() error
}
