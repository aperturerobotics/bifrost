//+build !js

package api

import (
	"context"

	"github.com/aperturerobotics/bifrost/node"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
	"google.golang.org/grpc"
)

// Version is the API version.
var Version = semver.MustParse("0.0.1")

// API implements the GRPC API.
type API struct {
	bus  bus.Bus
	node node.Node
}

// NewAPI constructs a new instance of the API.
func NewAPI(bus bus.Bus, node node.Node) (*API, error) {
	return &API{bus: bus, node: node}, nil
}

// GetNodeInfo returns the node information
func (a *API) GetNodeInfo(
	ctx context.Context,
	req *GetNodeInfoRequest,
) (*GetNodeInfoResponse, error) {
	prettyID := a.node.GetPeerID().Pretty()

	var controllerInfos []*controller.Info
	controllers := a.bus.GetControllers()
	for _, controller := range controllers {
		ci := controller.GetControllerInfo()
		controllerInfos = append(controllerInfos, &ci)
	}

	return &GetNodeInfoResponse{
		NodeId:             prettyID,
		RunningControllers: controllerInfos,
	}, nil
}

// RegisterAsGRPCServer registers the API to the GRPC instance.
func (a *API) RegisterAsGRPCServer(grpcServer *grpc.Server) {
	RegisterBifrostDaemonServiceServer(grpcServer, a)
}

// _ is a type assertion
var _ BifrostDaemonServiceServer = ((*API)(nil))
