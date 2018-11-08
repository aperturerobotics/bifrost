package node

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/directive"
)

// GetNode is a directive to lookup a node on a controller.
type GetNode interface {
	// Directive indicates GetNode is a directive.
	directive.Directive

	// GetNodePeerIDConstraint returns a specific peer ID node we are looking for.
	// If empty, any node is matched.
	GetNodePeerIDConstraint() peer.ID
}

// GetNodeSingleton implements GetNode with a peer ID constraint.
type GetNodeSingleton struct {
	peerIDConstraint peer.ID
}

// NewGetNodeSingleton constructs a new GetNodeSingleton directive.
func NewGetNodeSingleton(peerID peer.ID) *GetNodeSingleton {
	return &GetNodeSingleton{peerIDConstraint: peerID}
}

// GetNodePeerIDConstraint returns a specific peer ID node we are looking for.
// If empty, any node is matched.
func (d *GetNodeSingleton) GetNodePeerIDConstraint() peer.ID {
	return d.peerIDConstraint
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *GetNodeSingleton) Validate() error {
	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *GetNodeSingleton) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *GetNodeSingleton) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(GetNode)
	if !ok {
		return false
	}

	return d.peerIDConstraint == od.GetNodePeerIDConstraint()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *GetNodeSingleton) Superceeds(other directive.Directive) bool {
	return false
}

// _ is a type constraint
var _ GetNode = ((*GetNodeSingleton)(nil))
