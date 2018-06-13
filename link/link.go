package link

import (
	"time"

	"github.com/aperturerobotics/bifrost/peer"
)

// Link represents a one-hop connection between two peers.
// It is similar to net.PacketConn but does not have address fields.
type Link interface {
	// Read reads a packet from the connection, copying the payload into b.
	// It returns the number of bytes copied into b. ReadFrom can be made to
	// time out and return an Error with Timeout() == true after a fixed time
	// limit; see SetDeadline and SetReadDeadline.
	// Returns io.EOF when the connection is closed.
	Read(b []byte) (n int, err error)

	// Write writes a packet with payload b. an Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetWriteDeadline. On
	// packet-oriented connections, write timeouts are rare. The size of the
	// packet must be less than the MTU.
	Write(b []byte) (n int, err error)

	// GetMTU returns the maximum-transmission-unit (max packet size).
	// If 0, there is no packet size limited (indicate stream-oriented).
	GetMTU() int

	// GetIsReliable returns if this link is ordered and reliable. A link is
	// reliable if it provides both message ordering and message delivery
	// guarantees.
	GetIsReliable() bool

	// GetRemotePeer returns the identity of the remote peer if encrypted.
	// Returns a zero value if the connection is not pre-identified.
	GetRemotePeer() peer.ID

	// Close closes the connection.
	// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
	Close() error

	// SetDeadline sets the read and write deadlines associated with the
	// connection. It is equivalent to calling both SetReadDeadline and
	// SetWriteDeadline.
	//
	// A deadline is an absolute time after which I/O operations fail with a
	// timeout (see type Error) instead of blocking. The deadline applies to all
	// future and pending I/O, not just the immediately following call to
	// ReadFrom or WriteTo. After a deadline has been exceeded, the connection
	// can be refreshed by setting a deadline in the future.
	//
	// An idle timeout can be implemented by repeatedly extending the deadline
	// after successful ReadFrom or WriteTo calls.
	//
	// A zero value for t means I/O operations will not time out.
	SetDeadline(t time.Time) error

	// SetReadDeadline sets the deadline for future ReadFrom calls and any
	// currently-blocked ReadFrom call. A zero value for t means ReadFrom will
	// not time out.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the deadline for future WriteTo calls and any
	// currently-blocked WriteTo call. Even if write times out, it may return n
	// > 0, indicating that some of the data was successfully written. A zero
	// value for t means WriteTo will not time out.
	SetWriteDeadline(t time.Time) error
}
