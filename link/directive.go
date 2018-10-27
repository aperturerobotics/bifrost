package link

import (
	"errors"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/directive"
)

// EstablishLink is a directive to establish a link with a peer.
type EstablishLink interface {
	// Directive indicates EstablishLink is a directive.
	directive.Directive

	// EstablishLinkPeerIDConstraint returns a specific peer ID we are looking for.
	// Cannot be empty.
	EstablishLinkPeerIDConstraint() peer.ID
}

// EstablishLinkSingleton implements EstablishLink with a peer ID constraint.
type EstablishLinkSingleton struct {
	peerIDConstraint peer.ID
}

// NewEstablishLinkSingleton constructs a new EstablishLinkSingleton directive.
func NewEstablishLinkSingleton(peerID peer.ID) *EstablishLinkSingleton {
	return &EstablishLinkSingleton{peerIDConstraint: peerID}
}

// EstablishLinkPeerIDConstraint returns a specific peer ID node we are looking for.
// If empty, any node is matched.
func (d *EstablishLinkSingleton) EstablishLinkPeerIDConstraint() peer.ID {
	return d.peerIDConstraint
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *EstablishLinkSingleton) Validate() error {
	if len(d.peerIDConstraint) == 0 {
		return errors.New("peer id constraint required")
	}

	return nil
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *EstablishLinkSingleton) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(EstablishLink)
	if !ok {
		return false
	}

	return d.peerIDConstraint == od.EstablishLinkPeerIDConstraint()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *EstablishLinkSingleton) Superceeds(other directive.Directive) bool {
	return false
}

// _ is a type constraint
var _ EstablishLink = ((*EstablishLinkSingleton)(nil))
