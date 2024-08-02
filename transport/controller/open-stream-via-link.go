package transport_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/directive"
)

// openStreamViaLinkResolver resolves OpenStreamViaLink directives
type openStreamViaLinkResolver struct {
	c   *Controller
	ctx context.Context
	di  directive.Instance
	dir link.OpenStreamViaLink
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (o *openStreamViaLinkResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	c := o.c
	lnkUUID := o.dir.OpenStreamViaLinkUUID()
	openOpts := o.dir.OpenStreamViaLinkOpenOpts()
	protocolID := o.dir.OpenStreamViaLinkProtocolID()
	tptID := o.dir.OpenStreamViaLinkTransportConstraint()

	tpt, err := o.c.GetTransport(ctx)
	if err != nil {
		return err
	}

	if tptID != 0 && tpt.GetUUID() != tptID {
		return nil
	}

	var lnk link.Link
	for {
		err = c.bcast.Wait(ctx, func(broadcast func(), getWaitCh func() <-chan struct{}) (bool, error) {
			if c.execCtx == nil {
				return false, nil
			}

			for _, el := range c.links {
				if el.lnk.GetUUID() == lnkUUID {
					lnk = el.lnk
					return true, nil
				}
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

// resolveOpenStreamViaLink returns a resolver for opening a stream.
// Negotiates the protocol ID as well.
func (c *Controller) resolveOpenStreamViaLink(
	ctx context.Context,
	di directive.Instance,
	dir link.OpenStreamViaLink,
) ([]directive.Resolver, error) {
	// Try to skip this if it doesn't match this transport and we can lock.
	// Otherwise return a resolver and check later.
	tptID := dir.OpenStreamViaLinkTransportConstraint()
	if tptID != 0 {
		var skip bool
		_ = c.bcast.TryHoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
			skip = c.tpt != nil && c.tpt.GetUUID() != tptID
		})
		if skip {
			return nil, nil
		}
	}

	// Check transport constraint
	// Return resolver.
	return directive.Resolvers(&openStreamViaLinkResolver{
		c:   c,
		ctx: ctx,
		di:  di,
		dir: dir,
	}), nil
}

// _ is a type assertion
var _ directive.Resolver = ((*openStreamViaLinkResolver)(nil))
