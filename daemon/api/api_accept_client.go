package api

import (
	"github.com/aperturerobotics/bifrost/stream/grpc"
)

// AcceptRPCClient fulfills grpc RPC on the client side.
type AcceptRPCClient struct {
	BifrostDaemonService_AcceptStreamClient
}

// NewAcceptRPCClient builds a new AcceptRPCClient.
func NewAcceptRPCClient(client BifrostDaemonService_AcceptStreamClient) stream_grpc.RPC {
	return &AcceptRPCClient{
		BifrostDaemonService_AcceptStreamClient: client,
	}
}

// Send sends a packet.
func (r *AcceptRPCClient) Send(resp *stream_grpc.Data) error {
	return r.BifrostDaemonService_AcceptStreamClient.Send(&AcceptStreamRequest{
		Data: resp,
	})
}

// Recv receives a packet.
func (r *AcceptRPCClient) Recv() (*stream_grpc.Data, error) {
	msg, err := r.BifrostDaemonService_AcceptStreamClient.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_grpc.RPC = ((*AcceptRPCClient)(nil))
