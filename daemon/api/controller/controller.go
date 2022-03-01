package bifrost_api_controller

import (
	"context"
	"net"

	bifrost_api "github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/controllerbus/bus"
	cbapi "github.com/aperturerobotics/controllerbus/bus/api"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

// Version is the API version.
var Version = semver.MustParse("0.0.1")

// Controller implements the API controller. The controller looks up the Node,
// acquires its identity, listens and responds to incoming API calls.
type Controller struct {
	// le is the logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// listenAddr is the listen address
	listenAddr string
	// conf is the config
	conf *Config
}

// NewController constructs a new API controller.
func NewController(
	le *logrus.Entry,
	bus bus.Bus,
	listenAddr string,
	conf *Config,
) *Controller {
	return &Controller{
		le:         le,
		bus:        bus,
		listenAddr: listenAddr,
		conf:       conf,
	}
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"api controller",
	)
}

// Execute executes the API controller and the listener.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	// Construct the API
	api, err := bifrost_api.NewAPI(c.bus, c.conf.GetApiConfig())
	if err != nil {
		return err
	}

	mux := drpcmux.New()
	api.RegisterAsDRPCServer(mux)

	// controllerbus api
	if !c.conf.GetDisableBusApi() {
		bapi := cbapi.NewAPI(c.bus, c.conf.GetBusApiConfig())
		bapi.RegisterAsDRPCServer(mux)
	}

	c.le.Debug("starting listener")
	lis, err := net.Listen("tcp", c.listenAddr)
	if err != nil {
		return err
	}
	c.le.Debugf("api listening: %s", lis.Addr().String())

	srv := drpcserver.New(mux)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ctx, lis)
		_ = lis.Close()
	}()

	select {
	case <-ctx.Done():
		_ = lis.Close()
		return nil
	case err := <-errCh:
		return err
	}
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(ctx context.Context, di directive.Instance) (directive.Resolver, error) {
	return nil, nil
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	// nil references to help GC along
	c.le = nil
	c.bus = nil

	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
