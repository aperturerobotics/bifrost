package api

import (
	"context"

	"github.com/aperturerobotics/bifrost/node"
	"github.com/blang/semver"
	"google.golang.org/grpc"
)

// Version is the API version.
var Version = semver.MustParse("0.0.1")

// API implements the GRPC API.
type API struct {
	node node.Node
}

// NewAPI constructs a new instance of the API.
func NewAPI(node node.Node) (*API, error) {
	return &API{node: node}, nil
}

// GetNodeInfo returns the node information
func (a *API) GetNodeInfo(
	ctx context.Context,
	req *GetNodeInfoRequest,
) (*GetNodeInfoResponse, error) {
	prettyID := a.node.GetPeerID().Pretty()
	return &GetNodeInfoResponse{
		NodeId: prettyID,
	}, nil
}

// RegisterAsGRPCServer registers the API to the GRPC instance.
func (a *API) RegisterAsGRPCServer(grpcServer *grpc.Server) {
	RegisterBifrostDaemonServiceServer(grpcServer, a)
}

// _ is a type assertion
var _ BifrostDaemonServiceServer = ((*API)(nil))
