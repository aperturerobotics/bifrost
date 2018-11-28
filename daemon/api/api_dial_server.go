package api

import (
	"github.com/aperturerobotics/bifrost/stream/grpc"
)

// DialRPCServer fulfills grpc RPC on the server side.
type DialRPCServer struct {
	BifrostDaemonService_DialStreamServer
}

// NewDialRPCServer builds a new DialRPCServer.
func NewDialRPCServer(serv BifrostDaemonService_DialStreamServer) stream_grpc.RPC {
	return &DialRPCServer{
		BifrostDaemonService_DialStreamServer: serv,
	}
}

// Send sends a packet.
func (r *DialRPCServer) Send(resp *stream_grpc.Data) error {
	return r.BifrostDaemonService_DialStreamServer.Send(&DialStreamResponse{
		Data: resp,
	})
}

// Recv receives a packet.
func (r *DialRPCServer) Recv() (*stream_grpc.Data, error) {
	msg, err := r.BifrostDaemonService_DialStreamServer.Recv()
	return msg.GetData(), err
}

// _ is a type assertion
var _ stream_grpc.RPC = ((*DialRPCServer)(nil))
