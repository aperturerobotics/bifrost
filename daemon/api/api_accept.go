//+build !js

package api

import (
	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/bifrost/stream/grpc/accept"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
)

// acceptRPC fulfills grpcaccept
type acceptRPC struct {
	BifrostDaemonService_AcceptStreamServer
}

// Send sends a packet.
func (r *acceptRPC) Send(resp *stream_grpc.Response) error {
	return r.BifrostDaemonService_AcceptStreamServer.Send(&AcceptStreamResponse{
		Response: resp,
	})
}

// Recv receives a packet.
func (r *acceptRPC) Recv() (*stream_grpc.Request, error) {
	msg, err := r.BifrostDaemonService_AcceptStreamServer.Recv()
	return msg.GetRequest(), err
}

// AcceptStream accepts an incoming stream.
// Stream data is sent over the request / response streams.
func (a *API) AcceptStream(serv BifrostDaemonService_AcceptStreamServer) error {
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
	return ctrl.AttachRPC(&acceptRPC{BifrostDaemonService_AcceptStreamServer: serv})
}

// _ is a type assertion
var _ stream_grpc.RPC = ((*acceptRPC)(nil))
