package bifrost_api

import (
	stream_api "github.com/aperturerobotics/bifrost/stream/api"
	stream_api_accept "github.com/aperturerobotics/bifrost/stream/api/accept"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
)

// AcceptStream accepts an incoming stream.
// Stream data is sent over the request / response streams.
func (a *API) AcceptStream(serv stream_api.DRPCStreamService_AcceptStreamStream) error {
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

	// wait until it's ready
	val, _, valRef, err := loader.WaitExecControllerRunning(ctx, a.bus, dir, nil)
	if err != nil {
		return err
	}
	defer valRef.Release()

	ctrl := val.(*stream_api_accept.Controller)
	return ctrl.AttachRPC(stream_api.NewAcceptServerRPC(serv))
}
