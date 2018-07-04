package kcpl

import (
	"net"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/pkg/errors"
)

// Link represents a KCP-based connection/link.
type Link struct {
	// pc is the packet connection
	pc net.PacketConn
	// addr is the bound remote address
	addr *net.UDPAddr
	// neg is the negotiated identity data
	neg *identity.Result
	// sharedSecret is the shared secret
	sharedSecret [32]byte
	// peerID is the remote peer id
	peerID peer.ID
	// uuid is the unique ID of the link
	uuid uint64
}

// NewLink builds a new link.
func NewLink(
	pc net.PacketConn,
	transportUUID uint64,
	addr *net.UDPAddr,
	neg *identity.Result,
	sharedSecret [32]byte,
) *Link {
	pid, _ := peer.IDFromPublicKey(neg.Peer)
	uuid := scrc.Crc64(
		[]byte("udp"),
		[]byte(pc.LocalAddr().String()),
		[]byte(addr.String()),
		[]byte(pid),
	)
	return &Link{
		pc:           pc,
		addr:         addr,
		neg:          neg,
		sharedSecret: sharedSecret,
		peerID:       pid,
		uuid:         uuid,
	}
}

// GetUUID returns the host-unique ID of the link.
func (l *Link) GetUUID() uint64 {
	return l.uuid
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
