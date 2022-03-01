package stream_api

import stream_api_rpc "github.com/aperturerobotics/bifrost/stream/api/rpc"

// DialStreamClientRPC fulfills stream RPC on the client side.
type DialStreamClientRPC struct {
	DRPCStreamService_DialStreamClient
}

// NewDialStreamClientRPC builds a new DialStreamClientRPC.
func NewDialStreamClientRPC(
	client DRPCStreamService_DialStreamClient,
) stream_api_rpc.RPC {
	return &DialStreamClientRPC{
		DRPCStreamService_DialStreamClient: client,
	}
}

// Send sends a packet.
func (r *DialStreamClientRPC) Send(resp *stream_api_rpc.Data) error {
	return r.DRPCStreamService_DialStreamClient.Send(
		&DialStreamRequest{
			Data: resp,
		},
	)
}

// Recv receives a packet.
func (r *DialStreamClientRPC) Recv() (*stream_api_rpc.Data, error) {
	msg, err := r.DRPCStreamService_DialStreamClient.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_api_rpc.RPC = ((*DialStreamClientRPC)(nil))
