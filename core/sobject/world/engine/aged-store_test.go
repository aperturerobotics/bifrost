//go:build !js

package sobject_world_engine_test

import (
	"context"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime/pprof"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/s4wave/spacewave/bldr/storage"
	storage_native "github.com/s4wave/spacewave/bldr/storage/native"
	provider "github.com/s4wave/spacewave/core/provider"
	provider_local "github.com/s4wave/spacewave/core/provider/local"
	"github.com/s4wave/spacewave/core/sobject"
	sobject_world_engine "github.com/s4wave/spacewave/core/sobject/world/engine"
	"github.com/s4wave/spacewave/db/bucket"
	volume_controller "github.com/s4wave/spacewave/db/volume/controller"
	"github.com/s4wave/spacewave/db/world"
	world_mock "github.com/s4wave/spacewave/db/world/mock"
	"github.com/s4wave/spacewave/testbed"
	"github.com/sirupsen/logrus"
)

// holdRecorder collects every World writer lock hold the engine logs.
type holdRecorder struct {
	// mtx guards holds.
	mtx sync.Mutex
	// holds are the hold durations by holder name.
	holds map[string][]time.Duration
}

// Levels returns the levels the hold log uses.
func (r *holdRecorder) Levels() []logrus.Level { return []logrus.Level{logrus.WarnLevel} }

// Fire records a writer lock hold entry.
func (r *holdRecorder) Fire(e *logrus.Entry) error {
	// A slow commit entry carries each commit stage as name=duration.
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if e.Message == "slow World commit" {
		stages, _ := e.Data["stages"].(string)
		for _, f := range strings.Fields(stages) {
			name, v, _ := strings.Cut(f, "=")
			if d, err := time.ParseDuration(v); err == nil {
				r.holds["stage:"+name] = append(r.holds["stage:"+name], d)
			}
		}
		return nil
	}

	// A slow hold entry carries the holder name and the hold duration.
	if e.Message == "slow World writer lock hold" {
		holder, _ := e.Data["holder"].(string)
		held, _ := e.Data["held"].(time.Duration)
		r.holds[holder] = append(r.holds[holder], held)
	}
	return nil
}

// take returns the holds recorded since the last take and clears them.
func (r *holdRecorder) take() map[string][]time.Duration {
	// Swap in an empty map under the lock.
	r.mtx.Lock()
	defer r.mtx.Unlock()
	holds := r.holds
	r.holds = make(map[string][]time.Duration)
	return holds
}

// logHolds logs the count, median and longest hold of each holder.
func logHolds(t *testing.T, phase string, holds map[string][]time.Duration) {
	// Log the holders in name order.
	t.Helper()
	names := make([]string, 0, len(holds))
	for name := range holds {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		d := holds[name]
		slices.Sort(d)
		t.Logf("%s holder=%s n=%d p50=%s p99=%s max=%s", phase, name, len(d), d[len(d)/2], d[len(d)*99/100], d[len(d)-1])
	}
}

// gcStorage stores volumes like Storage with garbage collection every
// interval, so block passes run within a test's lifetime.
type gcStorage struct {
	storage.Storage
	// interval is the volume collection interval.
	interval string
}

// BuildVolumeConfig builds the Storage's volume config with the interval.
func (s *gcStorage) BuildVolumeConfig(id string, base *volume_controller.Config) (config.Config, error) {
	base = base.CloneVT()
	if base == nil {
		base = &volume_controller.Config{}
	}
	base.GcIntervalDur = s.interval
	return s.Storage.BuildVolumeConfig(id, base)
}

// dumpStall writes every goroutine's stack to AGED_STORE_DUMP once write i
// has been pending AGED_STORE_DUMP_AFTER (default 5s), so a stall shows what
// the write waited on. It returns the function that ends the wait.
func dumpStall(t *testing.T, i int) func() {
	// Dump only when a directory is set.
	dir := os.Getenv("AGED_STORE_DUMP")
	if dir == "" {
		return func() {}
	}
	after := 5 * time.Second
	if v := os.Getenv("AGED_STORE_DUMP_AFTER"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			t.Fatalf("AGED_STORE_DUMP_AFTER: %v", err)
		}
		after = d
	}

	// Write the dump when the timer fires before the write ends.
	timer := time.AfterFunc(after, func() {
		f, err := os.Create(filepath.Join(dir, "stall-"+strconv.Itoa(i)+".txt"))
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()
		if err := pprof.Lookup("goroutine").WriteTo(f, 2); err != nil {
			t.Error(err)
		}
	})
	return func() { timer.Stop() }
}

// envInt reads a positive integer from the environment, or def.
func envInt(t *testing.T, key string, def int) int {
	// Parse the variable when it is set.
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		t.Fatalf("%s: %q is not a positive integer", key, v)
	}
	return n
}

// TestAgedStoreWriterHolds ages a SharedObject World on an s4db volume with
// churning writes, then measures the writer lock holds of a fixed probe of
// writes. A hold that grows with the store's history shows as a probe hold
// that grows with AGED_STORE_WRITES while AGED_STORE_PROBE stays fixed.
func TestAgedStoreWriterHolds(t *testing.T) {
	// Read the workload sizes from the environment.
	ageWrites := envInt(t, "AGED_STORE_WRITES", 200)
	probeWrites := envInt(t, "AGED_STORE_PROBE", 100)
	population := envInt(t, "AGED_STORE_OBJECTS", 2000)
	perWrite := envInt(t, "AGED_STORE_PER_WRITE", 8)
	dir := os.Getenv("AGED_STORE_DIR")
	if dir == "" {
		dir = t.TempDir()
	}

	// Record every writer lock hold.
	defer sobject_world_engine.SetSlowHoldThreshold(0)()
	defer sobject_world_engine.SetSlowCommitThreshold(0)()

	// Start the testbed on s4db storage, collecting volume garbage at the
	// AGED_STORE_GC interval when it is set.
	ctx := t.Context()
	store := storage_native.NewS4db(false, dir)
	if gc := os.Getenv("AGED_STORE_GC"); gc != "" {
		store = &gcStorage{Storage: store, interval: gc}
	}
	tb, err := testbed.Default(ctx, testbed.WithStorages(store))
	if err != nil {
		t.Fatal(err)
	}
	defer tb.Release()

	// Silence the log and hook the recorder into it.
	tb.Logger.Logger.SetLevel(logrus.WarnLevel)
	tb.Logger.Logger.SetOutput(io.Discard)
	rec := &holdRecorder{holds: make(map[string][]time.Duration)}
	tb.Logger.Logger.AddHook(rec)

	// Register the engine and local provider factories.
	tb.StaticResolver.AddFactory(sobject_world_engine.NewFactory(tb.Bus))
	tb.StaticResolver.AddFactory(provider_local.NewFactory(tb.Bus))

	// Load the local provider on the s4db storage.
	_, provCtrlRef, err := tb.Bus.AddDirective(resolver.NewLoadControllerWithConfig(&provider_local.Config{
		ProviderId: "local",
		PeerId:     tb.Volume.GetPeerID().String(),
		StorageId:  tb.StorageID,
	}), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer provCtrlRef.Release()

	// Access the local provider account.
	provAcc, provAccRef, err := provider.ExAccessProviderAccount(ctx, tb.Bus, "local", "aged-store-account", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer provAccRef.Release()

	// Create the SharedObject.
	soProv, err := sobject.GetSharedObjectProviderAccountFeature(ctx, provAcc)
	if err != nil {
		t.Fatal(err)
	}
	soRef, err := soProv.CreateSharedObject(ctx, "aged-store", &sobject.SharedObjectMeta{BodyType: "aged-store"}, "", "")
	if err != nil {
		t.Fatal(err)
	}

	// Register the mock operations and start the SharedObject's World engine.
	engineID := "aged-store-engine"
	opc := world.NewLookupOpController("aged-store-ops", engineID, world_mock.LookupMockOp)
	relOpc, err := tb.Bus.AddController(ctx, opc, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer relOpc()
	eng, engRef, err := world.StartEngine(ctx, tb.Bus, sobject_world_engine.NewConfig(engineID, soRef), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer engRef.Release()

	// Each write replaces perWrite random objects of the population and moves
	// a graph edge between them, so World blocks churn and the store keeps
	// dead space.
	rng := rand.New(rand.NewPCG(1, 2))
	write := func(ctx context.Context, i int) {
		// Dump every goroutine once a write has been pending a while.
		defer dumpStall(t, i)()

		err := world.ExecTransaction(ctx, eng, true, func(ctx context.Context, ws world.WorldState) error {
			for j := range perWrite {
				key := "aged/" + strconv.Itoa(rng.IntN(population))
				if _, err := ws.DeleteObject(ctx, key); err != nil {
					return err
				}
				obj, err := ws.CreateObject(ctx, key, &bucket.ObjectRef{BucketId: "aged-" + strconv.Itoa(i) + "-" + strconv.Itoa(j)})
				world.ReleaseObjectState(obj)
				if err != nil {
					return err
				}
				peer := "aged/" + strconv.Itoa(rng.IntN(population))
				if peer != key {
					if err := ws.SetGraphQuad(ctx, world.NewGraphQuadWithKeys(key, "aged/link", peer, "")); err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create the population in batches.
	for lo := 0; lo < population; lo += 500 {
		err := world.ExecTransaction(ctx, eng, true, func(ctx context.Context, ws world.WorldState) error {
			for k := lo; k < min(lo+500, population); k++ {
				obj, err := ws.CreateObject(ctx, "aged/"+strconv.Itoa(k), &bucket.ObjectRef{BucketId: "aged"})
				world.ReleaseObjectState(obj)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Age the store, discarding the population's holds first.
	rec.take()
	start := time.Now()
	for i := range ageWrites {
		write(ctx, i)
	}
	t.Logf("aged %d writes in %s", ageWrites, time.Since(start))
	logHolds(t, "age", rec.take())

	// Probe the aged store with a fixed number of writes.
	start = time.Now()
	for i := range probeWrites {
		write(ctx, ageWrites+i)
	}
	t.Logf("probed %d writes in %s", probeWrites, time.Since(start))
	logHolds(t, "probe", rec.take())

	// Report the store's file sizes.
	if info, err := os.ReadDir(dir); err == nil {
		for _, e := range info {
			if fi, err := e.Info(); err == nil {
				t.Logf("store file %s %d bytes", e.Name(), fi.Size())
			}
		}
	}
}
