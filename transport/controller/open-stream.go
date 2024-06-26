package transport_controller

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
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
	estMsg := NewStreamEstablish(protocolID)

	tpt, err := o.c.GetTransport(ctx)
	if err != nil {
		return err
	}

	if !checkOpenStreamMatchesTpt(o.dir, tpt) {
		return nil
	}

	errCh := make(chan error, 1)
	strmCh := make(chan link.MountedStream, 1)

	c.mtx.Lock()
	lw := c.pushLinkWaiter(
		o.dir.OpenStreamWPTargetPeerID(),
		true,
		func(lnk link.Link, added bool) {
			select {
			case <-ctx.Done():
				return
			default:
			}

			strm, err := lnk.OpenStream(openOpts)
			if err != nil {
				errCh <- err
				/*
					if strm != nil {
						strm.Close()
					}
				*/
				return
			}
			_ = strm.SetWriteDeadline(time.Now().Add(streamEstablishTimeout))
			if _, err := writeStreamEstablishHeader(strm, estMsg); err != nil {
				errCh <- err
				strm.Close()
				return
			}

			_ = strm.SetDeadline(time.Time{})
			o.c.le.
				WithFields(logrus.Fields{
					"link-id":     lnk.GetUUID(),
					"protocol-id": protocolID,
					"src-peer":    lnk.GetLocalPeer().String(),
					"dst-peer":    lnk.GetRemotePeer().String(),
				}).
				Debug("opened stream with peer")
			strmCh <- newMountedStream(strm, openOpts, protocolID, lnk)
		},
	)
	c.mtx.Unlock()

	if lw != nil {
		defer func() {
			c.mtx.Lock()
			c.clearLinkWaiter(lw)
			c.mtx.Unlock()
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
	if tpt := c.tptCtr.GetValue(); tpt != nil {
		if !checkOpenStreamMatchesTpt(dir, tpt) {
			return nil, nil
		}
	}

	// Check transport constraint
	// Return resolver.
	return directive.Resolvers(&openStreamResolver{
		c:   c,
		ctx: ctx,
		di:  di,
		dir: dir,
	}), nil
}

// _ is a type assertion
var _ directive.Resolver = ((*openStreamResolver)(nil))
