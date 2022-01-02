package transport_controller

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/directive"
)

// crossDialWaitDur is the default amount of time to wait to avoid cross-dial.
var crossDialWaitDur = time.Second

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
	peerIDConst := o.dir.EstablishLinkTargetPeerId()
	peerIDPretty := peerIDConst.Pretty()

	wakeDialer := make(chan time.Time, 1)
	wakeDialer <- time.Now()

	linkIDs := make(map[link.Link]uint32)
	c.mtx.Lock()
	lw := o.c.pushLinkWaiter(peerIDConst, false, func(lnk link.Link, added bool) {
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

	defer func() {
		c.mtx.Lock()
		o.c.clearLinkWaiter(lw)
		c.mtx.Unlock()
	}()

	for {
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

		c.mtx.Lock()
		// Check the Static Peer Map for a address, push a dialer if exists.
		if spm := c.staticPeerMap; spm != nil && len(spm) > 0 {
			var hasLink bool
			for _, lnk := range c.links {
				if lnk.Link.GetRemotePeer() == peerIDConst {
					hasLink = true
					break
				}
			}
			// Skip pushing dialer if a link already exists.
			if !hasLink {
				if dOpts, ok := spm[peerIDPretty]; ok && dOpts.GetAddress() != "" {
					go func() {
						_ = c.PushDialer(ctx, peerIDConst, dOpts)
					}()
				}
			}
		}
		c.mtx.Unlock()
	}
}

// resolveEstablishLink returns a resolver for opening a stream.
// Negotiates the protocol ID as well.
func (c *Controller) resolveEstablishLink(
	ctx context.Context,
	di directive.Instance,
	dir link.EstablishLinkWithPeer,
) (directive.Resolver, error) {
	if len(dir.EstablishLinkTargetPeerId()) == 0 {
		return nil, nil
	}

	// Return resolver.
	return &establishLinkResolver{c: c, ctx: ctx, di: di, dir: dir}, nil
}

// _ is a type assertion
var _ directive.Resolver = ((*establishLinkResolver)(nil))
