package transport

import (
	"github.com/aperturerobotics/bifrost/link"
)

// Handler handles lifecycle events of a transport.
type Handler interface {
	// AddLink handles an incoming link.
	AddLink(link.Link)
}

// Transport is similar to a NIC, yielding links to remote peers.
type Transport interface {
	// GetUUID returns a host-unique ID for this transport.
	GetUUID() uint64
	// GetLinks returns the list of links this transport has active.
	GetLinks() []link.Link
	// RestoreLink instructs the transport to attempt to restore a link.
	// If the link would be a duplicate, return the existing link.
	// If the link is no longer valid, return nil.
	// In an exceptional case (invalid data), return an error.
	RestoreLink(*link.LinkInfo) (link.Link, error)
}

// FactoryHandler handles events from a factory.
type FactoryHandler interface {
	// AddTransport handles a new transport.
	AddTransport(Transport)
}

// Factory configures and tracks transports given configuration.
type Factory interface {
}

// Controller manages transport factories and transports.
type Controller interface {
	// RegisterTransportFactory registers a transport factory.
	RegisterTransportFactory(Factory) error
}
