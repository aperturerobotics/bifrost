package link

import (
	"errors"
	"strconv"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/controllerbus/directive"
)

// OpenStream is a directive to open a stream with a peer.
// Not de-duplicated, intended to be used with OneOff.
type OpenStream interface {
	// Directive indicates OpenStream is a directive.
	directive.Directive

	// OpenStreamProtocolID returns the protocol ID to negotiate with the peer.
	// Cannot be empty.
	OpenStreamProtocolID() protocol.ID
	// OpenStreamTargetPeerID returns the target peer ID.
	// Cannot be empty.
	OpenStreamTargetPeerID() peer.ID
	// OpenStreamOpenOpts returns the open stream options.
	// Cannot be empty.
	OpenStreamOpenOpts() stream.OpenOpts

	// OpenStreamSourcePeerID returns the source peer ID.
	// Can be empty.
	OpenStreamSourcePeerID() peer.ID
	// OpenStreamTransportConstraint returns a specific transport ID we want.
	// Can be empty.
	OpenStreamTransportConstraint() uint64
}

// OpenStreamWithPeer implements OpenStream with a peer ID constraint.
// Value: link.MountedStream
type OpenStreamWithPeer struct {
	protocolID                 protocol.ID
	sourcePeerID, targetPeerID peer.ID
	transportConstraint        uint64
	openOpts                   stream.OpenOpts
}

// NewOpenStreamWithPeer constructs a new OpenStreamWithPeer directive.
func NewOpenStreamWithPeer(
	protocolID protocol.ID,
	sourcePeerID, targetPeerID peer.ID,
	transportConstraint uint64,
	openOpts stream.OpenOpts,
) *OpenStreamWithPeer {
	return &OpenStreamWithPeer{
		protocolID:          protocolID,
		sourcePeerID:        sourcePeerID,
		targetPeerID:        targetPeerID,
		transportConstraint: transportConstraint,
		openOpts:            openOpts,
	}
}

// OpenStreamProtocolID returns a specific protocol ID to negotiate.
func (d *OpenStreamWithPeer) OpenStreamProtocolID() protocol.ID {
	return d.protocolID
}

// OpenStreamSourcePeerID returns a specific peer ID node we are originating from.
// Can be empty.
func (d *OpenStreamWithPeer) OpenStreamSourcePeerID() peer.ID {
	return d.sourcePeerID
}

// OpenStreamTargetPeerID returns a specific peer ID node we are looking for.
func (d *OpenStreamWithPeer) OpenStreamTargetPeerID() peer.ID {
	return d.targetPeerID
}

// OpenStreamTransportConstraint returns the transport ID constraint.
// If empty, any transport is matched.
func (d *OpenStreamWithPeer) OpenStreamTransportConstraint() uint64 {
	return d.transportConstraint
}

// OpenStreamOpenOpts returns the open options.
func (d *OpenStreamWithPeer) OpenStreamOpenOpts() stream.OpenOpts {
	return d.openOpts
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *OpenStreamWithPeer) Validate() error {
	if len(d.targetPeerID) == 0 {
		return errors.New("peer id constraint required")
	}

	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *OpenStreamWithPeer) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		MaxValueCount:   1,
		MaxValueHardCap: true,
	}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *OpenStreamWithPeer) IsEquivalent(other directive.Directive) bool {
	return false
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *OpenStreamWithPeer) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *OpenStreamWithPeer) GetName() string {
	return "OpenStreamWithPeer"
}

// GetDebugVals returns the directive arguments as key/value pairs.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (d *OpenStreamWithPeer) GetDebugVals() directive.DebugValues {
	vals := directive.NewDebugValues()
	vals["protocol-id"] = []string{string(d.OpenStreamProtocolID())}
	vals["target-peer"] = []string{d.OpenStreamTargetPeerID().Pretty()}
	vals["source-peer"] = []string{d.OpenStreamSourcePeerID().Pretty()}
	if tpt := d.OpenStreamTransportConstraint(); tpt != 0 {
		vals["transport"] = []string{strconv.FormatUint(tpt, 10)}
	}
	return vals
}

// _ is a type constraint
var _ OpenStream = ((*OpenStreamWithPeer)(nil))
