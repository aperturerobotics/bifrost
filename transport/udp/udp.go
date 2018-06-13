package udp

import (
	"context"
	"net"

	"github.com/aperturerobotics/bifrost/transport"
)

// UDP implements a UDP transport.
// It is unordered, unreliable, and unencrypted.
type UDP struct {
	// pc is the packet conn
	pc net.PacketConn
}

// NewUDP builds a new UDP transport, listening on the addr.
func NewUDP(listenAddr string) (*UDP, error) {
	pc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	return NewUDPFromPacketConn(pc), nil
}

// NewUDPFromPacketConn builds a new UDP listener from a packet conn.
func NewUDPFromPacketConn(pc net.PacketConn) *UDP {
	return &UDP{pc: pc}
}

// Execute processes the transport, emitting events to the handler.
// Fatal errors are returned.
func (u *UDP) Execute(ctx context.Context, handler transport.Handler) error {
	// TODO
	return nil
}

// _ is a type assertion.
var _ transport.Transport = ((*UDP)(nil))
