package api

import (
	"github.com/aperturerobotics/bifrost/stream/grpc"
)

// AcceptRPCServer fulfills grpc accept streams on the server.
type AcceptRPCServer struct {
	BifrostDaemonService_AcceptStreamServer
}

// NewAcceptRPCServer constructs a new AcceptRPCServer.
func NewAcceptRPCServer(serv BifrostDaemonService_AcceptStreamServer) stream_grpc.RPC {
	return &AcceptRPCServer{BifrostDaemonService_AcceptStreamServer: serv}
}

// Send sends a packet.
func (r *AcceptRPCServer) Send(resp *stream_grpc.Data) error {
	return r.BifrostDaemonService_AcceptStreamServer.Send(&AcceptStreamResponse{
		Data: resp,
	})
}

// Recv receives a packet.
func (r *AcceptRPCServer) Recv() (*stream_grpc.Data, error) {
	msg, err := r.BifrostDaemonService_AcceptStreamServer.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_grpc.RPC = ((*AcceptRPCServer)(nil))
