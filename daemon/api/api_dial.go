//+build !js

package api

import (
	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/bifrost/stream/grpc/dial"
)

// dialRPC fulfills grpc RPC
type dialRPC struct {
	BifrostDaemonService_DialStreamServer
}

// Send sends a packet.
func (r *dialRPC) Send(resp *stream_grpc.Response) error {
	return r.BifrostDaemonService_DialStreamServer.Send(&DialStreamResponse{
		Response: resp,
	})
}

// Recv receives a packet.
func (r *dialRPC) Recv() (*stream_grpc.Request, error) {
	msg, err := r.BifrostDaemonService_DialStreamServer.Recv()
	return msg.GetRequest(), err
}

// DialStream dials a outgoing stream.
// Stream data is sent over the request / response streams.
func (a *API) DialStream(serv BifrostDaemonService_DialStreamServer) error {
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
		&dialRPC{BifrostDaemonService_DialStreamServer: serv},
	)
}

// _ is a type assertion
var _ stream_grpc.RPC = ((*dialRPC)(nil))
