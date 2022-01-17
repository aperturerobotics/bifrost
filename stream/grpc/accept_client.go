package stream_grpc

import (
	stream_grpc_rpc "github.com/aperturerobotics/bifrost/stream/grpc/rpc"
)

// AcceptStreamClientRPC fulfills grpc RPC on the client side.
type AcceptStreamClientRPC struct {
	StreamService_AcceptStreamClient
}

// NewAcceptStreamClientRPC builds a new AcceptStreamClient.
func NewAcceptStreamClientRPC(
	client StreamService_AcceptStreamClient,
) stream_grpc_rpc.RPC {
	return &AcceptStreamClientRPC{
		StreamService_AcceptStreamClient: client,
	}
}

// Send sends a packet.
func (r *AcceptStreamClientRPC) Send(resp *stream_grpc_rpc.Data) error {
	return r.StreamService_AcceptStreamClient.Send(
		&AcceptStreamRequest{
			Data: resp,
		},
	)
}

// Recv receives a packet.
func (r *AcceptStreamClientRPC) Recv() (*stream_grpc_rpc.Data, error) {
	msg, err := r.StreamService_AcceptStreamClient.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_grpc_rpc.RPC = ((*AcceptStreamClientRPC)(nil))
