package api_controller

import (
	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/stream/grpc/accept"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
)

// AcceptStream accepts an incoming stream.
// Stream data is sent over the request / response streams.
func (a *API) AcceptStream(serv api.BifrostDaemonService_AcceptStreamServer) error {
	ctx := serv.Context()
	msg, err := serv.Recv()
	if err != nil {
		return err
	}

	conf := msg.GetConfig()
	if err := conf.Validate(); err != nil {
		return err
	}

	dir := resolver.NewLoadControllerWithConfigSingleton(conf)

	// executeController will execute the grpcaccept controller
	// wait until it's ready
	val, valRef, err := bus.ExecOneOff(ctx, a.bus, dir, nil)
	if err != nil {
		return err
	}
	defer valRef.Release()

	ctrl := val.(*stream_grpc_accept.Controller)
	return ctrl.AttachRPC(api.NewAcceptRPCServer(serv))
}
