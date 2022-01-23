package bifrost_api

import (
	peer_grpc "github.com/aperturerobotics/bifrost/peer/rpc"
	pubsub_grpc "github.com/aperturerobotics/bifrost/pubsub/grpc"
	stream_grpc "github.com/aperturerobotics/bifrost/stream/grpc"
	bus_grpc "github.com/aperturerobotics/controllerbus/bus/api"
	"google.golang.org/grpc"
)

// BifrostAPIServer is the bifrost daemon server interface.
type BifrostAPIServer interface {
	stream_grpc.StreamServiceServer
	peer_grpc.PeerServiceServer
	pubsub_grpc.PubSubServiceServer
}

// bus_grpc.ControllerBusServiceServer

// RegisterAsGRPCServer registers a server with a grpc server.
func RegisterAsGRPCServer(s BifrostAPIServer, grpcServer *grpc.Server) {
	stream_grpc.RegisterStreamServiceServer(grpcServer, s)
	peer_grpc.RegisterPeerServiceServer(grpcServer, s)
	// bus_grpc.RegisterControllerBusServiceServer(grpcServer, s)
	pubsub_grpc.RegisterPubSubServiceServer(grpcServer, s)
}

// BifrostAPIClient is the bifrost daemon client interface.
type BifrostAPIClient interface {
	stream_grpc.StreamServiceClient
	peer_grpc.PeerServiceClient
	bus_grpc.ControllerBusServiceClient
	pubsub_grpc.PubSubServiceClient
}

// bifrostAPIClient implements BifrostAPIClient.
type bifrostAPIClient struct {
	stream_grpc.StreamServiceClient
	peer_grpc.PeerServiceClient
	bus_grpc.ControllerBusServiceClient
	pubsub_grpc.PubSubServiceClient
}

// NewBifrostAPIClient constructs a new bifrost daemon client.
func NewBifrostAPIClient(cc *grpc.ClientConn) BifrostAPIClient {
	return &bifrostAPIClient{
		StreamServiceClient:        stream_grpc.NewStreamServiceClient(cc),
		PeerServiceClient:          peer_grpc.NewPeerServiceClient(cc),
		ControllerBusServiceClient: bus_grpc.NewControllerBusServiceClient(cc),
		PubSubServiceClient:        pubsub_grpc.NewPubSubServiceClient(cc),
	}
}
