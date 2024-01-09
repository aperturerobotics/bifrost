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
	//
	// This function should return as soon as possible, and start
	// additional goroutines to manage the lifecycle of the stream.
	//
	// The context will be canceled when the Link closes.
	// The context will /not/ be canceled when ms closes.
	// Any returned error indicates the stream should be closed.
	HandleMountedStream(ctx context.Context, ms MountedStream) error
}

// MountedStreamContext is the value attached to a Context containing
// information about the current mounted stream.
//
// Used in several places to pass stream info via Context, for example:
// - stream/srpc/server: attach stream info to RPC context
// - stream/drpc/server: attach stream info to RPC context
type MountedStreamContext = MountedStream

// mountedStreamContextKey is the context key used for WithValue.
type mountedStreamContextKey struct{}

// WithMountedStreamContext attaches a MountedStreamContext to a Context.
func WithMountedStreamContext(ctx context.Context, msc MountedStreamContext) context.Context {
	return context.WithValue(ctx, mountedStreamContextKey{}, msc)
}

// GetMountedStreamContext returns the MountedStreamContext from the Context or nil if unset.
func GetMountedStreamContext(ctx context.Context) MountedStreamContext {
	val := ctx.Value(mountedStreamContextKey{})
	msc, ok := val.(MountedStreamContext)
	if !ok {
		return nil
	}
	return msc
}
