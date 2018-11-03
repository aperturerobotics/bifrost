package transport_controller

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/directive"
)

// openStreamResolver resolves OpenStream directives
type openStreamResolver struct {
	c   *Controller
	ctx context.Context
	di  directive.Instance
	dir link.OpenStream
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (o *openStreamResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	c := o.c
	openOpts := o.dir.OpenStreamOpenOpts()
	protocolID := o.dir.OpenStreamProtocolID()
	estMsg := NewStreamEstablish(protocolID)

	errCh := make(chan error, 1)
	strmCh := make(chan link.MountedStream, 1)

	c.linksMtx.Lock()
	lw := c.pushLinkWaiter(o.dir.OpenStreamTargetPeerID(), func(lnk link.Link) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		strm, err := lnk.OpenStream(openOpts)
		if err != nil {
			errCh <- err
			if strm != nil {
				strm.Close()
			}
			return
		}
		strm.SetWriteDeadline(time.Now().Add(streamEstablishTimeout))
		if _, err := writeStreamEstablishHeader(strm, estMsg); err != nil {
			errCh <- err
			strm.Close()
			return
		}

		o.c.le.
			WithField("link-id", lnk.GetUUID()).
			WithField("protocol-id", protocolID).
			Debug("opened stream with peer")
		strmCh <- newMountedStream(strm, openOpts, protocolID, lnk)
	})
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

// resolveOpenStreamWithPeer returns a resolver for opening a stream.
// Negotiates the protocol ID as well.
func (c *Controller) resolveOpenStreamWithPeer(
	ctx context.Context,
	di directive.Instance,
	dir link.OpenStream,
) (directive.Resolver, error) {
	// Check transport constraint
	if tptConstraint := dir.OpenStreamTransportConstraint(); tptConstraint != 0 {
		if c.GetTransport().GetUUID() != tptConstraint {
			return nil, nil
		}
	}

	// Check peer ID constraint
	if srcPeerID := dir.OpenStreamSourcePeerID(); len(srcPeerID) != 0 {
		if srcPeerID != c.localPeerID {
			return nil, nil
		}
	}

	// Return resolver.
	return &openStreamResolver{c: c, ctx: ctx, di: di, dir: dir}, nil
}

// _ is a type assertion
var _ directive.Resolver = ((*openStreamResolver)(nil))
