package transport

import "github.com/aperturerobotics/controllerbus/config"

// Config contains common parameters all transport configs must have.
type Config interface {
	// Config indicates this is a controllerbus config.
	config.Config
	// GetNodePeerId returns the node peer ID constraint.
	GetNodePeerId() string
	// SetNodePeerId sets the node peer ID field.
	SetNodePeerId(peerID string)
}
