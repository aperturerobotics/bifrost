package transport

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	ma "github.com/multiformats/go-multiaddr"
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
	// Returns immediately.
	HandleLink(link.Link)
}

// Factory configures transports given listen addresses.
type Factory interface {
	// BuildListeners builds transport listeners given the multiaddress.
	// If the multiaddress format is not supported, returns nil.
	BuildListeners(listenAddr ma.Multiaddr) ([]Transport, error)
	// Execute processes any auto-configured transports.
	// For example, NICs might be detected and automatically yield transports.
	// If execute returns an error, will be retried. If err is nil, will not retry.
	Execute(ctx context.Context) error
}

// FactoryHandler handles events yielded by the factory.
type FactoryHandler interface {
	// HandleTransport handles an automatically built transport.
	// Returns immediately.
	HandleTransport(Transport)
}

// Controller manages transport factories and transports.
type Controller interface {
	// RegisterTransportFactory registers a transport factory.
	RegisterTransportFactory(Factory) error
}
