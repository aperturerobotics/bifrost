package tptaddr

import (
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
)

// DialTptAddr is a directive to establish a link with a peer via a transport address.
//
// Value: Link
type DialTptAddr interface {
	// Directive indicates DialTptAddr is a directive.
	directive.Directive

	// DialTptAddr returns the transport addr to dial.
	// usually {transport-type-id}|{addr}
	// Cannot be empty.
	DialTptAddr() string
	// DialTptAddrSourcePeerId returns the source peer ID.
	// Can be empty to allow any.
	DialTptAddrSourcePeerId() peer.ID
	// DialTptAddrTargetPeerId returns the target peer ID.
	// Cannot be empty.
	DialTptAddrTargetPeerId() peer.ID
}

// DialTptAddrValue is the type emitted when resolving DialTptAddr.
type DialTptAddrValue = link.Link

// dialTptAddr implements DialTptAddr
type dialTptAddr struct {
	addr      string
	src, dest peer.ID
}

// NewDialTptAddr constructs a new DialTptAddr directive.
func NewDialTptAddr(addr string, srcPeer, destPeer peer.ID) DialTptAddr {
	return &dialTptAddr{addr: addr, src: srcPeer, dest: destPeer}
}

// DialTptAddr returns the transport addr to dial.
// usually {transport-type-id}|{addr}
// Cannot be empty.
func (d *dialTptAddr) DialTptAddr() string {
	return d.addr
}

// DialTptAddrTargetPeerId returns the target peer ID.
func (d *dialTptAddr) DialTptAddrTargetPeerId() peer.ID {
	return d.dest
}

// DialTptAddrSourcePeerId returns the src peer ID.
func (d *dialTptAddr) DialTptAddrSourcePeerId() peer.ID {
	return d.src
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *dialTptAddr) Validate() error {
	if len(d.dest) == 0 {
		return errors.Wrap(peer.ErrEmptyPeerID, "destination")
	}
	if _, _, err := ParseTptAddr(d.addr); err != nil {
		return err
	}

	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *dialTptAddr) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *dialTptAddr) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(DialTptAddr)
	if !ok {
		return false
	}

	return d.DialTptAddrTargetPeerId() == od.DialTptAddrTargetPeerId() &&
		d.DialTptAddrSourcePeerId() == od.DialTptAddrSourcePeerId() &&
		d.DialTptAddr() == od.DialTptAddr()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *dialTptAddr) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *dialTptAddr) GetName() string {
	return "DialTptAddr"
}

// GetDebugVals returns the directive arguments as k/v pairs.
// This is not necessarily unique, and is primarily intended for display.
func (d *dialTptAddr) GetDebugVals() directive.DebugValues {
	vals := directive.NewDebugValues()
	vals["peer-id"] = []string{d.DialTptAddrTargetPeerId().Pretty()}
	vals["tpt-addr"] = []string{d.DialTptAddr()}
	if src := d.DialTptAddrSourcePeerId(); len(src) != 0 {
		vals["from-peer-id"] = []string{d.DialTptAddrSourcePeerId().Pretty()}
	}
	return vals
}

// _ is a type constraint
var _ DialTptAddr = ((*dialTptAddr)(nil))
