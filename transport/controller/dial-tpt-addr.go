package transport_controller

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/tptaddr"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/controllerbus/directive"
)

// dialTptAddrResolver resolves DialTptAddr directives
type dialTptAddrResolver struct {
	c   *Controller
	ctx context.Context
	di  directive.Instance
	dir tptaddr.DialTptAddr
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (o *dialTptAddrResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	tpt, err := o.c.GetTransport(ctx)
	if err != nil {
		return err
	}
	tptDialer, ok := tpt.(dialer.TransportDialer)
	if !ok {
		return nil
	}
	transportID, dialAddr, err := tptaddr.ParseTptAddr(o.dir.DialTptAddr())
	if err != nil {
		return nil
	}
	if !tptDialer.MatchTransportType(transportID) {
		return nil
	}

	c := o.c
	peerIDConst := o.dir.DialTptAddrTargetPeerId()

	wakeDialer := make(chan time.Time, 1)
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

		c.mtx.Lock()
		// If we already have a link, do nothing.
		var hasLink bool
		for _, lnk := range c.links {
			if lnk.lnk.GetRemotePeer() == peerIDConst {
				hasLink = true
				break
			}
		}
		if !hasLink {
			err = c.startLinkDialerLocked(peerIDConst, &dialer.DialerOpts{
				Address: dialAddr,
			}, tptDialer)
		}
		c.mtx.Unlock()
		if err != nil {
			return err
		}
	}
}

// resolveDialTptAddr returns a resolver for dialing a transport address.
func (c *Controller) resolveDialTptAddr(
	ctx context.Context,
	di directive.Instance,
	dir tptaddr.DialTptAddr,
) ([]directive.Resolver, error) {
	if len(dir.DialTptAddrTargetPeerId()) == 0 || dir.DialTptAddr() == "" {
		return nil, nil
	}

	// Return resolver.
	return directive.Resolvers(&dialTptAddrResolver{
		c:   c,
		ctx: ctx,
		di:  di,
		dir: dir,
	}), nil
}

// _ is a type assertion
var _ directive.Resolver = ((*dialTptAddrResolver)(nil))
