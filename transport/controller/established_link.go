package transport_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// establishedLink holds state for an established link.
type establishedLink struct {
	// le is the log entry
	le *logrus.Entry
	// Link is the link.
	Link link.Link
	// DirectiveInstance is the EstablishLink directive instance.
	DirectiveInstance directive.Instance
	// Controller is the transport controller
	Controller *Controller
	// Cancel closes any pending goroutines related to the link
	Cancel context.CancelFunc
}

// newEstablishedLink constructs the new establishedLink object.
// The EstablishLink directive is fulfilled on the controller bus.
func newEstablishedLink(
	le *logrus.Entry,
	rctx context.Context,
	b bus.Bus,
	lnk link.Link,
	ctrl *Controller,
) (*establishedLink, error) {
	// Construct EstablishLink directive.
	// The controller will match the directive to this link.
	// Close the reference after a hold-open time.
	// Close the link when the directive expires.
	di, dir, err := b.AddDirective(
		link.NewEstablishLinkWithPeer(lnk.GetRemotePeer()),
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Start the close goroutine
	ctx, ctxCancel := context.WithCancel(rctx)
	el := &establishedLink{
		le:                le.WithField("peer-id", lnk.GetRemotePeer().Pretty()),
		Link:              lnk,
		DirectiveInstance: di,
		Cancel:            ctxCancel,
		Controller:        ctrl,
	}
	di.AddDisposeCallback(func() {
		_ = lnk.Close()
		ctxCancel()
	})
	go el.acceptStreamPump(ctx)

	dir.Release()
	return el, nil
}

func (e *establishedLink) acceptStreamPump(ctx context.Context) {
	// accept streams
	lnk := e.Link
	ctrl := e.Controller
	defer e.Cancel()

	for {
		// e.le.Debug("waiting to accept stream")
		strm, strmOpts, err := lnk.AcceptStream()
		if err != nil {
			if err != context.Canceled {
				e.le.WithError(err).Warn("link accept stream errored")
			}
			return
		}

		if strm != nil {
			// e.le.Debug("accepted incoming stream")
			go ctrl.HandleIncomingStream(ctx, lnk, strm, strmOpts)
		}
	}
}
