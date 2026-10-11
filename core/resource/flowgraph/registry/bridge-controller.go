package resource_flowgraph_registry

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	s4wave_flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
)

// BridgeController publishes compile proxies for registered plugin node types.
type BridgeController struct {
	// b resolves the compiled controller config constructors.
	b bus.Bus
	// registry retains compiler capabilities and publishes registration changes.
	registry *RegistryResource
}

// NewBridgeController constructs the bridge for a daemon's node type registry.
func NewBridgeController(b bus.Bus, registry *RegistryResource) *BridgeController {
	return &BridgeController{b: b, registry: registry}
}

// GetControllerInfo describes the node type bridge.
func (c *BridgeController) GetControllerInfo() *controller.Info {
	return controller.NewInfo("resource/flowgraph-node-types", controller.MustParseVersion("0.0.1"), "plugin node type compile proxies")
}

// Execute leaves registration watches to the lookup resolvers.
func (c *BridgeController) Execute(context.Context) error {
	return nil
}

// HandleDirective follows registration arrival, withdrawal and replacement.
func (c *BridgeController) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	// Match standing demand even when its plugin has not registered yet.
	dir, ok := di.GetDirective().(s4wave_flowgraph.LookupFlowgraphNodeType)
	if !ok {
		return nil, nil
	}

	// Publish only this resolver's value as the compiler capability changes.
	return directive.R(directive.NewFuncResolver(func(ctx context.Context, handler directive.ResolverHandler) error {
		var current *nodeTypeRegistration
		var valueID uint32
		for {
			// Read the compiler and its change notification together.
			var next *nodeTypeRegistration
			var wait <-chan struct{}
			c.registry.bcast.HoldLock(func(_ func(), getWait func() <-chan struct{}) {
				next = c.registry.registrations[dir.LookupFlowgraphNodeTypeID()]
				wait = getWait()
			})

			// Withdraw old proxies before publishing a replacement.
			if next != current {
				if valueID != 0 {
					handler.RemoveValue(valueID)
					valueID = 0
				}
				current = next
				if next != nil {
					valueID, _ = handler.AddValue(NewNodeTypeProxy(c.b, next.metadata, next.client, next.ctx))
				}
			}

			// Keep unresolved demand alive until the registration changes.
			handler.MarkIdle(true)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-wait:
			}
			handler.MarkIdle(false)
		}
	}), nil)
}

// Close releases no state because registration Resources retain the compilers.
func (c *BridgeController) Close() error {
	return nil
}

// _ is a type assertion.
var _ controller.Controller = (*BridgeController)(nil)
