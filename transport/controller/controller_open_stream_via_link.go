package transport_controller

import (
	"context"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
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
	estMsg := NewStreamEstablish(protocolID)
	tptID := o.dir.OpenStreamViaLinkTransportConstraint()

	if tptID != 0 {
		var tpt transport.Transport
		select {
		case <-ctx.Done():
			return ctx.Err()
		case tpt = <-c.tptCh:
			c.tptCh <- tpt
		}
		if tpt.GetUUID() != tptID {
			return nil
		}
	}

	errCh := make(chan error, 1)
	strmCh := make(chan link.MountedStream, 1)
	var mtx sync.Mutex
	var done bool

	c.linksMtx.Lock()
	lw := c.pushLinkWaiter(
		peer.ID(""),
		false,
		func(lnk link.Link, added bool) {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if lnk.GetUUID() != lnkUUID || !added {
				return
			}

			mtx.Lock()
			isDone := done
			mtx.Unlock()
			if isDone {
				return
			}

			strm, err := lnk.OpenStream(openOpts)
			if err != nil {
				errCh <- err
				return
			}
			strm.SetWriteDeadline(time.Now().Add(streamEstablishTimeout))
			if _, err := writeStreamEstablishHeader(strm, estMsg); err != nil {
				errCh <- err
				strm.Close()
				return
			}

			strm.SetDeadline(time.Time{})
			mtx.Lock()
			if done {
				strm.Close()
				isDone = true
			} else {
				done = true
			}
			mtx.Unlock()
			if !isDone {
				o.c.le.
					WithField("link-id", lnk.GetUUID()).
					WithField("protocol-id", protocolID).
					Debug("opened stream with peer")
				strmCh <- newMountedStream(strm, openOpts, protocolID, lnk)
			}
		},
	)
	c.linksMtx.Unlock()

	if lw != nil {
		defer func() {
			c.linksMtx.Lock()
			c.clearLinkWaiter(lw)
			c.linksMtx.Unlock()
		}()
	}

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	case mstrm := <-strmCh:
		if _, accepted := handler.AddValue(mstrm); !accepted {
			mstrm.GetStream().Close()
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
) (directive.Resolver, error) {
	// opportune moment: if tpt is already available, filter
	tptID := dir.OpenStreamViaLinkTransportConstraint()
	if tptID != 0 {
		select {
		case tpt := <-c.tptCh:
			c.tptCh <- tpt
			if tpt.GetUUID() != tptID {
				return nil, nil
			}
		default:
		}
	}

	// Check transport constraint
	// Return resolver.
	return &openStreamViaLinkResolver{c: c, ctx: ctx, di: di, dir: dir}, nil
}

// _ is a type assertion
var _ directive.Resolver = ((*openStreamViaLinkResolver)(nil))
