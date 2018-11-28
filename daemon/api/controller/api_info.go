package api_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
)

// GetBusInfo returns the bus information
func (a *API) GetBusInfo(
	ctx context.Context,
	req *api.GetBusInfoRequest,
) (*api.GetBusInfoResponse, error) {
	var controllerInfos []*controller.Info
	controllers := a.bus.GetControllers()
	for _, controller := range controllers {
		ci := controller.GetControllerInfo()
		controllerInfos = append(controllerInfos, &ci)
	}

	directives := a.bus.GetDirectives()
	directiveInfo := make([]*api.DirectiveInfo, len(directives))
	for i, directive := range directives {
		dv := &api.DirectiveInfo{}
		dir := directive.GetDirective()
		debugVals := dir.GetDebugVals()
		if debugVals != nil {
			for key, vals := range debugVals {
				dv.DebugVals = append(dv.DebugVals, &api.DebugValue{
					Key:    key,
					Values: vals,
				})
			}
		}
		dv.Name = dir.GetName()
		directiveInfo[i] = dv
	}

	return &api.GetBusInfoResponse{
		RunningControllers: controllerInfos,
		RunningDirectives:  directiveInfo,
	}, nil
}

// GetPeerInfo returns the peer information
func (a *API) GetPeerInfo(
	ctx context.Context,
	req *api.GetPeerInfoRequest,
) (*api.GetPeerInfoResponse, error) {
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

	resp := &api.GetPeerInfoResponse{}
	di, dir, err := a.bus.AddDirective(
		peer.NewGetPeer(peerID),
		bus.NewCallbackHandler(func(v directive.Value) {
			pi, err := api.NewPeerInfo(v.(peer.Peer))
			if err != nil {
				return
			}
			resp.LocalPeers = append(resp.LocalPeers, pi)
		}, func(v directive.Value) {
			p := v.(peer.Peer)
			pi, err := api.NewPeerInfo(p)
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
