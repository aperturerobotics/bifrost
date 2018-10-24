package agent

import (
	"github.com/aperturerobotics/bifrost/peer"
)

// Agent is an identified process attached to a Node.
// The agent controller implements the agent interface.
// This is the end-user API/handle to an agent.
type Agent interface {
	// Peer indicates Agent is a Peer.
	peer.Peer
}
