package transport

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LookupTransport is a directive to lookup running transports.
// Value type: transport.Transport.
type LookupTransport interface {
	// Directive indicates LookupTransport is a directive.
	directive.Directive

	// LookupTransportNodeIDConstraint returns a specific node ID we are looking for.
	// Can be empty.
	LookupTransportNodeIDConstraint() peer.ID
	// LookupTransportIDConstraint returns a specific transport ID we are looking for.
	// Can be empty.
	LookupTransportIDConstraint() uint64
}

// lookupTransport implements LookupTransport
type lookupTransport struct {
	nodeIDConstraint      peer.ID
	transportIDConstraint uint64
}

// NewLookupTransport constructs a new LookupTransport directive.
func NewLookupTransport(nodeID peer.ID, transportID uint64) LookupTransport {
	return &lookupTransport{
		nodeIDConstraint:      nodeID,
		transportIDConstraint: transportID,
	}
}

// LookupTransportNodeIDConstraint returns a specific peer ID node we are looking for.
// If empty, any node is matched.
func (d *lookupTransport) LookupTransportNodeIDConstraint() peer.ID {
	return d.nodeIDConstraint
}

// LookupTransportIDConstraint returns a specific transport ID we are looking for.
// If empty, any transport id is matched.
func (d *lookupTransport) LookupTransportIDConstraint() uint64 {
	return d.transportIDConstraint
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *lookupTransport) Validate() error {
	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *lookupTransport) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *lookupTransport) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(LookupTransport)
	if !ok {
		return false
	}

	return d.LookupTransportIDConstraint() == od.LookupTransportIDConstraint() &&
		d.LookupTransportNodeIDConstraint() == od.LookupTransportNodeIDConstraint()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *lookupTransport) Superceeds(other directive.Directive) bool {
	return false
}

// _ is a type assertion
var _ LookupTransport = ((*lookupTransport)(nil))
