package transport_quic

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
)

// Dialer represents a ongoing attempt to dial an address
type Dialer struct {
	// t is the transport
	t *Transport
	// rootCtx is the root context
	rootCtx context.Context
	// ctx is the dialer ctx
	ctx context.Context
	// ctxCancel cancels ctx
	// called when execute() exits
	ctxCancel context.CancelFunc
	// peerID is the peer id
	// can be empty to indicate any peer
	peerID peer.ID
	// addr is the address
	addr string
}

// NewDialer constructs a new dialer.
func NewDialer(
	rctx context.Context,
	t *Transport,
	peerID peer.ID,
	addr string,
) (*Dialer, error) {
	d := &Dialer{
		t:       t,
		rootCtx: rctx,
		peerID:  peerID,
		addr:    addr,
	}
	d.ctx, d.ctxCancel = context.WithCancel(rctx)
	return d, nil
}

// Execute executes the dialer, yielding a Link.
func (d *Dialer) Execute() (*Link, error) {
	ctx := d.ctx
	defer d.ctxCancel()

	d.t.le.Debugf("quic dialing peer address: %s", d.addr)
	pc, raddr, err := d.t.dialFn(ctx, d.addr)
	if err != nil {
		return nil, err
	}
	return d.t.HandleConn(ctx, true, pc, raddr, d.peerID)
}
