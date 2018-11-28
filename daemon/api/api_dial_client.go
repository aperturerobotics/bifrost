package api

import (
	"github.com/aperturerobotics/bifrost/stream/grpc"
)

// DialRPCClient fulfills grpc RPC on the client side.
type DialRPCClient struct {
	BifrostDaemonService_DialStreamClient
}

// NewDialRPCClient builds a new DialRPCClient.
func NewDialRPCClient(client BifrostDaemonService_DialStreamClient) stream_grpc.RPC {
	return &DialRPCClient{
		BifrostDaemonService_DialStreamClient: client,
	}
}

// Send sends a packet.
func (r *DialRPCClient) Send(resp *stream_grpc.Data) error {
	return r.BifrostDaemonService_DialStreamClient.Send(&DialStreamRequest{
		Data: resp,
	})
}

// Recv receives a packet.
func (r *DialRPCClient) Recv() (*stream_grpc.Data, error) {
	msg, err := r.BifrostDaemonService_DialStreamClient.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_grpc.RPC = ((*DialRPCClient)(nil))
