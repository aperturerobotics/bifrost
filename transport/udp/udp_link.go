package udp

import (
	"errors"
	"net"
	"time"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/peer"
	// 	"github.com/xtaci/kcp-go"
)

// Link represents a UDP-based connection/link.
type Link struct {
	// pc is the packet connection
	pc net.PacketConn
	// addr is the bound remote address
	addr *net.UDPAddr
	// neg is the negotiated identity data
	neg *identity.Result
	// sharedSecret is the shared secret
	sharedSecret [32]byte
}

// NewLink builds a new link.
func NewLink(pc net.PacketConn, addr *net.UDPAddr, neg *identity.Result, sharedSecret [32]byte) *Link {
	return &Link{pc: pc, addr: addr, neg: neg, sharedSecret: sharedSecret}
}

// LocalAddr returns the local network address.
func (l *Link) LocalAddr() net.Addr {
	return l.pc.LocalAddr()
}

// Read reads a packet from the connection, copying the payload into b.
// It returns the number of bytes copied into b. ReadFrom can be made to
// time out and return an Error with Timeout() == true after a fixed time
// limit; see SetDeadline and SetReadDeadline.
// Returns io.EOF when the connection is closed.
func (l *Link) Read(b []byte) (n int, err error) {
	for {
		var addr net.Addr
		n, addr, err = l.pc.ReadFrom(b)
		if err != nil {
			return
		}

		ua, ok := addr.(*net.UDPAddr)
		if !ok {
			return 0, errors.New("expected udp addr, got other type")
		}

		// skip if the address is different
		if !ua.IP.Equal(l.addr.IP) {
			continue
		}

		return
	}
}

// Write writes a packet with payload b. an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline. On
// packet-oriented connections, write timeouts are rare. The size of the
// packet must be less than the MTU.
func (l *Link) Write(b []byte) (n int, err error) {
	return l.pc.WriteTo(b, l.addr)
}

// GetMTU returns the maximum-transmission-unit (max packet size).
// If 0, there is no packet size limited (indicate stream-oriented).
func (l *Link) GetMTU() int {
	return 1400 // TODO: is it possible to detect from the NIC?
}

// GetIsReliable returns if this link is ordered and reliable. A link is
// reliable if it provides both message ordering and message delivery
// guarantees.
func (l *Link) GetIsReliable() bool {
	return false
}

// GetRemotePeer returns the identity of the remote peer if encrypted.
// Returns a zero value if the connection is not pre-identified.
func (l *Link) GetRemotePeer() peer.ID {
	return peer.ID("")
}

// Close closes the connection.
// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
func (l *Link) Close() error {
	return l.pc.Close()
}

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
func (l *Link) SetDeadline(t time.Time) error {
	return l.pc.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future ReadFrom calls and any
// currently-blocked ReadFrom call. A zero value for t means ReadFrom will
// not time out.
func (l *Link) SetReadDeadline(t time.Time) error {
	return l.pc.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future WriteTo calls and any
// currently-blocked WriteTo call. Even if write times out, it may return n
// > 0, indicating that some of the data was successfully written. A zero
// value for t means WriteTo will not time out.
func (l *Link) SetWriteDeadline(t time.Time) error {
	return l.pc.SetWriteDeadline(t)
}
