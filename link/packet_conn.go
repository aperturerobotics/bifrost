package link

import (
	"net"
)

// PacketConn wraps a Link to make it net.PacketConn compatible.
type PacketConn struct {
	// Link is the underlying link.
	Link
	// getLocalAddr gets the local address.
	getLocalAddr func() net.Addr
}

// NewPacketConn builds a new packet conn wrapper.
func NewPacketConn(lnk Link) *PacketConn {
	type localAddrIdentifier interface {
		LocalAddr() net.Addr
	}

	pc := &PacketConn{Link: lnk}
	if lai, ok := lnk.(localAddrIdentifier); ok {
		pc.getLocalAddr = lai.LocalAddr
	} else {
		pc.getLocalAddr = func() net.Addr {
			return nil
		}
	}

	return pc
}

// ReadFrom reads a packet from the connection,
// copying the payload into b. It returns the number of
// bytes copied into b and the return address that
// was on the packet.
// ReadFrom can be made to time out and return
// an Error with Timeout() == true after a fixed time limit;
// see SetDeadline and SetReadDeadline.
func (p *PacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, err = p.Link.Read(b)
	addr = p.LocalAddr()
	return
}

// WriteTo writes a packet with payload b to addr.
// WriteTo can be made to time out and return
// an Error with Timeout() == true after a fixed time limit;
// see SetDeadline and SetWriteDeadline.
// On packet-oriented connections, write timeouts are rare.
func (p *PacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return p.Link.Write(b)
}

// LocalAddr returns the local network address.
// A dummy address is used if necessary.
func (p *PacketConn) LocalAddr() net.Addr {
	return p.getLocalAddr()
}
