package link

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/golang/protobuf/proto"
	"time"
)

// Link represents a one-hop connection between two peers.
type Link interface {
	// OpenStream opens a stream on the link, with the given parameters.
	OpenStream(encrypted, reliable bool) (Stream, error)
	// GetRemotePeer returns the identity of the remote peer.
	GetRemotePeer() peer.ID
	// GetInfo returns information about the link.
	// TODO: Returns nil if the link is lost?
	GetInfo() *LinkInfo
	// Close closes the connection.
	// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
	Close() error
}

// Stream is a stream-based data channel between two peers over a link.
type Stream interface {
	// ID returns the unique stream ID.
	ID() uint32
	// Read data from the stream.
	Read(b []byte) (n int, err error)
	// Write data to the stream.
	Write(b []byte) (n int, err error)
	// SetReadDeadline sets the read deadline as defined by
	// net.Conn.SetReadDeadline.
	// A zero time value disables the deadline.
	SetReadDeadline(t time.Time) error
	// SetWriteDeadline sets the write deadline as defined by
	// net.Conn.SetWriteDeadline.
	// A zero time value disables the deadline.
	SetWriteDeadline(t time.Time) error
	// SetDeadline sets both read and write deadlines as defined by
	// net.Conn.SetDeadline.
	// A zero time value disables the deadlines.
	SetDeadline(t time.Time) error
	// Close implements net.Conn
	Close() error
}

// NewLinkInfo builds new link information.
func NewLinkInfo(
	uuid, transportUUID uint64,
	routable bool,
	remotePeerID peer.ID,
	innerData proto.Message,
) *LinkInfo {
	d, _ := proto.Marshal(innerData)
	return &LinkInfo{
		Uuid:          uuid,
		TransportUuid: transportUUID,
		Routable:      routable,
		RemotePeerId:  []byte(remotePeerID),
		Data:          d,
	}
}
