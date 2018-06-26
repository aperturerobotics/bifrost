package pconn

import (
	"context"
	"net"

	"github.com/aperturerobotics/bifrost/handshake/identity"
)

// inflightHandshake is an on-going handshake.
type inflightHandshake struct {
	ctxCancel context.CancelFunc
	hs        identity.Handshaker
	addr      net.Addr
}
