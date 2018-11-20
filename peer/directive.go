package peer

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// GetPeer is a directive to lookup a peer on a controller.
type GetPeer interface {
	// Directive indicates GetPeer is a directive.
	directive.Directive

	// GetPeerIDConstraint returns a specific peer ID node we are looking for.
	// If empty, any node is matched.
	GetPeerIDConstraint() ID
}

// GetPeerSingleton implements GetPeer with a peer ID constraint.
type GetPeerSingleton struct {
	peerIDConstraint ID
}

// NewGetPeerSingleton constructs a new GetPeerSingleton directive.
func NewGetPeerSingleton(peerID ID) *GetPeerSingleton {
	return &GetPeerSingleton{peerIDConstraint: peerID}
}

// GetPeerIDConstraint returns a specific peer ID node we are looking for.
// If empty, any node is matched.
func (d *GetPeerSingleton) GetPeerIDConstraint() ID {
	return d.peerIDConstraint
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *GetPeerSingleton) Validate() error {
	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *GetPeerSingleton) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *GetPeerSingleton) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(GetPeer)
	if !ok {
		return false
	}

	return d.peerIDConstraint == od.GetPeerIDConstraint()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *GetPeerSingleton) Superceeds(other directive.Directive) bool {
	return false
}

// _ is a type constraint
var _ GetPeer = ((*GetPeerSingleton)(nil))
