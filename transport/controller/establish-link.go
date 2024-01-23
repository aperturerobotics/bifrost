package transport_controller

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/controllerbus/directive"
)

// crossDialWaitDur is the default amount of time to wait to avoid cross-dial.
var crossDialWaitDur = time.Millisecond * 250

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
	targetPeerID := o.dir.EstablishLinkTargetPeerId()
	tpt, err := o.c.GetTransport(ctx)
	if err != nil {
		return err
	}

	// If the directive filtered by source peer ID, add a EstablishLink
	// directive for the destination as well to ensure there is a reference.
	//
	// XXX: the current approach to using EstablishLink to determine the
	// lifecycle of the Link (when it is released) is a bit messy and
	// counter-intuitive and could be improved by instead using a refcount
	// mechanism to release links when no directives ask for them.
	sourcePeerID := tpt.GetPeerID()
	if o.dir.EstablishLinkSourcePeerId() != sourcePeerID {
		_, estLinkRef, err := o.c.bus.AddDirective(link.NewEstablishLinkWithPeer(sourcePeerID, targetPeerID), nil)
		if err != nil {
			return err
		}
		defer estLinkRef.Release()
	}

	wakeDialer := make(chan time.Time, 1)
	linkIDs := make(map[link.Link]uint32)
	c.mtx.Lock()
	lw := o.c.pushLinkWaiter(targetPeerID, false, func(lnk link.Link, added bool) {
		if added {
			if _, ok := linkIDs[lnk]; !ok {
				if vid, ok := handler.AddValue(lnk); ok {
					linkIDs[lnk] = vid
				}
			}
		} else {
			if vid, ok := linkIDs[lnk]; ok {
				handler.RemoveValue(vid)
				delete(linkIDs, lnk)
				if len(linkIDs) == 0 {
					var extraWaitDur time.Duration
					if lnk.GetLocalPeer() < lnk.GetRemotePeer() {
						extraWaitDur = crossDialWaitDur
					}
					select {
					case wakeDialer <- time.Now().Add(extraWaitDur):
					default:
					}
				}
			}
		}
	})
	c.mtx.Unlock()

	// Remove the link waiter when the resolver exits.
	defer func() {
		c.mtx.Lock()
		o.c.clearLinkWaiter(lw)
		c.mtx.Unlock()
	}()

	// Attempt to dial the peer if no link is active and we have an address to dial.
	tptDialer, ok := tpt.(dialer.TransportDialer)
	if !ok {
		// No transport dialer, just wait for a link.
		<-ctx.Done()
		return nil
	}

	dialerOpts, err := tptDialer.GetPeerDialer(ctx, targetPeerID)
	if err != nil {
		return err
	}

	if dialerOpts.GetAddress() == "" {
		// No address, wait for a link.
		<-ctx.Done()
		return nil
	}

	var waitWake bool
	for {
		if waitWake {
			select {
			case <-ctx.Done():
				return nil
			case waitUntil := <-wakeDialer:
				tu := time.Until(waitUntil)
				if tu > time.Millisecond*50 {
					tt := time.NewTimer(tu)
					select {
					case <-ctx.Done():
						tt.Stop()
						return nil
					case <-tt.C:
					}
				}
			}
		} else {
			waitWake = true
		}

		var hasLink bool
		for _, lnk := range c.links {
			if lnk.lnk.GetRemotePeer() == targetPeerID {
				hasLink = true
				break
			}
		}
		if !hasLink {
			if err := c.PushDialer(ctx, targetPeerID, dialerOpts); err != nil {
				return err
			}
		}
	}
}

// resolveEstablishLink returns a resolver for opening a stream.
// Negotiates the protocol ID as well.
func (c *Controller) resolveEstablishLink(
	ctx context.Context,
	di directive.Instance,
	dir link.EstablishLinkWithPeer,
) ([]directive.Resolver, error) {
	if len(dir.EstablishLinkTargetPeerId()) == 0 {
		return nil, nil
	}

	// Return resolver.
	return directive.Resolvers(&establishLinkResolver{
		c:   c,
		ctx: ctx,
		di:  di,
		dir: dir,
	}), nil
}

// _ is a type assertion
var _ directive.Resolver = ((*establishLinkResolver)(nil))
