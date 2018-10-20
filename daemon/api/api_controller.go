package api

import (
	"context"
	"net"

	"github.com/aperturerobotics/bifrost/node"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Controller implements the API controller. The controller looks up the Node,
// acquires its identity, constructs the GRPC listener, and responds to incoming
// API calls.
type Controller struct {
	// le is the logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// peerIDConstraint constrains the node peer id
	peerIDConstraint peer.ID
	// listenAddr is the listen address
	listenAddr string
}

// NewController constructs a new API controller.
func NewController(
	le *logrus.Entry,
	listenAddr string,
	bus bus.Bus,
	nodePeerIDConstraint peer.ID,
) *Controller {
	return &Controller{
		le:         le,
		bus:        bus,
		listenAddr: listenAddr,

		peerIDConstraint: nodePeerIDConstraint,
	}
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"grpc api controller",
	)
}

// Execute executes the API controller and the listener.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	// Acquire a handle to the node.
	c.le.
		WithField("peer-id", c.peerIDConstraint.Pretty()).
		Info("looking up node with peer ID")
	n, err := node.GetNodeWithPeerID(ctx, c.bus, c.peerIDConstraint)
	if err != nil {
		return err
	}

	// Construct the API
	api, err := NewAPI(c.bus, n)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	api.RegisterAsGRPCServer(server)

	lis, err := net.Listen("tcp", c.listenAddr)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		server.Stop()
		return nil
	case err := <-errCh:
		return err
	}
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(di directive.Instance) (directive.Resolver, error) {
	return nil, nil
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	// nil references to help GC along
	c.le = nil
	c.bus = nil
	c.peerIDConstraint = ""

	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
