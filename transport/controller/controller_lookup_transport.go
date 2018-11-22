package transport_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/controllerbus/directive"
)

// lookupTransportResolver resolves lookupTransport directives
type lookupTransportResolver struct {
	c   *Controller
	ctx context.Context
	dir transport.LookupTransport
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (o *lookupTransportResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	var tpt transport.Transport
	select {
	case tpt = <-o.c.tptCh:
		o.c.tptCh <- tpt
	case <-ctx.Done():
		return ctx.Err()
	}

	if !checkLookupMatchesTpt(o.dir, tpt) {
		return nil
	}

	handler.AddValue(tpt)
	return nil
}

// checkLookupMatchesTpt checks if a lookuptransport matches a tpt
func checkLookupMatchesTpt(dir transport.LookupTransport, tpt transport.Transport) bool {
	if tptIDConstraint := dir.LookupTransportIDConstraint(); tptIDConstraint != 0 {
		if tpt.GetUUID() != tptIDConstraint {
			return false
		}
	}
	if peerIDConstraint := dir.LookupTransportPeerIDConstraint(); len(peerIDConstraint) != 0 {
		if tpt.GetPeerID() != peerIDConstraint {
			return false
		}
	}

	return true
}

// resolveLookupTransport returns a resolver for looking up a transport.
func (c *Controller) resolveLookupTransport(
	ctx context.Context,
	di directive.Instance,
	dir transport.LookupTransport,
) (directive.Resolver, error) {
	select {
	case tpt := <-c.tptCh:
		c.tptCh <- tpt
		if !checkLookupMatchesTpt(dir, tpt) {
			return nil, nil
		}
	default:
	}

	// Return resolver.
	return &lookupTransportResolver{c: c, ctx: ctx, dir: dir}, nil
}

// _ is a type assertion
var _ directive.Resolver = ((*lookupTransportResolver)(nil))
