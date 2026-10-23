package rwc

import (
	"io"
	"net"
	"time"

	"github.com/pkg/errors"
)

// RwcPacketConn wraps a ReadWriteCloser to implement net.PacketConn.
//
// Read and write deadlines are not implemented.
// The addresses are ignored when reading / writing.
type RwcPacketConn struct {
	rwc   io.ReadWriteCloser
	laddr net.Addr
	raddr net.Addr
}

// NewRwcPacketConn creates a new RwcPacketConn with the provided ReadWriteCloser and local/remote addresses.
func NewRwcPacketConn(rwc io.ReadWriteCloser, laddr, raddr net.Addr) *RwcPacketConn {
	return &RwcPacketConn{rwc: rwc, laddr: laddr, raddr: raddr}
}

// ReadFrom reads a packet from the connection, ignoring the provided address.
//
// Read errors are wrapped so they are not a net.Error: the wrapper has no
// deadlines, so no read error clears by retrying. errors.Is still finds io.EOF.
func (pc *RwcPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = pc.rwc.Read(p)
	if err != nil {
		err = errors.Wrap(err, "read packet")
	}
	return n, pc.raddr, err
}

// WriteTo writes a packet to the connection, ignoring the provided address.
func (pc *RwcPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return pc.rwc.Write(p)
}

// Close closes the connection.
func (pc *RwcPacketConn) Close() error {
	return pc.rwc.Close()
}

// LocalAddr returns the local network address.
func (pc *RwcPacketConn) LocalAddr() net.Addr {
	return pc.laddr
}

// SetDeadline sets the read and write deadlines associated with the connection.
// It is a no-op: RwcPacketConn does not support deadlines.
func (pc *RwcPacketConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls.
// It is a no-op: RwcPacketConn does not support deadlines.
func (pc *RwcPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
// It is a no-op: RwcPacketConn does not support deadlines.
func (pc *RwcPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// _ is a type assertion
var _ net.PacketConn = (*RwcPacketConn)(nil)
