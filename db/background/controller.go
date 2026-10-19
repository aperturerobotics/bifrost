// Package background runs low-priority download jobs that yield to the work a
// person is waiting for.
//
// A component registers a job with the RunJob directive. The controller runs
// equal keys once, admits a fixed number of jobs at a time, and retries a
// failed job with backoff. Each attempt runs with a context marked by
// block.WithBackground, so the stores the job reads from admit its fetches only
// while no foreground fetch needs the link.
package background

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

// Version is the version of the controller implementation.
var Version = controller.MustParseVersion("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "spacewave/background"

// Controller resolves RunJob directives.
type Controller struct {
	// le is the logger.
	le *logrus.Entry
	// conf is the controller configuration.
	conf *Config
	// slots admits running jobs in arrival order.
	slots *semaphore.Weighted
}

// NewController constructs a background job controller.
func NewController(le *logrus.Entry, conf *Config) *Controller {
	return &Controller{
		le:    le,
		conf:  conf,
		slots: semaphore.NewWeighted(int64(conf.GetRunSlotsOrDefault())),
	}
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() *controller.Info {
	return controller.NewInfo(ControllerID, Version, "background job controller")
}

// Execute executes the controller. Jobs run in their directive resolvers.
func (c *Controller) Execute(ctx context.Context) error {
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
func (c *Controller) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	if d, ok := di.GetDirective().(RunJob); ok {
		return directive.R(newJobResolver(c, d), nil)
	}
	return nil, nil
}

// Close releases any resources used by the controller.
func (c *Controller) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = (*Controller)(nil)
