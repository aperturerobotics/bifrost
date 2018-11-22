package stream_forwarding

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// Controller implements the API controller. The controller looks up the Node,
// acquires its identity, constructs the GRPC listener, and responds to incoming
// API calls.
type Controller struct {
	// le is the logger
	le *logrus.Entry
	// conf is the config
	conf *Config
}

// NewController constructs a new API controller.
func NewController(
	le *logrus.Entry,
	conf *Config,
) *Controller {
	return &Controller{
		le:   le,
		conf: conf,
	}
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"forwarding controller",
	)
}

// Execute executes the forwarding controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	// TODO listen, etc, etc, etc.
	<-ctx.Done()
	return nil
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
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
