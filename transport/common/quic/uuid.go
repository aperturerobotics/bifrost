package transport_quic

import (
	"net"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/util/scrc"
)

// NewTransportUUID builds the UUID for a transport with a local address and peer id
func NewTransportUUID(localAddr string, peerID peer.ID) uint64 {
	return scrc.Crc64(
		[]byte("bifrost/quic/"),
		[]byte(localAddr),
		[]byte("/"),
		[]byte(peerID.Pretty()),
	)
}

// NewLinkUUID builds the UUID for a link
func NewLinkUUID(localAddr, remoteAddr net.Addr, peerID peer.ID) uint64 {
	return scrc.Crc64(
		[]byte("quic"),
		[]byte(localAddr.String()),
		[]byte(remoteAddr.String()),
		[]byte(peerID),
	)
}
