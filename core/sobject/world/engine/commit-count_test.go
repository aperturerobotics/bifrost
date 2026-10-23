//go:build !js

package sobject_world_engine_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	storage_native "github.com/s4wave/spacewave/bldr/storage/native"
	provider "github.com/s4wave/spacewave/core/provider"
	provider_local "github.com/s4wave/spacewave/core/provider/local"
	"github.com/s4wave/spacewave/core/sobject"
	sobject_world_engine "github.com/s4wave/spacewave/core/sobject/world/engine"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/s4db"
	volume_s4db "github.com/s4wave/spacewave/db/volume/s4db"
	"github.com/s4wave/spacewave/db/world"
	world_mock "github.com/s4wave/spacewave/db/world/mock"
	"github.com/s4wave/spacewave/testbed"
	"github.com/sirupsen/logrus"
)

// localWorld is the World engine of a SharedObject in a local provider
// account whose volume is an s4db database.
type localWorld struct {
	// bus is the testbed bus the SharedObject mounts on.
	bus bus.Bus
	// le is the testbed logger the engine logs to.
	le *logrus.Entry
	// eng is the World engine.
	eng sobject_world_engine.Engine
	// db is the account volume's database.
	db *s4db.DB
	// soRef identifies the SharedObject.
	soRef *sobject.SharedObjectRef
}

// startLocalWorld starts a local provider on s4db storage under dir and the
// World engine of the SharedObject soRef in its account, creating the
// SharedObject when soRef is nil. Test cleanup stops them.
func startLocalWorld(t *testing.T, dir string, soRef *sobject.SharedObjectRef) *localWorld {
	// Open the account and its SharedObject.
	t.Helper()
	ctx := t.Context()
	tb, acc, soRef := openLocalAccount(t, dir, soRef)

	// Start its World engine with the mock operations.
	engineID := "local-world-engine"
	relOpc, err := tb.Bus.AddController(ctx, world.NewLookupOpController("local-world-ops", engineID, world_mock.LookupMockOp), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(relOpc)
	eng, engRef, err := world.StartEngine(ctx, tb.Bus, sobject_world_engine.NewConfig(engineID, soRef), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(engRef.Release)

	// Find the account database.
	return newLocalWorld(t, tb, acc, eng, soRef)
}

// openLocalAccount starts a local provider on s4db storage under dir and
// opens the SharedObject soRef in its account, creating it when soRef is nil.
// Test cleanup stops them.
func openLocalAccount(t *testing.T, dir string, soRef *sobject.SharedObjectRef) (*testbed.Testbed, provider.ProviderAccount, *sobject.SharedObjectRef) {
	// Start a testbed on s4db storage.
	t.Helper()
	ctx := t.Context()
	tb, err := testbed.Default(ctx, testbed.WithStorages(storage_native.NewS4db(false, dir)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tb.Release)

	// Run the local provider and the World engine factory on it.
	tb.StaticResolver.AddFactory(sobject_world_engine.NewFactory(tb.Bus))
	tb.StaticResolver.AddFactory(provider_local.NewFactory(tb.Bus))
	_, provRef, err := tb.Bus.AddDirective(resolver.NewLoadControllerWithConfig(&provider_local.Config{
		ProviderId: "local",
		PeerId:     tb.Volume.GetPeerID().String(),
		StorageId:  tb.StorageID,
	}), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(provRef.Release)

	// Open the account and its SharedObject.
	acc, accRef, err := provider.ExAccessProviderAccount(ctx, tb.Bus, "local", "local-world-account", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(accRef.Release)
	if soRef == nil {
		soProv, err := sobject.GetSharedObjectProviderAccountFeature(ctx, acc)
		if err != nil {
			t.Fatal(err)
		}
		soRef, err = soProv.CreateSharedObject(ctx, "local-world", &sobject.SharedObjectMeta{BodyType: "local-world"}, "", "")
		if err != nil {
			t.Fatal(err)
		}
	}
	return tb, acc, soRef
}

// newLocalWorld returns the World eng of the SharedObject soRef in the local
// account acc on tb.
func newLocalWorld(t *testing.T, tb *testbed.Testbed, acc provider.ProviderAccount, eng sobject_world_engine.Engine, soRef *sobject.SharedObjectRef) *localWorld {
	t.Helper()
	db := volume_s4db.GetDB(acc.(*provider_local.ProviderAccount).GetVolume())
	if db == nil {
		t.Fatal("account volume is not s4db")
	}
	return &localWorld{bus: tb.Bus, le: tb.Logger, eng: eng, db: db, soRef: soRef}
}

// create commits a World write that creates the object key.
func (w *localWorld) create(t *testing.T, key string) {
	// Build the write and commit it.
	t.Helper()
	tx := w.newWrite(t, key)
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
}

// newWrite returns an uncommitted World write that creates the object key.
func (w *localWorld) newWrite(t *testing.T, key string) world.Tx {
	// Open a write and create the object in it.
	t.Helper()
	tx, err := w.eng.NewTransaction(t.Context(), true)
	if err != nil {
		t.Fatal(err)
	}
	obj, err := tx.CreateObject(t.Context(), key, &bucket.ObjectRef{BucketId: "local-world-bucket"})
	world.ReleaseObjectState(obj)
	if err != nil {
		tx.Discard()
		t.Fatal(err)
	}
	return tx
}

// seq returns the sequence of the last database commit.
func (w *localWorld) seq() uint64 {
	return w.db.Seq()
}

// TestLocalWorldCommitPhysicalCommits checks that one local SharedObject World
// commit is one database commit: the blocks, the operation, the
// accepted World and the replay save publish together.
func TestLocalWorldCommitPhysicalCommits(t *testing.T) {
	// Commit one object per transaction, after a warm-up commit.
	w := startLocalWorld(t, t.TempDir(), nil)
	w.create(t, "commit-count/0")
	const n = 8
	before := w.seq()
	for i := 1; i <= n; i++ {
		w.create(t, "commit-count/"+strconv.Itoa(i))
	}
	if got := w.seq() - before; got != n {
		t.Fatalf("%d World commits made %d physical commits", n, got)
	}
}

// TestCheckpointInstallPhysicalCommits checks that a checkpoint and its
// install are two database commits: the SharedObject state, then the hold of
// the checkpoint World, the replay span and the replay save together.
func TestCheckpointInstallPhysicalCommits(t *testing.T) {
	// Write one operation short of a checkpoint and let the engine settle.
	w, so := startGatedWorld(t, t.TempDir(), nil)
	var n int
	for ; admittedOps(t, so) < sobject.MinCheckpointOperations-1; n++ {
		w.create(t, "checkpoint/"+strconv.Itoa(n))
	}
	waitSeqIdle(t, w)

	// Write the operation that reaches it, and wait for the checkpoint.
	w.create(t, "checkpoint/"+strconv.Itoa(n))
	written := w.seq()
	durable, relDurable, err := so.AccessSharedObjectState(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer relDurable()
	_, err = durable.WaitValueWithValidator(t.Context(), func(snap sobject.SharedObjectStateSnapshot) (bool, error) {
		return checkpointTrimmed(t.Context(), snap)
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// The checkpoint and its install commit twice.
	if got := waitSeqIdle(t, w) - written; got != 2 {
		t.Fatalf("checkpoint and install made %d physical commits, want 2", got)
	}
}

// waitSeqIdle waits until the database makes no commit for 300 ms and returns
// its last sequence.
func waitSeqIdle(t *testing.T, w *localWorld) uint64 {
	// Poll the sequence until it holds still.
	t.Helper()
	last, since := w.seq(), time.Now()
	deadline := since.Add(10 * time.Second)
	for time.Since(since) < 300*time.Millisecond {
		if time.Now().After(deadline) {
			t.Fatal("database kept committing")
		}
		time.Sleep(10 * time.Millisecond)
		if seq := w.seq(); seq != last {
			last, since = seq, time.Now()
		}
	}
	return last
}
