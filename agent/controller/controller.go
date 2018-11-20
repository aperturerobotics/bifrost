package agent_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/agent"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "bifrost/agent/1"

// Controller is the Agent controller.
// It implements agent.Agent as a controller.
type Controller struct {
	// Peer is the underlying peer
	peer.Peer
	// le is the root logger
	le *logrus.Entry
}

// NewController constructs a new agent controller.
// If privKey is nil, one will be generated.
func NewController(le *logrus.Entry, privKey crypto.PrivKey) (*Controller, error) {
	var err error

	p, err := peer.NewPeer(privKey)
	if err != nil {
		return nil, err
	}

	return &Controller{
		le:   le,
		Peer: p,
	}, nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	// Acquire handle to node GetNode.
	// Register agent private key with node.
	// Issue directive looking up the agent listener / handler
	// Process incoming agent API calls.

	return nil
}

// HandleDirective asks if the handler can resolve the directive.
func (c *Controller) HandleDirective(
	ctx context.Context,
	di directive.Instance,
) (directive.Resolver, error) {
	return nil, nil
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"agent controller "+c.GetPeerID().Pretty(),
	)
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))

// _ is a type assertion
var _ agent.Agent = ((*Controller)(nil))
