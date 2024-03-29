package bifrost_api

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
	peer_api "github.com/aperturerobotics/bifrost/peer/api"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
)

// GetPeerInfo returns the peer information
func (a *API) GetPeerInfo(
	ctx context.Context,
	req *peer_api.GetPeerInfoRequest,
) (*peer_api.GetPeerInfoResponse, error) {
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

	resp := &peer_api.GetPeerInfoResponse{}
	di, dir, err := a.bus.AddDirective(
		peer.NewGetPeer(peerID),
		bus.NewCallbackHandler(func(v directive.AttachedValue) {
			pi, err := peer_api.NewPeerInfo(v.GetValue().(peer.Peer))
			if err != nil {
				return
			}
			resp.LocalPeers = append(resp.LocalPeers, pi)
		}, func(v directive.AttachedValue) {
			p := v.GetValue().(peer.Peer)
			pi, err := peer_api.NewPeerInfo(p)
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

	errCh := make(chan error, 1)
	rcb := di.AddIdleCallback(func(errs []error) {
		if len(errs) != 0 {
			select {
			case errCh <- errs[0]:
				return
			default:
			}
		}

		subCtxCancel()
	})
	if rcb != nil {
		defer rcb()
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case <-subCtx.Done():
	}

	return resp, nil
}
