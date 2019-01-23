package peer

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
)

// GetPeerWithID gets a peer.
// If peer ID is empty, selects any peer.
func GetPeerWithID(
	ctx context.Context,
	b bus.Bus,
	peerIDConstraint ID,
) (Peer, directive.Reference, error) {
	v, ref, err := bus.ExecOneOff(ctx, b, NewGetPeer(peerIDConstraint), nil)
	if err != nil {
		return nil, nil, err
	}
	return v.GetValue().(Peer), ref, nil
}
