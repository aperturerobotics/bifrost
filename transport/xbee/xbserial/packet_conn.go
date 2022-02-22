package xbserial

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

// XBeeAddr is the xbee address.
type XBeeAddr uint64

// Network is the name of the network (tcp uses "tcp")
func (a XBeeAddr) Network() string {
	return "xbee64"
}

// string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
func (a XBeeAddr) String() string {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(a))
	return hex.EncodeToString(data)
}

// ParseXBeeAddr parses a hex xbee address.
func ParseXBeeAddr(adr string) (XBeeAddr, error) {
	data, err := hex.DecodeString(adr)
	if err != nil {
		return 0, err
	}

	if len(data) != 8 {
		return 0, errors.New("xbee hex address invalid")
	}

	return XBeeAddr(binary.BigEndian.Uint64(data)), nil
}

// _ is a type assertion
var _ net.Addr = ((XBeeAddr)(0))

// PacketConn wraps a xbee serial connection into a packet conn.
// Assumes only one reader at a time to eliminate a mutex lock
type PacketConn struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	xb        *XBeeSerial

	localAddr XBeeAddr

	pqMtx          sync.Mutex
	pqHead, pqTail *incPacket
	pqWake         chan struct{}

	rd time.Time // read deadline
	wd time.Time // write deadline
}

// NewPacketConn creates a PacketConn by looking up the local address.
func NewPacketConn(ctx context.Context, xb *XBeeSerial) (*PacketConn, error) {
	laddr, err := xb.ReadLocalAddress(ctx)
	if err != nil {
		return nil, err
	}

	return NewPacketConnWithAddr(ctx, xb, laddr), nil
}

// NewPacketConnWithAddr constructs a PacketConn with addresses.
func NewPacketConnWithAddr(ctx context.Context, xb *XBeeSerial, localAddr XBeeAddr) *PacketConn {
	pc := &PacketConn{xb: xb, localAddr: localAddr, pqWake: make(chan struct{})}
	pc.ctx, pc.ctxCancel = context.WithCancel(ctx)
	xb.SetDataHandler(pc.handleIncomingData)
	return pc
}

// ReadFrom reads a packet from the connection, copying the payload into p.
// Respects deadlines.
// If Close() is called, return an error.
func (c *PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	deadline := c.rd
	ctx := c.ctx
	if !deadline.IsZero() {
		var ctxCancel context.CancelFunc
		ctx, ctxCancel = context.WithDeadline(c.ctx, deadline)
		defer ctxCancel()
	}

	var pkt *incPacket
	for {
		c.pqMtx.Lock()
		if c.pqHead != nil {
			pkt = c.pqHead
			c.pqHead = pkt.next
			if c.pqHead == nil {
				c.pqTail = nil
			}
		}
		c.pqMtx.Unlock()

		if pkt != nil {
			copy(p, pkt.data)
			if len(p) < len(pkt.data) {
				return len(p), net.Addr(pkt.addr), io.ErrShortBuffer
			}
			return len(pkt.data), net.Addr(pkt.addr), nil
		}

		select {
		case <-ctx.Done():
			return 0, nil, ctx.Err()
		case <-c.pqWake:
		}
	}
}

// WriteTo writes a packet with payload p to addr.
// Addr is expected to be a XBeeAddr (will return an error otherwise)
// WriteTo can be made to time out and return
// an Error with Timeout() == true after a fixed time limit;
// see SetDeadline and SetWriteDeadline.
// On packet-oriented connections, write timeouts are rare.
func (c *PacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	xbAddr, ok := addr.(XBeeAddr)
	if !ok || xbAddr == 0 {
		return 0, errors.New("address is not valid")
	}

	// no retries + extended transmit timeout
	var options byte = 0x01 | 0x40
	if err := c.xb.TxToAddr(c.ctx, uint64(xbAddr), 0, 0, 0, 0, 0, 0, options, p); err != nil {
		return 0, err
	}

	return len(p), nil
}

// LocalAddr returns the local network address.
func (c *PacketConn) LocalAddr() net.Addr {
	return net.Addr(c.localAddr)
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
func (c *PacketConn) SetDeadline(t time.Time) error {
	_ = c.SetReadDeadline(t)
	_ = c.SetWriteDeadline(t)
	return nil
}

// SetReadDeadline sets the deadline for future ReadFrom calls
// and any currently-blocked ReadFrom call.
// A zero value for t means ReadFrom will not time out.
func (c *PacketConn) SetReadDeadline(t time.Time) error {
	c.rd = t
	return nil
}

// SetWriteDeadline sets the deadline for future WriteTo calls
// and any currently-blocked WriteTo call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means WriteTo will not time out.
func (c *PacketConn) SetWriteDeadline(t time.Time) error {
	c.wd = t
	return nil
}

// Close closes the connection.
// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
func (c *PacketConn) Close() error {
	c.ctxCancel()
	return nil
}

// incPacket is an incoming packet.
type incPacket struct {
	addr XBeeAddr
	data []byte
	next *incPacket
}

// handleIncomingData handles an incoming packet.
func (c *PacketConn) handleIncomingData(data []byte, addr XBeeAddr) {
	t := &incPacket{
		addr: addr,
		data: data,
	}
	c.pqMtx.Lock()
	if c.pqTail == nil {
		c.pqTail = t
		c.pqHead = t
	} else {
		c.pqTail.next = t
		c.pqTail = t
	}
	c.pqMtx.Unlock()

	select {
	case c.pqWake <- struct{}{}:
	default:
	}
}

// _ is type assertion
var _ net.PacketConn = ((*PacketConn)(nil))
