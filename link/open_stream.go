package link

import (
	"errors"
	"strconv"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/controllerbus/directive"
)

// OpenStreamWithPeer is a directive to open a stream with a peer.
// Not de-duplicated, intended to be used with OneOff.
type OpenStreamWithPeer interface {
	// Directive indicates OpenStreamWithPeer is a directive.
	directive.Directive

	// OpenStreamWithPeerProtocolID returns the protocol ID to negotiate with the peer.
	// Cannot be empty.
	OpenStreamWPProtocolID() protocol.ID
	// OpenStreamWithPeerTargetPeerID returns the target peer ID.
	// Cannot be empty.
	OpenStreamWPTargetPeerID() peer.ID
	// OpenStreamWithPeerOpenOpts returns the open stream options.
	// Cannot be empty.
	OpenStreamWPOpenOpts() stream.OpenOpts

	// OpenStreamWithPeerSourcePeerID returns the source peer ID.
	// Can be empty.
	OpenStreamWPSourcePeerID() peer.ID
	// OpenStreamWithPeerTransportConstraint returns a specific transport ID we want.
	// Can be empty.
	OpenStreamWPTransportConstraint() uint64
}

// openStreamWithPeer implements OpenStreamWithPeer with a peer ID constraint.
// Value: link.MountedStream
type openStreamWithPeer struct {
	protocolID                 protocol.ID
	sourcePeerID, targetPeerID peer.ID
	transportConstraint        uint64
	openOpts                   stream.OpenOpts
}

// NewOpenStreamWithPeer constructs a new openStreamWithPeer directive.
func NewOpenStreamWithPeer(
	protocolID protocol.ID,
	sourcePeerID, targetPeerID peer.ID,
	transportConstraint uint64,
	openOpts stream.OpenOpts,
) OpenStreamWithPeer {
	return &openStreamWithPeer{
		protocolID:          protocolID,
		sourcePeerID:        sourcePeerID,
		targetPeerID:        targetPeerID,
		transportConstraint: transportConstraint,
		openOpts:            openOpts,
	}
}

// OpenStreamWithPeerProtocolID returns a specific protocol ID to negotiate.
func (d *openStreamWithPeer) OpenStreamWPProtocolID() protocol.ID {
	return d.protocolID
}

// OpenStreamWithPeerSourcePeerID returns a specific peer ID node we are originating from.
// Can be empty.
func (d *openStreamWithPeer) OpenStreamWPSourcePeerID() peer.ID {
	return d.sourcePeerID
}

// OpenStreamWithPeerTargetPeerID returns a specific peer ID node we are looking for.
func (d *openStreamWithPeer) OpenStreamWPTargetPeerID() peer.ID {
	return d.targetPeerID
}

// OpenStreamWithPeerTransportConstraint returns the transport ID constraint.
// If empty, any transport is matched.
func (d *openStreamWithPeer) OpenStreamWPTransportConstraint() uint64 {
	return d.transportConstraint
}

// OpenStreamWithPeerOpenOpts returns the open options.
func (d *openStreamWithPeer) OpenStreamWPOpenOpts() stream.OpenOpts {
	return d.openOpts
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *openStreamWithPeer) Validate() error {
	if len(d.targetPeerID) == 0 {
		return errors.New("peer id constraint required")
	}

	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *openStreamWithPeer) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		MaxValueCount:   1,
		MaxValueHardCap: true,
	}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *openStreamWithPeer) IsEquivalent(other directive.Directive) bool {
	return false
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *openStreamWithPeer) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *openStreamWithPeer) GetName() string {
	return "OpenStreamWithPeer"
}

// GetDebugVals returns the directive arguments as key/value pairs.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (d *openStreamWithPeer) GetDebugVals() directive.DebugValues {
	vals := directive.NewDebugValues()
	vals["protocol-id"] = []string{string(d.OpenStreamWPProtocolID())}
	vals["target-peer"] = []string{d.OpenStreamWPTargetPeerID().Pretty()}
	vals["source-peer"] = []string{d.OpenStreamWPSourcePeerID().Pretty()}
	if tpt := d.OpenStreamWPTransportConstraint(); tpt != 0 {
		vals["transport"] = []string{strconv.FormatUint(tpt, 10)}
	}
	return vals
}

// _ is a type constraint
var _ OpenStreamWithPeer = ((*openStreamWithPeer)(nil))
