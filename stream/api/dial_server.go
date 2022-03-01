package stream_api

import stream_api_rpc "github.com/aperturerobotics/bifrost/stream/api/rpc"

// DialServerRPC fulfills the RPC on the server side.
type DialServerRPC struct {
	DRPCStreamService_DialStreamStream
}

// NewDialServerRPC builds a new DialServerRPC.
func NewDialServerRPC(
	serv DRPCStreamService_DialStreamStream,
) stream_api_rpc.RPC {
	return &DialServerRPC{
		DRPCStreamService_DialStreamStream: serv,
	}
}

// Send sends a packet.
func (r *DialServerRPC) Send(resp *stream_api_rpc.Data) error {
	return r.DRPCStreamService_DialStreamStream.Send(
		&DialStreamResponse{
			Data: resp,
		},
	)
}

// Recv receives a packet.
func (r *DialServerRPC) Recv() (*stream_api_rpc.Data, error) {
	msg, err := r.DRPCStreamService_DialStreamStream.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_api_rpc.RPC = ((*DialServerRPC)(nil))
