package stream_grpc

import (
	"github.com/aperturerobotics/bifrost/stream/grpc/rpc"
)

// AcceptServerRPC fulfills grpc accept streams on the server.
type AcceptServerRPC struct {
	StreamService_AcceptStreamServer
}

// NewAcceptServerRPC constructs a new AcceptServerRPC.
func NewAcceptServerRPC(
	serv StreamService_AcceptStreamServer,
) stream_grpc_rpc.RPC {
	return &AcceptServerRPC{
		StreamService_AcceptStreamServer: serv,
	}
}

// Send sends a packet.
func (r *AcceptServerRPC) Send(resp *stream_grpc_rpc.Data) error {
	return r.StreamService_AcceptStreamServer.Send(
		&AcceptStreamResponse{
			Data: resp,
		},
	)
}

// Recv receives a packet.
func (r *AcceptServerRPC) Recv() (*stream_grpc_rpc.Data, error) {
	msg, err := r.StreamService_AcceptStreamServer.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_grpc_rpc.RPC = ((*AcceptServerRPC)(nil))
