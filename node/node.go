package node

import (
	"github.com/aperturerobotics/bifrost/peer"
)

// Node is a full routable peer in the network.
type Node interface {
	// Peer indicates Node is a Peer.
	peer.Peer

	// RegisterTransport registers a new transport with the node.
	// Returns a release handle,
	// AcceptLink decides if
}
