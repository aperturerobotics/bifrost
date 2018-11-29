package transport_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/directive"
)

// establishLinkResolver resolves establishLink directives
type establishLinkResolver struct {
	c   *Controller
	ctx context.Context
	di  directive.Instance
	dir link.EstablishLinkWithPeer
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (o *establishLinkResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	c := o.c
	peerIDConst := o.dir.EstablishLinkWPIDConstraint()
	peerIDPretty := peerIDConst.Pretty()

	if spm := c.staticPeerMap; spm != nil {
		if dOpts, ok := spm[peerIDPretty]; ok && dOpts.GetAddress() != "" {
			go c.PushDialer(ctx, peerIDConst, dOpts)
		}
	}

	linkIDs := make(map[link.Link]uint32)
	c.linksMtx.Lock()
	lw := o.c.pushLinkWaiter(peerIDConst, false, func(lnk link.Link, added bool) {
		if added {
			if vid, ok := handler.AddValue(lnk); ok {
				linkIDs[lnk] = vid
			}
		} else {
			if vid, ok := linkIDs[lnk]; ok {
				handler.RemoveValue(vid)
				delete(linkIDs, lnk)
			}
		}
	})
	c.linksMtx.Unlock()

	defer func() {
		c.linksMtx.Lock()
		o.c.clearLinkWaiter(lw)
		c.linksMtx.Unlock()
	}()

	<-ctx.Done()
	return nil
}

// resolveEstablishLink returns a resolver for opening a stream.
// Negotiates the protocol ID as well.
func (c *Controller) resolveEstablishLink(
	ctx context.Context,
	di directive.Instance,
	dir link.EstablishLinkWithPeer,
) (directive.Resolver, error) {
	if dir.EstablishLinkWPIDConstraint() == peer.ID("") {
		return nil, nil
	}

	// Return resolver.
	return &establishLinkResolver{c: c, ctx: ctx, di: di, dir: dir}, nil
}

// _ is a type assertion
var _ directive.Resolver = ((*establishLinkResolver)(nil))
