package agent

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/directive"
)

// AttachAgentToNode is a directive to attach an agent to a node.
type AttachAgentToNode interface {
	// Directive indicates AttachAgentToNode is a directive.
	directive.Directive

	// AttachAgentToNodeID returns a specific node ID we are looking for.
	// Cannot be empty.
	AttachAgentToNodeID() peer.ID
}

// AttachAgentToNodeSingleton implements AttachAgentToNode with a peer ID constraint.
type AttachAgentToNodeSingleton struct {
	peerIDConstraint peer.ID
}

// NewAttachAgentToNodeSingleton constructs a new AttachAgentToNodeSingleton directive.
func NewAttachAgentToNodeSingleton(peerID peer.ID) *AttachAgentToNodeSingleton {
	return &AttachAgentToNodeSingleton{peerIDConstraint: peerID}
}

// AttachAgentToNodeID returns a specific peer ID node we are looking for.
// If empty, any node is matched.
func (d *AttachAgentToNodeSingleton) AttachAgentToNodeID() peer.ID {
	return d.peerIDConstraint
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *AttachAgentToNodeSingleton) Validate() error {
	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *AttachAgentToNodeSingleton) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *AttachAgentToNodeSingleton) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(AttachAgentToNode)
	if !ok {
		return false
	}

	return d.peerIDConstraint == od.AttachAgentToNodeID()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *AttachAgentToNodeSingleton) Superceeds(other directive.Directive) bool {
	return false
}

// _ is a type constraint
var _ AttachAgentToNode = ((*AttachAgentToNodeSingleton)(nil))
