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

// EstablishLinkWithPeer implements EstablishLink with a peer ID constraint.
type EstablishLinkWithPeer struct {
	peerIDConstraint peer.ID
}

// NewEstablishLinkWithPeer constructs a new EstablishLinkWithPeer directive.
func NewEstablishLinkWithPeer(peerID peer.ID) *EstablishLinkWithPeer {
	return &EstablishLinkWithPeer{peerIDConstraint: peerID}
}

// EstablishLinkPeerIDConstraint returns a specific peer ID node we are looking for.
// If empty, any node is matched.
func (d *EstablishLinkWithPeer) EstablishLinkPeerIDConstraint() peer.ID {
	return d.peerIDConstraint
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *EstablishLinkWithPeer) Validate() error {
	if len(d.peerIDConstraint) == 0 {
		return errors.New("peer id constraint required")
	}

	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *EstablishLinkWithPeer) GetValueOptions() directive.ValueOptions {
	// no value cap
	return directive.ValueOptions{}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *EstablishLinkWithPeer) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(EstablishLink)
	if !ok {
		return false
	}

	return d.peerIDConstraint == od.EstablishLinkPeerIDConstraint()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *EstablishLinkWithPeer) Superceeds(other directive.Directive) bool {
	return false
}

// _ is a type constraint
var _ EstablishLink = ((*EstablishLinkWithPeer)(nil))
