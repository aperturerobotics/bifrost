package webrtc

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/util/scrc"
)

// NewTransportUUID builds the UUID for a transport with a local address and peer id
func NewTransportUUID(transportType string, peerID peer.ID) uint64 {
	return scrc.Crc64(
		[]byte(ControllerID),
		[]byte("/"),
		[]byte(transportType),
		[]byte("/"),
		[]byte(peerID.String()),
	)
}

// NewLinkUUID builds the UUID for a link
func NewLinkUUID(transportType, localPeerID, remotePeerID peer.ID) uint64 {
	return scrc.Crc64(
		[]byte(ControllerID),
		[]byte("/"),
		[]byte(transportType),
		[]byte("/"),
		[]byte(localPeerID.String()),
		[]byte("/"),
		[]byte(remotePeerID.String()),
	)
}
