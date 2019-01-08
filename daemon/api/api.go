package api

import (
	"github.com/aperturerobotics/bifrost/peer/grpc"
	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/controllerbus/grpc"
	"google.golang.org/grpc"
)

// BifrostDaemonServer is the bifrost daemon server interface.
type BifrostDaemonServer interface {
	stream_grpc.StreamServiceServer
	peer_grpc.PeerServiceServer
	controllerbus_grpc.ControllerBusServiceServer
}

// RegisterAsGRPCServer registers a server with a grpc server.
func RegisterAsGRPCServer(s BifrostDaemonServer, grpcServer *grpc.Server) {
	stream_grpc.RegisterStreamServiceServer(grpcServer, s)
	peer_grpc.RegisterPeerServiceServer(grpcServer, s)
	controllerbus_grpc.RegisterControllerBusServiceServer(grpcServer, s)
}

// BifrostDaemonClient is the bifrost daemon client interface.
type BifrostDaemonClient interface {
	stream_grpc.StreamServiceClient
	peer_grpc.PeerServiceClient
	controllerbus_grpc.ControllerBusServiceClient
}

// bifrostDaemonClient implements BifrostDaemonClient.
type bifrostDaemonClient struct {
	stream_grpc.StreamServiceClient
	peer_grpc.PeerServiceClient
	controllerbus_grpc.ControllerBusServiceClient
}

// NewBifrostDaemonClient constructs a new bifrost daemon client.
func NewBifrostDaemonClient(cc *grpc.ClientConn) BifrostDaemonClient {
	return &bifrostDaemonClient{
		StreamServiceClient:        stream_grpc.NewStreamServiceClient(cc),
		PeerServiceClient:          peer_grpc.NewPeerServiceClient(cc),
		ControllerBusServiceClient: controllerbus_grpc.NewControllerBusServiceClient(cc),
	}
}
