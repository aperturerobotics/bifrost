package link

import (
	"errors"
	"strconv"

	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/controllerbus/directive"
)

// OpenStreamViaLink is a directive to open a stream with a peer over an
// established link.
// Not de-duplicated, intended to be used with OneOff.
type OpenStreamViaLink interface {
	// Directive indicates OpenStreamViaLink is a directive.
	directive.Directive

	// OpenStreamViaLinkUUID returns the link UUID to select.
	// Cannot be empty.
	OpenStreamViaLinkUUID() uint64
	// OpenStreamViaLinkProtocolID returns the protocol ID to negotiate with the peer.
	// Cannot be empty.
	OpenStreamViaLinkProtocolID() protocol.ID
	// OpenStreamViaLinkOpenOpts returns the open stream options.
	// Cannot be empty.
	OpenStreamViaLinkOpenOpts() stream.OpenOpts
	// OpenStreamViaLinkTransportConstraint returns a specific transport ID we want.
	// Can be empty.
	// Used to guarantee the same link is selected.
	OpenStreamViaLinkTransportConstraint() uint64
}

// openStreamViaLink implements OpenStreamViaLink with a peer ID constraint.
// Value: link.MountedStream
type openStreamViaLink struct {
	linkUUID            uint64
	protocolID          protocol.ID
	transportConstraint uint64
	openOpts            stream.OpenOpts
}

// NewOpenStreamViaLink constructs a new openStreamViaLink directive.
func NewOpenStreamViaLink(
	linkUUID uint64,
	protocolID protocol.ID,
	openOpts stream.OpenOpts,
	transportConstraint uint64,
) OpenStreamViaLink {
	return &openStreamViaLink{
		linkUUID:            linkUUID,
		protocolID:          protocolID,
		openOpts:            openOpts,
		transportConstraint: transportConstraint,
	}
}

// OpenStreamViaLinkUUID returns the link UUID to select.
// Cannot be empty.
func (d *openStreamViaLink) OpenStreamViaLinkUUID() uint64 {
	return d.linkUUID
}

// OpenStreamViaLinkProtocolID returns a specific protocol ID to negotiate.
func (d *openStreamViaLink) OpenStreamViaLinkProtocolID() protocol.ID {
	return d.protocolID
}

// OpenStreamViaLinkTransportConstraint returns the transport ID constraint.
// If empty, any transport is matched.
func (d *openStreamViaLink) OpenStreamViaLinkTransportConstraint() uint64 {
	return d.transportConstraint
}

// OpenStreamViaLinkOpenOpts returns the open options.
func (d *openStreamViaLink) OpenStreamViaLinkOpenOpts() stream.OpenOpts {
	return d.openOpts
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *openStreamViaLink) Validate() error {
	if d.linkUUID == 0 {
		return errors.New("link uuid must be specified")
	}
	if len(d.protocolID) == 0 {
		return errors.New("protocol id required")
	}

	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *openStreamViaLink) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		MaxValueCount:   1,
		MaxValueHardCap: true,
	}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *openStreamViaLink) IsEquivalent(other directive.Directive) bool {
	return false
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *openStreamViaLink) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *openStreamViaLink) GetName() string {
	return "OpenStreamViaLink"
}

// GetDebugVals returns the directive arguments as key/value pairs.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (d *openStreamViaLink) GetDebugVals() directive.DebugValues {
	vals := directive.NewDebugValues()
	vals["protocol-id"] = []string{string(d.OpenStreamViaLinkProtocolID())}
	vals["link-uuid"] = []string{strconv.FormatUint(d.OpenStreamViaLinkUUID(), 10)}
	if tpt := d.OpenStreamViaLinkTransportConstraint(); tpt != 0 {
		vals["transport"] = []string{strconv.FormatUint(tpt, 10)}
	}
	return vals
}

// _ is a type constraint
var _ OpenStreamViaLink = ((*openStreamViaLink)(nil))
