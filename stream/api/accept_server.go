package stream_api

import stream_api_rpc "github.com/aperturerobotics/bifrost/stream/api/rpc"

// AcceptServerRPC fulfills rpc accept streams on the server.
type AcceptServerRPC struct {
	DRPCStreamService_AcceptStreamStream
}

// NewAcceptServerRPC constructs a new AcceptServerRPC.
func NewAcceptServerRPC(
	serv DRPCStreamService_AcceptStreamStream,
) stream_api_rpc.RPC {
	return &AcceptServerRPC{
		DRPCStreamService_AcceptStreamStream: serv,
	}
}

// Send sends a packet.
func (r *AcceptServerRPC) Send(resp *stream_api_rpc.Data) error {
	return r.DRPCStreamService_AcceptStreamStream.Send(
		&AcceptStreamResponse{
			Data: resp,
		},
	)
}

// Recv receives a packet.
func (r *AcceptServerRPC) Recv() (*stream_api_rpc.Data, error) {
	msg, err := r.DRPCStreamService_AcceptStreamStream.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_api_rpc.RPC = ((*AcceptServerRPC)(nil))
