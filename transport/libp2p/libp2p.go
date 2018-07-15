package libp2p

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
	lt "github.com/libp2p/go-libp2p-transport"
	ma "github.com/multiformats/go-multiaddr"
)

// TransportFactory wraps a libp2p transport into a bifrost factory.
type TransportFactory struct {
	tpt          lt.Transport
	controllerID string
	ver          semver.Version
}

// NewTransportFactory wraps a libp2p transport into a Bifrost factory.
func NewTransportFactory(
	controllerID string,
	tptVer semver.Version,
	tpt lt.Transport,
) *TransportFactory {
	return &TransportFactory{
		tpt:          tpt,
		controllerID: controllerID,
		ver:          tptVer,
	}
}

// GetControllerID returns the unique ID for the controller.
func (t *TransportFactory) GetControllerID() string {
	return t.controllerID
}

// ConstructConfig constructs an instance of the controller configuration.
func (t *TransportFactory) ConstructConfig() config.Config {
	return NewTransportConfig(t.controllerID, TransportConfigInner{})
}

// Construct constructs the associated controller given configuration.
func (t *TransportFactory) Construct(
	conf config.Config,
	opts controller.ConstructOpts,
) (controller.Controller, error) {
	type tconf interface {
		ParseListenMultiaddr() (ma.Multiaddr, error)
	}
	tc, ok := conf.(tconf)
	if !ok {
		return nil, errors.New("config type not recognized")
	}

	listenMultiaddr, err := tc.ParseListenMultiaddr()
	if err != nil {
		return nil, err
	}

	le := opts.GetLogger()
	return NewListener(le, t.tpt, listenMultiaddr), nil
}

// GetVersion returns the version of this controller.
func (t *TransportFactory) GetVersion() semver.Version {
	return t.ver
}

// _ is a type assertion
var _ controller.Factory = ((*TransportFactory)(nil))
