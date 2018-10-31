package link

import (
	"errors"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/controllerbus/directive"
)

// HandleMountedStream is a directive to return a mounted stream handler for a
// protocol ID.
// Value is of type link.MountedStreamHandler.
type HandleMountedStream interface {
	// Directive indicates HandleMountedStream is a directive.
	directive.Directive

	// HandleMountedStreamProtocolID returns the protocol ID we are requesting a
	// handler for. Cannot be empty.
	HandleMountedStreamProtocolID() protocol.ID
	// HandleMountedStreamLocalPeerID returns the local peer ID we are
	// requesting a handler for. Cannot be empty.
	HandleMountedStreamLocalPeerID() peer.ID
	// HandleMountedStreamRemotePeerID returns the remote peer ID we are
	// requesting a handler for. Cannot be empty.
	HandleMountedStreamRemotePeerID() peer.ID
}

// HandleMountedStreamWithProtocolID implements HandleMountedStream with a
// protocol ID and peer ID.
type HandleMountedStreamWithProtocolID struct {
	protocolID   protocol.ID
	localPeerID  peer.ID
	remotePeerID peer.ID
}

// NewHandleMountedStreamWithProtocolID constructs a new HandleMountedStreamWithProtocolID directive.
func NewHandleMountedStreamWithProtocolID(
	protocolID protocol.ID,
	localPeerID, remotePeerID peer.ID,
) *HandleMountedStreamWithProtocolID {
	return &HandleMountedStreamWithProtocolID{
		protocolID:   protocolID,
		localPeerID:  localPeerID,
		remotePeerID: remotePeerID,
	}
}

// HandleMountedStreamProtocolID returns the protocol ID we are requesting a
// handler for. Cannot be empty.
func (d *HandleMountedStreamWithProtocolID) HandleMountedStreamProtocolID() protocol.ID {
	return d.protocolID
}

// HandleMountedStreamLocalPeerID returns the local peer ID we are
// requesting a handler for. Cannot be empty.
func (d *HandleMountedStreamWithProtocolID) HandleMountedStreamLocalPeerID() peer.ID {
	return d.localPeerID
}

// HandleMountedStreamRemotePeerID returns the remote peer ID we are
// requesting a handler for. Cannot be empty.
func (d *HandleMountedStreamWithProtocolID) HandleMountedStreamRemotePeerID() peer.ID {
	return d.remotePeerID
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *HandleMountedStreamWithProtocolID) Validate() error {
	if len(d.localPeerID) == 0 {
		return errors.New("local peer id required")
	}
	if len(d.remotePeerID) == 0 {
		return errors.New("remote peer id required")
	}
	if len(d.protocolID) == 0 {
		return errors.New("protocol id required")
	}

	return nil
}

// GetValueOptions returns options relating to value handling.
func (d *HandleMountedStreamWithProtocolID) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		MaxValueCount:   1,
		MaxValueHardCap: true,
	}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *HandleMountedStreamWithProtocolID) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(HandleMountedStream)
	if !ok {
		return false
	}

	return d.protocolID == od.HandleMountedStreamProtocolID() &&
		d.localPeerID == od.HandleMountedStreamLocalPeerID()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *HandleMountedStreamWithProtocolID) Superceeds(other directive.Directive) bool {
	return false
}

// _ is a type constraint
var _ HandleMountedStream = ((*HandleMountedStreamWithProtocolID)(nil))
