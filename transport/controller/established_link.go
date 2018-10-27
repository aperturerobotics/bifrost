package controller

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// linkHoldOpenDur is the minimum amount of time to hold a link open.
// TODO: move this to a more configurable location
var linkHoldOpenDur = time.Duration(10) * time.Second

// establishedLink holds state for an established link.
type establishedLink struct {
	// le is the log entry
	le *logrus.Entry
	// Link is the link.
	Link link.Link
	// DirectiveInstance is the EstablishLink directive instance.
	DirectiveInstance directive.Instance
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
) (*establishedLink, error) {
	// Construct EstablishLink directive.
	// The controller will match the directive to this link.
	// Close the reference after a hold-open time.
	// Close the link when the directive expires.
	di, dir, err := b.AddDirective(
		link.NewEstablishLinkSingleton(lnk.GetRemotePeer()),
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
	}
	di.AddDisposeCallback(func() { _ = lnk.Close() })
	go el.initialHoldOpen(ctx, dir)

	return el, nil
}

// initialHoldOpen manages the initial hold-open period for the link.
func (e *establishedLink) initialHoldOpen(ctx context.Context, ref directive.Reference) {
	ctxCancel := e.Cancel
	defer ctxCancel()

	ht := time.NewTimer(linkHoldOpenDur)
	defer ht.Stop()

	e.DirectiveInstance.AddDisposeCallback(func() {
		e.le.Debug("establish link directive expired, closing link")
		e.Link.Close()
		e.Cancel()
	})
	select {
	case <-ctx.Done():
	case <-ht.C:
		e.le.
			WithField("link-uuid", e.Link.GetUUID()).
			WithField("duration", linkHoldOpenDur.String()).
			Debug("link hold-open duration expired")
	}

	ref.Release()
}
