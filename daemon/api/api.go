//+build !js

package api

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Version is the API version.
var Version = semver.MustParse("0.0.1")

// API implements the GRPC API.
type API struct {
	bus bus.Bus
}

// NewAPI constructs a new instance of the API.
func NewAPI(bus bus.Bus) (*API, error) {
	return &API{bus: bus}, nil
}

// GetBusInfo returns the bus information
func (a *API) GetBusInfo(
	ctx context.Context,
	req *GetBusInfoRequest,
) (*GetBusInfoResponse, error) {
	var controllerInfos []*controller.Info
	controllers := a.bus.GetControllers()
	for _, controller := range controllers {
		ci := controller.GetControllerInfo()
		controllerInfos = append(controllerInfos, &ci)
	}

	return &GetBusInfoResponse{
		RunningControllers: controllerInfos,
	}, nil
}

// NewPeerInfo builds peer info from a peer.
func NewPeerInfo(p peer.Peer) (*PeerInfo, error) {
	pi := &PeerInfo{}
	pi.PeerId = peer.IDB58Encode(p.GetPeerID())
	return pi, nil
}

// GetPeerInfo returns the peer information
func (a *API) GetPeerInfo(
	ctx context.Context,
	req *GetPeerInfoRequest,
) (*GetPeerInfoResponse, error) {
	var peerID peer.ID
	if peerIDStr := req.GetPeerId(); peerIDStr != "" {
		var err error
		peerID, err = peer.IDB58Decode(peerIDStr)
		if err != nil {
			return nil, errors.Wrap(err, "decode peer id constraint")
		}
	}

	subCtx, subCtxCancel := context.WithCancel(ctx)
	defer subCtxCancel()

	resp := &GetPeerInfoResponse{}
	di, dir, err := a.bus.AddDirective(
		peer.NewGetPeer(peerID),
		bus.NewCallbackHandler(func(v directive.Value) {
			pi, err := NewPeerInfo(v.(peer.Peer))
			if err != nil {
				return
			}
			resp.LocalPeers = append(resp.LocalPeers, pi)
		}, func(v directive.Value) {
			p := v.(peer.Peer)
			pi, err := NewPeerInfo(p)
			if err != nil {
				return
			}
			for i, r := range resp.LocalPeers {
				if pi.PeerId == r.PeerId {
					resp.LocalPeers[i] = resp.LocalPeers[len(resp.LocalPeers)-1]
					resp.LocalPeers[len(resp.LocalPeers)-1] = nil
					resp.LocalPeers = resp.LocalPeers[:len(resp.LocalPeers)-1]
					break
				}
			}
		}, func() {
			subCtxCancel()
		}),
	)
	if err != nil {
		return nil, err
	}
	defer dir.Release()

	rcb := di.AddIdleCallback(func() {
		subCtxCancel()
	})
	if rcb != nil {
		defer rcb()
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-subCtx.Done():
	}

	return resp, nil
}

// RegisterAsGRPCServer registers the API to the GRPC instance.
func (a *API) RegisterAsGRPCServer(grpcServer *grpc.Server) {
	RegisterBifrostDaemonServiceServer(grpcServer, a)
}

// _ is a type assertion
var _ BifrostDaemonServiceServer = ((*API)(nil))
