package peer

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
)

// GetPeerWithID executes a GetNodeSingleton directive.
// If peer ID is empty, selects any node.
func GetPeerWithID(
	ctx context.Context,
	b bus.Bus,
	peerIDConstraint ID,
) (Peer, error) {
	v, ref, err := bus.ExecOneOff(ctx, b, NewGetPeerSingleton(peerIDConstraint), nil)
	if err != nil {
		return nil, err
	}
	ref.Release()

	return v.(Peer), nil
}
