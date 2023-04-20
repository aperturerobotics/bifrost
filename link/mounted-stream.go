package link

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
)

// MountedStream is a stream attached to a Link. This is produced and managed by
// the link controller. A mounted stream is produced after the initial stream
// negotiation is completed.
type MountedStream interface {
	// GetStream returns the underlying stream object.
	GetStream() stream.Stream
	// GetProtocolID returns the protocol ID of the stream.
	GetProtocolID() protocol.ID
	// GetOpenOpts returns the options used to open the stream.
	GetOpenOpts() stream.OpenOpts
	// GetPeerID returns the peer ID for the other end of the stream.
	GetPeerID() peer.ID
	// GetLink returns the associated link carrying the stream.
	GetLink() Link
}

// MountedStreamHandler handles an incoming mounted stream.
type MountedStreamHandler interface {
	// HandleMountedStream handles an incoming mounted stream.
	// Any returned error indicates the stream should be closed.
	// This function should return as soon as possible, and start
	// additional goroutines to manage the lifecycle of the stream.
	// Typically EstablishLink is asserted in HandleMountedStream.
	HandleMountedStream(context.Context, MountedStream) error
}
