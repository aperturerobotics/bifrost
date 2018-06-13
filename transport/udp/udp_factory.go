package udp

import (
	"github.com/aperturerobotics/bifrost/transport"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	mafmt "github.com/whyrusleeping/mafmt"
)

// MaddrFmt is the multiaddr format for UDP.
var MaddrFmt = mafmt.UDP

// Factory configures UDP transports.
type Factory struct {
}

// Register registers the UDP factory with a controller.
func Register(c transport.Controller) error {
	return c.RegisterTransportFactory(&Factory{})
}

// BuildListeners builds transport listeners given the multiaddress.
// If the multiaddress format is not supported, returns nil.
func (f *Factory) BuildListeners(listenAddr ma.Multiaddr) ([]transport.Transport, error) {
	if !MaddrFmt.Matches(listenAddr) {
		return nil, nil
	}

	na, err := manet.ToNetAddr(listenAddr)
	if err != nil {
		return nil, err
	}

	u, err := NewUDP(na.String())
	if err != nil {
		return nil, err
	}

	return []transport.Transport{u}, nil
}

// Execute executes the UDP factory.
// For example, NICs might be detected and automatically yield transports.
// If execute returns an error, will be retried. If err is nil, will not retry.
func (f *Factory) Execute(ctx context.Context) error {
	// TODO: implement UDP discovery.
	return nil
}

// _ is a type assertion.
var _ transport.Factory = ((*Factory)(nil))
