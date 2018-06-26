package pconn

import (
	"context"
	"io"
	"net"
	"time"
)

const dummyKcpRemoteStr = "127.0.0.1:1000"

var dummyKcpRemoteAddr net.Addr

func init() {
	dummyKcpRemoteAddr, _ = net.ResolveUDPAddr("udp", dummyKcpRemoteStr)
}

// timeoutError is returned for any timeouts.
type timeoutError struct {
	error
	temp bool
}

// Timeout returns if this is a timeout or not.
func (c *timeoutError) Timeout() bool {
	return true
}

// Temporary returns if this is a temporary error or not.
func (c *timeoutError) Temporary() bool {
	return c.temp
}

// kcpPacketConn implements net.PacketConn for the KCP session implementation
type kcpPacketConn struct {
	ctx    context.Context
	readCh chan []byte

	localAddr, remoteAddr net.Addr

	writer func(b []byte, addr net.Addr) (n int, err error)
	closer func() error
}

func newKcpPacketConn(
	ctx context.Context,
	localAddr, remoteAddr net.Addr,
	writer func(b []byte, addr net.Addr) (n int, err error),
	closer func() error,
) *kcpPacketConn {
	return &kcpPacketConn{
		ctx:        ctx,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
		writer:     writer,
		closer:     closer,
		readCh:     make(chan []byte),
	}
}

// pushPacket pushes an incoming packet to the read channel.
func (c *kcpPacketConn) pushPacket(dat []byte) {
	select {
	case <-c.ctx.Done():
	case c.readCh <- dat:
	}
}

// ReadFrom reads a packet from the connection,
// copying the payload into b. It returns the number of
// bytes copied into b and the return address that
// was on the packet.
// ReadFrom can be made to time out and return
// an Error with Timeout() == true after a fixed time limit;
// see SetDeadline and SetReadDeadline.
func (c *kcpPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	// Read from channel
	select {
	case <-c.ctx.Done():
		return 0, nil, c.ctx.Err()
	case buf := <-c.readCh:
		if len(buf) > len(b) {
			return 0, nil, io.ErrShortBuffer
		}

		copy(b, buf)
		return len(buf), dummyKcpRemoteAddr, nil
	}
}

// WriteTo writes a packet with payload b to addr.
// WriteTo can be made to time out and return
// an Error with Timeout() == true after a fixed time limit;
// see SetDeadline and SetWriteDeadline.
// On packet-oriented connections, write timeouts are rare.
func (c *kcpPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return c.writer(b, c.remoteAddr)
}

// LocalAddr returns the local network address.
func (c *kcpPacketConn) LocalAddr() net.Addr {
	return c.localAddr
}

// Close closes the connection.
// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
func (c *kcpPacketConn) Close() error {
	return c.closer()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to ReadFrom or
// WriteTo. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful ReadFrom or WriteTo calls.
//
// A zero value for t means I/O operations will not time out.
func (c *kcpPacketConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future ReadFrom calls
// and any currently-blocked ReadFrom call.
// A zero value for t means ReadFrom will not time out.
func (c *kcpPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future WriteTo calls
// and any currently-blocked WriteTo call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means WriteTo will not time out.
func (c *kcpPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// _ is a type assertion
var _ net.PacketConn = ((*kcpPacketConn)(nil))
