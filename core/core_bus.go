package core

import (
	"context"

	bifrosteg "github.com/aperturerobotics/bifrost/entitygraph"
	"github.com/aperturerobotics/bifrost/link/hold-open"
	nctr "github.com/aperturerobotics/bifrost/peer/controller"
	"github.com/aperturerobotics/bifrost/pubsub/floodsub/controller"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	cbc "github.com/aperturerobotics/controllerbus/core"
	egc "github.com/aperturerobotics/entitygraph/controller"
	"github.com/sirupsen/logrus"
)

// NewCoreBus constructs a standard in-memory bus stack with Bifrost controllers.
func NewCoreBus(
	ctx context.Context,
	le *logrus.Entry,
	builtInFactories ...controller.Factory,
) (bus.Bus, *static.Resolver, error) {
	b, sr, err := cbc.NewCoreBus(ctx, le, builtInFactories...)
	if err != nil {
		return nil, nil, err
	}

	AddFactories(b, sr)
	return b, sr, nil
}

// AddFactories adds factories to an existing static resolver.
func AddFactories(b bus.Bus, sr *static.Resolver) {
	sr.AddFactory(wtpt.NewFactory(b))
	sr.AddFactory(udptpt.NewFactory(b))
	sr.AddFactory(nctr.NewFactory(b))
	sr.AddFactory(egc.NewFactory(b))
	sr.AddFactory(bifrosteg.NewFactory(b))
	sr.AddFactory(floodsub_controller.NewFactory(b))
	sr.AddFactory(link_holdopen_controller.NewFactory(b))
}
