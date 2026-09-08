//go:build js

package opfsbrowserharness

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	cbc "github.com/aperturerobotics/controllerbus/core"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	"github.com/pkg/errors"
	browser_storage "github.com/s4wave/spacewave/bldr/storage/browser"
	block_store_inmem "github.com/s4wave/spacewave/db/block/store/inmem"
	block_store_overlay "github.com/s4wave/spacewave/db/block/store/overlay"
	"github.com/s4wave/spacewave/db/bucket"
	lookup_concurrent "github.com/s4wave/spacewave/db/bucket/lookup/concurrent"
	bucket_setup "github.com/s4wave/spacewave/db/bucket/setup"
	node_controller "github.com/s4wave/spacewave/db/node/controller"
	volume "github.com/s4wave/spacewave/db/volume"
	world "github.com/s4wave/spacewave/db/world"
	world_block_engine "github.com/s4wave/spacewave/db/world/block/engine"
	"github.com/sirupsen/logrus"
)

// engineBucketID is the bucket the lean world engine uses.
const engineBucketID = "test-bucket"

// World is an opened transport-free Hydra world.
type World struct {
	// Bus is the controller bus.
	Bus bus.Bus
	// StaticResolver resolves controllers on the bus.
	StaticResolver *static.Resolver
	// WS is the writable world state handle.
	WS world.WorldState

	// cancel ends the World lifecycle.
	cancel context.CancelFunc
	// rels releases the controller references.
	rels []func()
}

// Close releases every resource the world holds. The world state must not
// be used after Close returns.
func (w *World) Close() {
	for _, rel := range w.rels {
		rel()
	}
	w.cancel()
}

// OpenWorld constructs a Hydra world backed by the OPFS volume,
// mirroring the browser storage wiring in bldr/storage/browser. Browser
// targets only: OPFS handles come from navigator.storage.
func OpenWorld(ctx context.Context) (*World, error) {
	// Keep the standalone browser proof quiet except for failures.
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	le := logrus.NewEntry(log)

	// Bind all resources to the World lifecycle.
	ctx, cancel := context.WithCancel(ctx)
	w := &World{cancel: cancel}
	fail := func(err error) (*World, error) {
		w.Close()
		return nil, err
	}

	// Register the same storage and World factories used by the browser.
	b, sr, err := cbc.NewCoreBus(ctx, le)
	if err != nil {
		return fail(err)
	}
	w.Bus = b
	w.StaticResolver = sr
	sr.AddFactory(bucket_setup.NewFactory(b))
	sr.AddFactory(node_controller.NewFactory(b))
	sr.AddFactory(lookup_concurrent.NewFactory(b))
	storage := browser_storage.NewOpfsStorage("")
	storage.AddFactories(b, sr)
	sr.AddFactory(block_store_inmem.NewFactory(b))
	sr.AddFactory(block_store_overlay.NewFactory(b))
	sr.AddFactory(boilerplate_controller.NewFactory(b))
	sr.AddFactory(world_block_engine.NewFactory(b))

	// Start the config-set controller and retain its directive.
	_, _, configRef, err := loader.WaitExecControllerRunning(
		ctx,
		b,
		resolver.NewLoadControllerWithConfig(&configset_controller.Config{}),
		nil,
	)
	if err != nil {
		return fail(errors.Wrap(err, "configset controller"))
	}
	w.rels = append(w.rels, configRef.Release)

	// Start the node controller and retain its directive.
	_, _, nodeRef, err := loader.WaitExecControllerRunning(
		ctx,
		b,
		resolver.NewLoadControllerWithConfig(&node_controller.Config{}),
		nil,
	)
	if err != nil {
		return fail(errors.Wrap(err, "node controller"))
	}
	w.rels = append(w.rels, nodeRef.Release)

	// Use the browser storage owner to select the current volume format.
	volumeConfig, err := storage.BuildVolumeConfig(volumeRoot, nil)
	if err != nil {
		return fail(errors.Wrap(err, "build volume config"))
	}
	dv, _, volRef, err := loader.WaitExecControllerRunning(
		ctx,
		b,
		resolver.NewLoadControllerWithConfig(volumeConfig),
		nil,
	)
	if err != nil {
		return fail(errors.Wrap(err, "opfs volume controller"))
	}
	w.rels = append(w.rels, volRef.Release)

	// Open the volume and initialize the World bucket.
	vc := dv.(volume.Controller)
	v, err := vc.GetVolume(ctx)
	if err != nil {
		return fail(errors.Wrap(err, "get volume"))
	}
	if _, _, _, err := v.ApplyBucketConfig(ctx, &bucket.Config{
		Id:  engineBucketID,
		Rev: 1,
	}); err != nil {
		return fail(errors.Wrap(err, "apply bucket config"))
	}

	// Start the World engine over the browser volume.
	transformConf, err := engineTransformConfig(engineBucketID)
	if err != nil {
		return fail(errors.Wrap(err, "build transform config"))
	}
	initRef := &bucket.ObjectRef{
		BucketId:      engineBucketID,
		TransformConf: transformConf,
	}
	engConf := world_block_engine.NewConfig(
		"opfs-harness-engine",
		v.GetID(), engineBucketID,
		"opfs-harness-engine-store",
		initRef,
		nil,
		false,
	)
	_, ctrlRef, err := world_block_engine.StartEngineWithConfig(ctx, b, engConf)
	if err != nil {
		return fail(errors.Wrap(err, "start world engine"))
	}
	w.rels = append(w.rels, ctrlRef.Release)

	// Expose the engine through the ordinary transaction owner.
	busEngine := world.NewBusEngine(ctx, b, "opfs-harness-engine")
	w.WS = world.NewEngineWorldState(busEngine, true)
	return w, nil
}
