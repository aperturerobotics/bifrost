package bifrost_api

import (
	stream_grpc "github.com/aperturerobotics/bifrost/stream/grpc"
	stream_grpc_dial "github.com/aperturerobotics/bifrost/stream/grpc/dial"
)

// DialStream dials a outgoing stream.
// Stream data is sent over the request / response streams.
func (a *API) DialStream(serv stream_grpc.StreamService_DialStreamServer) error {
	ctx := serv.Context()
	msg, err := serv.Recv()
	if err != nil {
		return err
	}

	conf := msg.GetConfig()
	if err := conf.Validate(); err != nil {
		return err
	}

	return stream_grpc_dial.ProcessRPC(
		ctx,
		a.bus,
		conf,
		stream_grpc.NewDialServerRPC(serv),
	)
}
