package core

import (
	"context"

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

	sr.AddFactory(wtpt.NewFactory(b))
	sr.AddFactory(udptpt.NewFactory(b))
	sr.AddFactory(nctr.NewFactory(b))
	sr.AddFactory(egc.NewFactory(b))
	sr.AddFactory(floodsub_controller.NewFactory(b))

	return b, sr, nil
}
