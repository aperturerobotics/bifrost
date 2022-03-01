package bifrost_api

import (
	peer_api "github.com/aperturerobotics/bifrost/peer/api"
	pubsub_api "github.com/aperturerobotics/bifrost/pubsub/api"
	stream_api "github.com/aperturerobotics/bifrost/stream/api"
	bus_api "github.com/aperturerobotics/controllerbus/bus/api"
	"storj.io/drpc"
)

// BifrostAPIServer is the bifrost daemon server interface.
type BifrostAPIServer interface {
	bus_api.DRPCControllerBusServiceServer
	stream_api.DRPCStreamServiceServer
	peer_api.DRPCPeerServiceServer
	pubsub_api.DRPCPubSubServiceServer
}

// RegisterAsDRPCServer registers a server with a DRPC mux.
func RegisterAsDRPCServer(s BifrostAPIServer, mux drpc.Mux) {
	_ = bus_api.DRPCRegisterControllerBusService(mux, s)
	_ = stream_api.DRPCRegisterStreamService(mux, s)
	_ = peer_api.DRPCRegisterPeerService(mux, s)
	_ = pubsub_api.DRPCRegisterPubSubService(mux, s)
}

// BifrostAPIClient is the bifrost daemon client interface.
type BifrostAPIClient interface {
	bus_api.DRPCControllerBusServiceClient
	stream_api.DRPCStreamServiceClient
	peer_api.DRPCPeerServiceClient
	pubsub_api.DRPCPubSubServiceClient
}

// bifrostAPIClient implements BifrostAPIClient.
type bifrostAPIClient struct {
	bus_api.DRPCControllerBusServiceClient
	stream_api.DRPCStreamServiceClient
	peer_api.DRPCPeerServiceClient
	pubsub_api.DRPCPubSubServiceClient
	cc drpc.Conn
}

// NewBifrostAPIClient constructs a new bifrost api client.
func NewBifrostAPIClient(cc drpc.Conn) BifrostAPIClient {
	return &bifrostAPIClient{
		DRPCControllerBusServiceClient: bus_api.NewDRPCControllerBusServiceClient(cc),
		DRPCStreamServiceClient:        stream_api.NewDRPCStreamServiceClient(cc),
		DRPCPeerServiceClient:          peer_api.NewDRPCPeerServiceClient(cc),
		DRPCPubSubServiceClient:        pubsub_api.NewDRPCPubSubServiceClient(cc),
		cc:                             cc,
	}
}

// DRPCConn returns the drpc connection.
func (c *bifrostAPIClient) DRPCConn() drpc.Conn { return c.cc }
