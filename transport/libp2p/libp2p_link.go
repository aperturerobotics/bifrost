package libp2p

import (
	"github.com/aperturerobotics/bifrost/link"
	bp "github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/bifrost/util/scrc"
	lt "github.com/libp2p/go-libp2p-transport"
)

// Link wraps a libp2p connection with a smux link.
type Link struct {
	conn          lt.Conn
	uuid          uint64
	transportUUID uint64
}

// NewLink builds a new link.
func NewLink(transportUUID uint64, conn lt.Conn) *Link {
	uuid := scrc.Crc64(
		[]byte(conn.LocalMultiaddr().String()),
		[]byte(conn.RemoteMultiaddr().String()),
		[]byte(conn.RemotePeer().Pretty()),
	)
	return &Link{conn: conn, uuid: uuid, transportUUID: transportUUID}
}

// GetUUID returns the host-unique ID.
// This should be repeatable between re-constructions of the same link.
func (l *Link) GetUUID() uint64 {
	return l.uuid
}

// GetTransportUUID returns the unique ID of the transport.
func (l *Link) GetTransportUUID() uint64 {
	return l.transportUUID
}

// OpenStream opens a stream on the link, with the given parameters.
func (l *Link) OpenStream(opts stream.OpenOpts) (stream.Stream, error) {
	// All libp2p transports are reliable + encrypted
	return l.conn.OpenStream()
}

// GetRemotePeer returns the identity of the remote peer.
func (l *Link) GetRemotePeer() bp.ID {
	return bp.ID(l.conn.RemotePeer())
}

// Close closes the link.
// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
func (l *Link) Close() error {
	return l.conn.Close()
}

// _ is a type assertion
var _ link.Link = ((*Link)(nil))
