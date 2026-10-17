package world_block_engine_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/testbed"
	"github.com/s4wave/spacewave/db/world"
	world_block_engine "github.com/s4wave/spacewave/db/world/block/engine"
	"github.com/sirupsen/logrus"
)

// failFirstRunFactory constructs a controller whose first run fails before
// serving an engine, then the real World engine controller.
type failFirstRunFactory struct {
	*world_block_engine.Factory
	runs atomic.Int32
}

// Construct fails the first run and constructs the real controller after.
func (f *failFirstRunFactory) Construct(ctx context.Context, conf config.Config, opts controller.ConstructOpts) (controller.Controller, error) {
	if f.runs.Add(1) == 1 {
		return failedRun{}, nil
	}
	return f.Factory.Construct(ctx, conf, opts)
}

// failedRun is a controller run that exits with an error before serving an
// engine.
type failedRun struct{}

// GetControllerInfo identifies the failed run.
func (failedRun) GetControllerInfo() *controller.Info {
	return controller.NewInfo(world_block_engine.ControllerID, controller.MustParseVersion("0.0.1"), "failed run")
}

// HandleDirective resolves nothing.
func (failedRun) HandleDirective(context.Context, directive.Instance) ([]directive.Resolver, error) {
	return nil, nil
}

// Execute fails at once.
func (failedRun) Execute(context.Context) error { return errors.New("first run failed") }

// Close releases nothing.
func (failedRun) Close() error { return nil }

// TestStartEngineFollowsRestart checks that StartEngine returns the engine of
// the run the loader starts after a run fails before serving one.
func TestStartEngineFollowsRestart(t *testing.T) {
	// Start the storage testbed.
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	tb, err := testbed.NewTestbed(ctx, logrus.NewEntry(logrus.New()), testbed.WithVerbose(false))
	if err != nil {
		t.Fatal(err)
	}
	defer tb.Release()

	// Start the engine through a factory whose first run fails.
	factory := &failFirstRunFactory{Factory: world_block_engine.NewFactory(tb.Bus)}
	conf := world_block_engine.NewConfig(
		"test-restart-engine",
		tb.Volume.GetID(), tb.BucketId,
		"test-restart-engine-store",
		&bucket.ObjectRef{BucketId: tb.BucketId},
		nil,
		false,
	)
	eng, ref, err := world.StartEngine(ctx, tb.Bus, conf, factory)
	if err != nil {
		t.Fatal(err)
	}
	defer ref.Release()

	// The engine is the second run's and accepts a write.
	if runs := factory.runs.Load(); runs < 2 {
		t.Fatalf("controller runs = %d, want at least 2", runs)
	}
	tx, err := eng.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()
	objState, err := tx.CreateObject(ctx, "restarted", nil)
	world.ReleaseObjectState(objState)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
