package transport

import (
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/controller"
)

// Handler handles newly built links.
type Handler interface {
	// AddLink handles a new link.
	AddLink(link.Link)
}

// Transport is similar to a NIC, yielding links to remote peers.
type Transport interface {
	controller.Controller

	// GetUUID returns a host-unique ID for this transport.
	GetUUID() uint64
	// GetLinks returns the list of links this transport has active.
	GetLinks() []link.Link
}
