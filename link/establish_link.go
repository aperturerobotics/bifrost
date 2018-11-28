package link

import (
	"errors"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/directive"
)

// EstablishLinkWithPeer is a directive to establish a link with a peer.
type EstablishLinkWithPeer interface {
	// Directive indicates EstablishLinkWithPeer is a directive.
	directive.Directive

	// EstablishLinkWPIDConstraint returns the target peer ID.
	// Cannot be empty.
	EstablishLinkWPIDConstraint() peer.ID
}

// establishLinkWithPeer implements EstablishLinkWithPeer with a peer ID constraint.
type establishLinkWithPeer struct {
	peerIDConstraint peer.ID
}

// NewEstablishLinkWithPeer constructs a new EstablishLinkWithPeer directive.
func NewEstablishLinkWithPeer(peerID peer.ID) EstablishLinkWithPeer {
	return &establishLinkWithPeer{peerIDConstraint: peerID}
}

// EstablishLinkWPIDConstraint returns the target peer ID.
func (d *establishLinkWithPeer) EstablishLinkWPIDConstraint() peer.ID {
	return d.peerIDConstraint
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *establishLinkWithPeer) Validate() error {
	if len(d.peerIDConstraint) == 0 {
		return errors.New("peer id constraint required")
	}

	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *establishLinkWithPeer) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *establishLinkWithPeer) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(EstablishLinkWithPeer)
	if !ok {
		return false
	}

	return d.peerIDConstraint == od.EstablishLinkWPIDConstraint()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *establishLinkWithPeer) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *establishLinkWithPeer) GetName() string {
	return "establishLinkWithPeer"
}

// GetDebugVals returns the directive arguments as k/v pairs.
// This is not necessarily unique, and is primarily intended for display.
func (d *establishLinkWithPeer) GetDebugVals() directive.DebugValues {
	vals := directive.NewDebugValues()
	vals["peer-id"] = []string{d.peerIDConstraint.Pretty()}
	return vals
}

// _ is a type constraint
var _ EstablishLinkWithPeer = ((*establishLinkWithPeer)(nil))
