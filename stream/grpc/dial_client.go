package stream_grpc

import (
	"github.com/aperturerobotics/bifrost/stream/grpc/rpc"
)

// DialStreamClientRPC fulfills stream RPC on the client side.
type DialStreamClientRPC struct {
	StreamService_DialStreamClient
}

// NewDialStreamClientRPC builds a new DialStreamClientRPC.
func NewDialStreamClientRPC(
	client StreamService_DialStreamClient,
) stream_grpc_rpc.RPC {
	return &DialStreamClientRPC{
		StreamService_DialStreamClient: client,
	}
}

// Send sends a packet.
func (r *DialStreamClientRPC) Send(resp *stream_grpc_rpc.Data) error {
	return r.StreamService_DialStreamClient.Send(
		&DialStreamRequest{
			Data: resp,
		},
	)
}

// Recv receives a packet.
func (r *DialStreamClientRPC) Recv() (*stream_grpc_rpc.Data, error) {
	msg, err := r.StreamService_DialStreamClient.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_grpc_rpc.RPC = ((*DialStreamClientRPC)(nil))
