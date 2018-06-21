package transport

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
)

// Transport is similar to a NIC, yielding links to remote peers.
type Transport interface {
	// Execute processes the transport, emitting events to the handler.
	// Fatal errors are returned.
	Execute(ctx context.Context, handler Handler) error
}

// Handler handles lifecycle events of a transport.
type Handler interface {
	// HandleLink handles an incoming link.
	HandleLink(link.Link)
}

// FactoryHandler handles events from a factory.
type FactoryHandler interface {
	Handler
	// HandleTransport handles a new transport.
	HandleTransport(Transport)
}

// Factory configures transports given listen addresses.
type Factory interface {
	// Execute processes the factory and produces transports.
	// For example, NICs might be detected and automatically yield transports.
	// If execute returns an error, will be retried. If err is nil, will not retry.
	Execute(ctx context.Context, handler FactoryHandler) error
}

// Controller manages transport factories and transports.
type Controller interface {
	// RegisterTransportFactory registers a transport factory.
	RegisterTransportFactory(Factory) error
}
