package prom_plugin

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/starpc/srpc"
	"github.com/pkg/errors"
	bldr_plugin "github.com/s4wave/spacewave/bldr/plugin"
	resource "github.com/s4wave/spacewave/bldr/resource"
	resource_client "github.com/s4wave/spacewave/bldr/resource/client"
	resource_server "github.com/s4wave/spacewave/bldr/resource/server"
	bifrost_rpc "github.com/s4wave/spacewave/net/rpc"
	s4wave_objecttype_registry "github.com/s4wave/spacewave/sdk/objecttype/registry"
	s4wave_plugin_registration "github.com/s4wave/spacewave/sdk/plugin/registration"
	s4wave_prom "github.com/s4wave/spacewave/sdk/prom"
)

// ControllerID is the controller identifier.
const ControllerID = "plugin/prom"

// PluginID is the manifest ID of the Prometheus plugin.
const PluginID = "spacewave-prom"

// Version is the component version.
var Version = controller.MustParseVersion("0.0.1")

// Controller registers prom/target with the core plugin and serves its
// resources.
type Controller struct {
	*bus.BusController[*Config]
	// Holder publishes the registrations and reports when they are ready.
	*s4wave_plugin_registration.Holder

	// mux serves the plugin's resource service.
	mux srpc.Invoker
}

// NewFactory constructs the component factory.
func NewFactory(b bus.Bus) controller.Factory {
	return bus.NewBusControllerFactory(
		b,
		ConfigID,
		ControllerID,
		Version,
		"Prometheus metrics plugin",
		func() *Config {
			return &Config{}
		},
		func(base *bus.BusController[*Config]) (*Controller, error) {
			// Serve the target ObjectType through the resource server.
			h := newHandler(base.GetLogger(), base.GetConfig().CommitWindow())
			rootMux := resource_server.NewResourceMux(func(mux srpc.Mux) error {
				return s4wave_objecttype_registry.SRPCRegisterObjectTypeHandlerService(mux, h)
			})
			serverMux := srpc.NewMux()
			if err := resource_server.NewResourceServer(rootMux).Register(serverMux); err != nil {
				return nil, err
			}
			return &Controller{
				BusController: base,
				Holder:        s4wave_plugin_registration.NewHolder(register),
				mux:           serverMux,
			}, nil
		},
	)
}

// Execute registers prom/target with the core plugin until ctx ends.
func (c *Controller) Execute(ctx context.Context) error {
	return c.Hold(ctx, c.GetBus())
}

// HandleDirective serves the plugin's resource service.
func (c *Controller) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	if d, ok := di.GetDirective().(bifrost_rpc.LookupRpcService); ok && d.LookupRpcServiceID() == resource.SRPCResourceServiceServiceID {
		return directive.R(bifrost_rpc.NewLookupRpcServiceResolver(c.mux), nil)
	}
	return nil, nil
}

// register registers the prom/target ObjectType through scope. It returns the
// registrations made, even on error.
func register(
	ctx context.Context,
	client *resource_client.Client,
	scope srpc.Client,
) ([]resource_client.ResourceRef, error) {
	ot, err := s4wave_objecttype_registry.NewSRPCObjectTypeRegistryResourceServiceClient(scope).
		RegisterObjectType(ctx, &s4wave_objecttype_registry.RegisterObjectTypeRequest{
			TypeId:   s4wave_prom.TargetTypeID,
			PluginId: PluginID,
			Metadata: &s4wave_objecttype_registry.ObjectTypeMetadata{
				DisplayName: "Scrape target",
				IconName:    "LuActivity",
			},
		})
	if err != nil {
		return nil, errors.Wrap(err, "register object type")
	}
	if ot.GetResourceId() == 0 {
		return nil, errors.New("register object type: no registration returned")
	}
	return []resource_client.ResourceRef{client.CreateResourceReference(ot.GetResourceId())}, nil
}

// _ is a type assertion.
var (
	_ controller.Controller                  = (*Controller)(nil)
	_ bldr_plugin.InitialCapabilityRegistrar = (*Controller)(nil)
)
