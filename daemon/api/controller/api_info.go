package bifrost_api_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/peer/grpc"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/controllerbus/bus/api"
	"github.com/pkg/errors"
)

// GetBusInfo returns the bus information
func (a *API) GetBusInfo(
	ctx context.Context,
	req *controllerbus_grpc.GetBusInfoRequest,
) (*controllerbus_grpc.GetBusInfoResponse, error) {
	var controllerInfos []*controller.Info
	controllers := a.bus.GetControllers()
	for _, controller := range controllers {
		ci := controller.GetControllerInfo()
		controllerInfos = append(controllerInfos, &ci)
	}

	directives := a.bus.GetDirectives()
	directiveInfo := make([]*controllerbus_grpc.DirectiveInfo, len(directives))
	for i, diri := range directives {
		dv := &controllerbus_grpc.DirectiveInfo{}
		dir := diri.GetDirective()
		debugDir, debugDirOk := dir.(directive.Debuggable)
		if debugDirOk {
			debugVals := debugDir.GetDebugVals()
			if debugVals != nil {
				for key, vals := range debugVals {
					dv.DebugVals = append(
						dv.DebugVals,
						&controllerbus_grpc.DebugValue{
							Key:    key,
							Values: vals,
						},
					)
				}
			}
		}
		dv.Name = dir.GetName()
		directiveInfo[i] = dv
	}

	return &controllerbus_grpc.GetBusInfoResponse{
		RunningControllers: controllerInfos,
		RunningDirectives:  directiveInfo,
	}, nil
}

// GetPeerInfo returns the peer information
func (a *API) GetPeerInfo(
	ctx context.Context,
	req *peer_grpc.GetPeerInfoRequest,
) (*peer_grpc.GetPeerInfoResponse, error) {
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

	resp := &peer_grpc.GetPeerInfoResponse{}
	di, dir, err := a.bus.AddDirective(
		peer.NewGetPeer(peerID),
		bus.NewCallbackHandler(func(v directive.AttachedValue) {
			pi, err := peer_grpc.NewPeerInfo(v.GetValue().(peer.Peer))
			if err != nil {
				return
			}
			resp.LocalPeers = append(resp.LocalPeers, pi)
		}, func(v directive.AttachedValue) {
			p := v.GetValue().(peer.Peer)
			pi, err := peer_grpc.NewPeerInfo(p)
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
