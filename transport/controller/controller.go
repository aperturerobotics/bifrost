package controller

import (
	"context"
	"strings"
	"sync"

	"github.com/aperturerobotics/bifrost/node"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// Constructor constructs a transport with common parameters.
type Constructor func(le *logrus.Entry, pkey crypto.PrivKey) (transport.Transport, error)

// Controller implements a common transport controller.
// The controller looks up the Node, acquires its identity, constructs the
// transport, and manages the lifecycle of dialing and accepting links.
type Controller struct {
	// le is the logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// ctor is the constructor
	ctor Constructor
	// peerIDConstraint constrains the node peer id
	peerIDConstraint peer.ID

	// tptMtx is the transport mutex
	tptMtx sync.Mutex
	// tpt is the transport
	tpt transport.Transport

	// transportID is the transport identifier.
	transportID string
	// transportVersion is the transport version
	transportVersion semver.Version
}

// NewController constructs a new transport controller.
func NewController(
	le *logrus.Entry,
	bus bus.Bus,
	nodePeerIDConstraint peer.ID,
	ctor Constructor,
	transportID string,
	transportVersion semver.Version,
) *Controller {
	return &Controller{
		le:   le,
		bus:  bus,
		ctor: ctor,

		peerIDConstraint: nodePeerIDConstraint,
		transportID:      transportID,
		transportVersion: transportVersion,
	}
}

// GetControllerID returns the controller ID.
func (c *Controller) GetControllerID() string {
	return strings.Join([]string{"bifrost", "transport", c.transportID, c.transportVersion.String()}, "/")
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		c.GetControllerID(),
		c.transportVersion,
		"transport controller "+c.transportID+"@"+c.transportVersion.String(),
	)
}

// Execute executes the transport controller and the transport.
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

	// Get the priv key
	privKey := n.GetPrivKey()

	// Construct the transport
	tpt, err := c.ctor(
		c.le,
		privKey,
	)
	if err != nil {
		return err
	}

	c.tptMtx.Lock()
	c.tpt = tpt
	c.tptMtx.Unlock()

	tptErr := make(chan error, 1)
	go func() {
		c.le.Debug("executing transport")
		if err := tpt.Execute(ctx); err != nil {
			tptErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-tptErr:
		return err
	}
}

// GetTransport returns the controlled transport.
// This may be nil until the transport is constructed.
func (c *Controller) GetTransport() transport.Transport {
	c.tptMtx.Lock()
	defer c.tptMtx.Unlock()

	return c.tpt
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
	c.ctor = nil
	c.le = nil
	c.bus = nil
	c.peerIDConstraint = ""

	c.tptMtx.Lock()
	tpt := c.tpt
	c.tpt = nil
	c.tptMtx.Unlock()

	if tpt != nil {
		return tpt.Close()
	}

	return nil
}

// _ is a type assertion
var _ transport.Controller = ((*Controller)(nil))
