package bifrost_api

import (
	stream_grpc "github.com/aperturerobotics/bifrost/stream/grpc"
	stream_grpc_accept "github.com/aperturerobotics/bifrost/stream/grpc/accept"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
)

// AcceptStream accepts an incoming stream.
// Stream data is sent over the request / response streams.
func (a *API) AcceptStream(serv stream_grpc.StreamService_AcceptStreamServer) error {
	ctx := serv.Context()
	msg, err := serv.Recv()
	if err != nil {
		return err
	}

	conf := msg.GetConfig()
	if err := conf.Validate(); err != nil {
		return err
	}

	dir := resolver.NewLoadControllerWithConfig(conf)

	// executeController will execute the grpcaccept controller
	// wait until it's ready
	val, _, valRef, err := loader.WaitExecControllerRunning(ctx, a.bus, dir, nil)
	if err != nil {
		return err
	}
	defer valRef.Release()

	ctrl := val.(*stream_grpc_accept.Controller)
	return ctrl.AttachRPC(stream_grpc.NewAcceptServerRPC(serv))
}
