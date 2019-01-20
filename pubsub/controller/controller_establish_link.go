package pubsub_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/directive"
)

// handleEstablishLink handles an EstablishLink directive.
func (c *Controller) handleEstablishLink(
	ctx context.Context,
	di directive.Instance,
	d link.EstablishLinkWithPeer,
) {
	handler := newEstablishLinkHandler(c)
	ref := di.AddReference(handler, true)
	if ref == nil {
		return
	}
	handler.ref = ref
	c.cleanupRefs = append(c.cleanupRefs, ref)
}
