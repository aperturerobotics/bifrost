package stream_grpc

import (
	"github.com/aperturerobotics/bifrost/stream/grpc/rpc"
)

// DialServerRPC fulfills grpc RPC on the server side.
type DialServerRPC struct {
	StreamService_DialStreamServer
}

// NewDialServerRPC builds a new DialServerRPC.
func NewDialServerRPC(
	serv StreamService_DialStreamServer,
) stream_grpc_rpc.RPC {
	return &DialServerRPC{
		StreamService_DialStreamServer: serv,
	}
}

// Send sends a packet.
func (r *DialServerRPC) Send(resp *stream_grpc_rpc.Data) error {
	return r.StreamService_DialStreamServer.Send(
		&DialStreamResponse{
			Data: resp,
		},
	)
}

// Recv receives a packet.
func (r *DialServerRPC) Recv() (*stream_grpc_rpc.Data, error) {
	msg, err := r.StreamService_DialStreamServer.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_grpc_rpc.RPC = ((*DialServerRPC)(nil))
