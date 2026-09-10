package core

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	cbc "github.com/aperturerobotics/controllerbus/core"
	db_storage "github.com/s4wave/spacewave/db/core/storage"
	"github.com/s4wave/spacewave/db/dex/psecho"
	bifrostcore "github.com/s4wave/spacewave/net/core"
	nctr "github.com/s4wave/spacewave/net/peer/controller"
	"github.com/sirupsen/logrus"
)

// NewCoreBus constructs a standard in-memory bus stack with basic Hydra controllers.
func NewCoreBus(
	ctx context.Context,
	le *logrus.Entry,
	opts ...cbc.Option,
) (bus.Bus, *static.Resolver, error) {
	// Construct the controller bus and static resolver.
	b, sr, err := cbc.NewCoreBus(ctx, le, opts...)
	if err != nil {
		return nil, nil, err
	}

	// Register the standard controller factories.
	AddFactories(b, sr)
	return b, sr, nil
}

// AddFactories adds factories to an existing static resolver.
func AddFactories(b bus.Bus, sr *static.Resolver) {
	// Register platform and network factories.
	addNativeFactories(b, sr)
	bifrostcore.AddFactories(b, sr)

	// Register peer and storage factories.
	sr.AddFactory(nctr.NewFactory(b))
	db_storage.AddFactories(b, sr)

	// Register the pub-sub exchange factory.
	sr.AddFactory(psecho.NewFactory(b))
}
