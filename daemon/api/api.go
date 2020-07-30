package api

import (
	peer_grpc "github.com/aperturerobotics/bifrost/peer/grpc"
	pubsub_grpc "github.com/aperturerobotics/bifrost/pubsub/grpc"
	stream_grpc "github.com/aperturerobotics/bifrost/stream/grpc"
	bus_grpc "github.com/aperturerobotics/controllerbus/bus/api"
	"google.golang.org/grpc"
)

// BifrostDaemonServer is the bifrost daemon server interface.
type BifrostDaemonServer interface {
	stream_grpc.StreamServiceServer
	peer_grpc.PeerServiceServer
	bus_grpc.ControllerBusServiceServer
	pubsub_grpc.PubSubServiceServer
}

// RegisterAsGRPCServer registers a server with a grpc server.
func RegisterAsGRPCServer(s BifrostDaemonServer, grpcServer *grpc.Server) {
	stream_grpc.RegisterStreamServiceServer(grpcServer, s)
	peer_grpc.RegisterPeerServiceServer(grpcServer, s)
	bus_grpc.RegisterControllerBusServiceServer(grpcServer, s)
	pubsub_grpc.RegisterPubSubServiceServer(grpcServer, s)
}

// BifrostDaemonClient is the bifrost daemon client interface.
type BifrostDaemonClient interface {
	stream_grpc.StreamServiceClient
	peer_grpc.PeerServiceClient
	bus_grpc.ControllerBusServiceClient
	pubsub_grpc.PubSubServiceClient
}

// bifrostDaemonClient implements BifrostDaemonClient.
type bifrostDaemonClient struct {
	stream_grpc.StreamServiceClient
	peer_grpc.PeerServiceClient
	bus_grpc.ControllerBusServiceClient
	pubsub_grpc.PubSubServiceClient
}

// NewBifrostDaemonClient constructs a new bifrost daemon client.
func NewBifrostDaemonClient(cc *grpc.ClientConn) BifrostDaemonClient {
	return &bifrostDaemonClient{
		StreamServiceClient:        stream_grpc.NewStreamServiceClient(cc),
		PeerServiceClient:          peer_grpc.NewPeerServiceClient(cc),
		ControllerBusServiceClient: bus_grpc.NewControllerBusServiceClient(cc),
		PubSubServiceClient:        pubsub_grpc.NewPubSubServiceClient(cc),
	}
}
