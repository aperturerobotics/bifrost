package node_controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/bifrost/node"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "bifrost/node/1"

// Controller is the Node controller.
// It implements node.Node as a controller.
type Controller struct {
	// Peer is the underlying peer
	peer.Peer
	// le is the root logger
	le *logrus.Entry

	// transportsMtx guards the transports map
	transportsMtx sync.Mutex
	// transports are the running transports
	transports map[uint64]transport.Transport
}

// NewController constructs a new node controller.
// If privKey is nil, one will be generated.
func NewController(le *logrus.Entry, privKey crypto.PrivKey) (*Controller, error) {
	var err error

	p, err := peer.NewPeer(privKey)
	if err != nil {
		return nil, err
	}

	return &Controller{
		Peer: p,

		le:         le,
		transports: make(map[uint64]transport.Transport),
	}, nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	// TODO implement core node management loop
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(di directive.Instance) (directive.Resolver, error) {
	dir := di.GetDirective()
	if d, ok := dir.(node.GetNode); ok {
		return c.resolveGetNode(d), nil
	}

	return nil, nil
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"node controller "+c.GetPeerID().Pretty(),
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
var _ node.Node = ((*Controller)(nil))
