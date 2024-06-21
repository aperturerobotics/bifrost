package transport_controller

import (
	"context"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
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
		tpt, err := o.c.GetTransport(ctx)
		if err != nil {
			return err
		}
		if tpt.GetUUID() != tptID {
			return nil
		}
	}

	errCh := make(chan error, 1)
	strmCh := make(chan link.MountedStream, 1)
	var mtx sync.Mutex
	var done bool

	c.mtx.Lock()
	linkWaiterCallback := func(establishedLink link.Link, successfullyAdded bool) {
		correctLink := establishedLink.GetUUID() == lnkUUID
		if !correctLink || !successfullyAdded || ctx.Err() != nil {
			return
		}

		// Check if the operation was already completed.
		mtx.Lock()
		operationAlreadyCompleted := done
		mtx.Unlock()

		// If so, no further action is needed.
		if operationAlreadyCompleted {
			return
		}

		// Attempt to open a stream with the link.
		stream, streamErr := establishedLink.OpenStream(openOpts)
		if streamErr != nil {
			errCh <- streamErr
			return
		}

		_ = stream.SetWriteDeadline(time.Now().Add(streamEstablishTimeout))

		_, headerWriteErr := writeStreamEstablishHeader(stream, estMsg)
		if headerWriteErr != nil {
			errCh <- headerWriteErr
			stream.Close()
			return
		}

		_ = stream.SetDeadline(time.Time{})

		mtx.Lock()
		if done {
			stream.Close()
			operationAlreadyCompleted = true
		} else {
			done = true
		}
		mtx.Unlock()

		if !operationAlreadyCompleted {
			o.c.le.
				WithField("link-id", establishedLink.GetUUID()).
				WithField("protocol-id", protocolID).
				Debug("opened stream with peer")

			strmCh <- newMountedStream(stream, openOpts, protocolID, establishedLink)
		}
	}

	// Register the callback and wait for the link to establish.
	lw := c.pushLinkWaiter(peer.ID(""), false, linkWaiterCallback)

	// Unlock mutex
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

// resolveOpenStreamViaLink returns a resolver for opening a stream.
// Negotiates the protocol ID as well.
func (c *Controller) resolveOpenStreamViaLink(
	ctx context.Context,
	di directive.Instance,
	dir link.OpenStreamViaLink,
) ([]directive.Resolver, error) {
	// opportune moment: if tpt is already available, filter
	tptID := dir.OpenStreamViaLinkTransportConstraint()
	if tptID != 0 {
		if tpt := c.tptCtr.GetValue(); tpt != nil {
			if tpt.GetUUID() != tptID {
				return nil, nil
			}
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
