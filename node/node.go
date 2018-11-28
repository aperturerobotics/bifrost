package node

import (
	"github.com/aperturerobotics/bifrost/peer"
)

// Node is a full routable peer in the network.
type Node interface {
	// Peer indicates Node is a Peer.
	peer.Peer
}
