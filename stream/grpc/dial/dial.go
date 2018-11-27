package stream_grpc_dial

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/controllerbus/bus"
)

// ProcessRPC processes an RPC by dialing the desired target.
func ProcessRPC(
	ctx context.Context,
	b bus.Bus,
	conf *Config,
	rpc stream_grpc.RPC,
) error {
	if err := conf.Validate(); err != nil {
		return err
	}

	localPeerID, err := conf.ParseLocalPeerID()
	if err != nil {
		return err
	}

	remotePeerID, err := conf.ParsePeerID()
	if err != nil {
		return err
	}

	// Dial the target.
	strm, rel, err := link.OpenStreamWithPeerEx(
		ctx,
		b,
		protocol.ID(conf.GetProtocolId()),
		localPeerID, remotePeerID,
		conf.GetTransportId(),
		stream.OpenOpts{
			Reliable:  conf.GetReliable(),
			Encrypted: conf.GetEncrypted(),
		},
	)
	if err != nil {
		return err
	}
	defer rel()

	defer strm.GetStream().Close()
	return stream_grpc.AttachRPCToStream(rpc, strm.GetStream())
}
