package transport

// ProvideTransport directive provides a transport to all controllers.
type ProvideTransport interface {
	// GetProvidedTransport returns the provided transport.
	GetProvidedTransport() Transport
}
