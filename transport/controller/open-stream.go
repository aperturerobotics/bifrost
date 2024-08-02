package transport_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/controllerbus/directive"
)

// openStreamResolver resolves OpenStream directives
type openStreamResolver struct {
	c   *Controller
	ctx context.Context
	di  directive.Instance
	dir link.OpenStreamWithPeer
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (o *openStreamResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	c := o.c
	openOpts := o.dir.OpenStreamWPOpenOpts()
	protocolID := o.dir.OpenStreamWPProtocolID()
	tgtPeerID := o.dir.OpenStreamWPTargetPeerID()

	tpt, err := o.c.GetTransport(ctx)
	if err != nil {
		return err
	}

	if !checkOpenStreamMatchesTpt(o.dir, tpt) {
		return nil
	}

	var lnk link.Link
	for {
		err = c.bcast.Wait(ctx, func(broadcast func(), getWaitCh func() <-chan struct{}) (bool, error) {
			if c.execCtx == nil {
				return false, nil
			}

			peerLinks := c.linksByPeerID[tgtPeerID]
			for _, peerLink := range peerLinks {
				// if lnk == peerLink.lnk, that link already failed.
				// flush it and continue
				if lnk != nil && peerLink.lnk == lnk {
					c.flushEstablishedLink(peerLink, false)
					continue
				}

				lnk = peerLinks[0].lnk
				return true, nil
			}

			// no matching link yet
			return false, nil
		})
		if ctx.Err() != nil {
			return context.Canceled
		}
		if err != nil {
			return err
		}

		mstrm, err := c.openStreamWithLink(lnk, openOpts, protocolID)
		if ctx.Err() != nil {
			if mstrm != nil {
				_ = mstrm.GetStream().Close()
			}
			return context.Canceled
		}
		if err != nil {
			c.loggerForLink(lnk).WithError(err).Warn("unable to open stream, closing link")
			_ = lnk.Close()
			continue
		}

		// success
		if _, accepted := handler.AddValue(mstrm); !accepted {
			_ = mstrm.GetStream().Close()
		}
		return nil
	}
}

// checkOpenStreamMatchesTpt checks if a OpenStream matches a tpt
func checkOpenStreamMatchesTpt(dir link.OpenStreamWithPeer, tpt transport.Transport) bool {
	if tptConstraint := dir.OpenStreamWPTransportConstraint(); tptConstraint != 0 {
		if tpt.GetUUID() != tptConstraint {
			return false
		}
	}

	// Check peer ID constraint
	if srcPeerID := dir.OpenStreamWPSourcePeerID(); len(srcPeerID) != 0 {
		if srcPeerID != tpt.GetPeerID() {
			return false
		}
	}

	return true
}

// resolveOpenStreamWithPeer returns a resolver for opening a stream.
// Negotiates the protocol ID as well.
func (c *Controller) resolveOpenStreamWithPeer(
	ctx context.Context,
	di directive.Instance,
	dir link.OpenStreamWithPeer,
) ([]directive.Resolver, error) {
	// Try to skip this if it doesn't match this transport and we can lock.
	// Otherwise return a resolver and check later.
	var skip bool
	_ = c.bcast.TryHoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
		skip = c.tpt != nil && !checkOpenStreamMatchesTpt(dir, c.tpt)
	})
	if skip {
		return nil, nil
	}

	return directive.R(&openStreamResolver{
		c:   c,
		ctx: ctx,
		di:  di,
		dir: dir,
	}, nil)
}

// _ is a type assertion
var _ directive.Resolver = ((*openStreamResolver)(nil))
