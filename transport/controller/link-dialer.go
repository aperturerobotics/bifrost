package transport_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/transport/common/dialer"
)

// linkDialerKey is the peer ID and link address tuple.
type linkDialerKey struct {
	peerID      string
	dialAddress string
}

// linkDialer is a link dialer instance.
type linkDialer struct {
	// dialer is the dialer
	dialer *dialer.Dialer
	// cancel cancels the dialer
	cancel context.CancelFunc
}

// executeDialer executes the link dialer.
func (c *Controller) executeDialer(
	ctx context.Context,
	key linkDialerKey,
	ld *linkDialer,
) {
	err := ld.dialer.Execute(ctx)
	if err != nil && err != context.Canceled {
		c.le.WithError(err).Warn("dialer exited with error")
	}
	ld.cancel()

	c.mtx.Lock()
	if d := c.linkDialers[key]; d == ld {
		delete(c.linkDialers, key)
	}
	c.mtx.Unlock()
}
