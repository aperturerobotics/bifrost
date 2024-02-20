package stream_srpc

import (
	"context"
	"errors"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/starpc/srpc"
	"github.com/sirupsen/logrus"
)

// NewOpenStreamFunc constructs a new OpenStreamFunc which establishes a
// connection with the given peer on-demand when a RPC call is made.
//
// transportID and srcPeer are optional
// starts a read pump in a goroutine
func NewOpenStreamFunc(
	b bus.Bus,
	protocolID protocol.ID,
	srcPeer, destPeer peer.ID,
	transportID uint64,
) srpc.OpenStreamFunc {
	return func(
		ctx context.Context,
		msgHandler srpc.PacketDataHandler,
		closeHandler srpc.CloseHandler,
	) (srpc.PacketWriter, error) {
		return EstablishSrpcStream(
			ctx,
			b,
			protocolID,
			srcPeer, destPeer,
			transportID,
			msgHandler, closeHandler,
		)
	}
}

// NewMultiOpenStreamFunc builds a func which attempts multiple peers.
//
// if timeoutDur <= 0, uses no timeout.
func NewMultiOpenStreamFunc(
	b bus.Bus,
	le *logrus.Entry,
	protocolID protocol.ID,
	srcPeer peer.ID, destPeers []peer.ID,
	transportID uint64,
	timeoutDur time.Duration,
) srpc.OpenStreamFunc {
	return func(
		ctx context.Context,
		msgHandler srpc.PacketDataHandler,
		closeHandler srpc.CloseHandler,
	) (srpc.PacketWriter, error) {
		var lastErr error
		for _, destPeer := range destPeers {
			var estCtx context.Context
			var estCtxCancel context.CancelFunc
			if timeoutDur > 0 {
				estCtx, estCtxCancel = context.WithTimeout(ctx, timeoutDur)
			} else {
				estCtx, estCtxCancel = context.WithCancel(ctx)
			}

			le := le.WithField("server-peer-id", destPeer.String())
			writer, err := EstablishSrpcStream(
				estCtx,
				b,
				protocolID,
				srcPeer, destPeer,
				transportID,
				msgHandler,
				closeHandler,
			)
			estCtxCancel()
			if err != nil {
				le.WithError(err).Warn("unable to establish srpc conn")
				lastErr = err
				continue
			}
			return writer, nil
		}

		if lastErr == nil {
			lastErr = errors.New("connection failed")
		}

		return nil, lastErr
	}
}

// EstablishSrpcStream establishes a srpc stream via a Bifrost stream.
//
// transportID and srcPeer are optional
// starts a read pump in a goroutine
func EstablishSrpcStream(
	ctx context.Context,
	b bus.Bus,
	protocolID protocol.ID,
	srcPeer, destPeer peer.ID,
	transportID uint64,
	msgHandler srpc.PacketDataHandler,
	closeHandler srpc.CloseHandler,
) (srpc.PacketWriter, error) {
	// TODO: EstablishLinkWithPeer via transport id?
	_, lkRel, err := b.AddDirective(
		link.NewEstablishLinkWithPeer(srcPeer, destPeer),
		nil,
	)
	if err != nil {
		return nil, err
	}

	ms, msRel, err := link.OpenStreamWithPeerEx(
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
		return nil, err
	}

	rw := srpc.NewPacketReadWriter(ms.GetStream())
	go func() {
		rw.ReadPump(msgHandler, closeHandler)
		msRel()
		lkRel.Release()
	}()

	return rw, nil
}
