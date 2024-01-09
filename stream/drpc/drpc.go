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
func NewDrpcConn(ctx context.Context, strm stream.Stream, opts *DrpcOpts) (*drpcconn.Conn, error) {
	opt, err := opts.BuildOpts()
	if err != nil {
		return nil, err
	}
	return drpcconn.NewWithOptions(strm, opt), nil
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
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	// TODO: EstablishLinkWithPeer via transport id?
	_, lkRel, err := b.AddDirective(
		link.NewEstablishLinkWithPeer(srcPeer, destPeer),
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	ms, msRef, err := link.OpenStreamWithPeerEx(
		ctx,
		b,
		protocolID,
		srcPeer, destPeer, transportID,
		stream.OpenOpts{
			Reliable:  true,
			Encrypted: true,
		},
	)
	if err != nil {
		lkRel.Release()
		return nil, nil, err
	}

	rel := func() {
		msRef()
		lkRel.Release()
	}

	conn, err := NewDrpcConn(ctx, ms.GetStream(), opts)
	if err != nil {
		rel()
		return nil, nil, err
	}

	return conn, rel, err
}
