package stream_drpc

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/controllerbus/bus"
	"storj.io/drpc/drpcconn"
)

// NewDrpcConn constructs a new dprc conn from a stream.
func NewDrpcConn(ctx context.Context, strm stream.Stream, opts *DrpcOpts) *drpcconn.Conn {
	return drpcconn.NewWithOptions(strm, opts.BuildOpts())
}

// EstablishDrpcConn establishes a drpc connection with a peer.
//
// srcPeer, transportID can be empty.
func EstablishDrpcConn(
	ctx context.Context,
	b bus.Bus,
	opts *DrpcOpts,
	protocolID protocol.ID,
	srcPeer, destPeer peer.ID,
	transportID uint64,
) (*drpcconn.Conn, func(), error) {
	ms, msRef, err := link.OpenStreamWithPeerEx(
		ctx,
		b,
		protocolID,
		srcPeer, destPeer, 0,
		stream.OpenOpts{
			Reliable:  true,
			Encrypted: true,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	conn := NewDrpcConn(ctx, ms.GetStream(), opts)
	return conn, msRef, nil
}
