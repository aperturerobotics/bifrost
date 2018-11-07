package entitygraph_controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/entitygraph"
	"github.com/aperturerobotics/entitygraph/entity"
	"github.com/aperturerobotics/entitygraph/store"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "bifrost/entitygraph/1"

// Controller exposes Bifrost resources to the Entity Graph.
// It handles CollectEntityGraph directives.
// It handles EstablishLink directives to observe running and pending links.
type Controller struct {
	// le is the root logger
	le *logrus.Entry
	// store is the entity store
	store *store.Store

	// mtx guards the collectors map
	mtx sync.Mutex
	// cleanupRefs are the refs to cleanup
	cleanupRefs []directive.Reference
	// handlers should be called with values from store
	handlers []store.Handler
	// values are the known values
	values map[store.EntityMapKey]entity.Entity
}

// NOTE: CollectEntityGraph de-duplicates entity objects!
// That means:
// - emit two node Entity objects per EstablishLink interest
// - emit one Entity<Link> object per EstablishLink interest
// - emit one Entity<Link> object per known Link yielded by EstablishLink

// NewController constructs a new entitygraph controller.
func NewController(le *logrus.Entry) *Controller {
	c := &Controller{
		le:     le,
		values: make(map[store.EntityMapKey]entity.Entity),
	}
	c.store = store.NewStore(newStoreHandler(c))
	return c
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	<-ctx.Done()
	c.mtx.Lock()
	for _, ref := range c.cleanupRefs {
		ref.Release()
	}
	c.cleanupRefs = nil
	c.mtx.Unlock()
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(
	ctx context.Context,
	di directive.Instance,
) (directive.Resolver, error) {
	dir := di.GetDirective()
	switch d := dir.(type) {
	case link.EstablishLink:
		c.handleEstablishLink(ctx, di, d)
	case entitygraph.CollectEntityGraph:
		return c.handleCollectEntityGraph(ctx, di, d)
	}

	return nil, nil
}

// handleCollectEntityGraph handles a CollectEntityGraph directive.
func (c *Controller) handleCollectEntityGraph(
	ctx context.Context,
	di directive.Instance,
	d entitygraph.CollectEntityGraph,
) (directive.Resolver, error) {
	return newCollectResolver(ctx, c), nil
}

// handleEstablishLink handles an EstablishLink directive.
func (c *Controller) handleEstablishLink(
	ctx context.Context,
	di directive.Instance,
	d link.EstablishLink,
) {
	handler := newEstablishLinkHandler(c)
	ref := di.AddReference(handler, true)
	if ref == nil {
		return
	}
	handler.ref = ref

	c.mtx.Lock()
	c.cleanupRefs = append(c.cleanupRefs, ref)
	c.mtx.Unlock()
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"bifrost entitygraph reporter controller ",
	)
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
