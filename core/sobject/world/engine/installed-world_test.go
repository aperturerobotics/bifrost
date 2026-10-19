package sobject_world_engine

import (
	"bytes"
	"context"
	"io"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/util/ccontainer"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/block"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
	block_transform "github.com/s4wave/spacewave/db/block/transform"
	"github.com/s4wave/spacewave/db/bucket"
	store_kvtx_inmem "github.com/s4wave/spacewave/db/store/kvtx/inmem"
	"github.com/s4wave/spacewave/db/world"
	world_block_tx "github.com/s4wave/spacewave/db/world/block/tx"
	world_mock "github.com/s4wave/spacewave/db/world/mock"
	"github.com/sirupsen/logrus"
)

// warnLog collects the messages logged at warning level.
type warnLog struct {
	mtx  sync.Mutex
	msgs []string
}

// Levels returns the level the hook collects.
func (l *warnLog) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel}
}

// Fire collects a message.
func (l *warnLog) Fire(entry *logrus.Entry) error {
	l.mtx.Lock()
	l.msgs = append(l.msgs, entry.Message)
	l.mtx.Unlock()
	return nil
}

// has reports whether a collected message equals msg.
func (l *warnLog) has(msg string) bool {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return slices.Contains(l.msgs, msg)
}

// TestEngineKeepsInstalledWorldWhenReplayStopsBehindIt replays an operation
// from another writer that sorts before operations the engine already
// installed and needs a block that is not available. Replay stops at it, and
// the World it would reach lacks the installed operations, so the engine must
// keep the installed World, which readers see, and log the stop. Once the block
// arrives, the engine installs the World with every operation, in the settled
// order.
func TestEngineKeepsInstalledWorldWhenReplayStopsBehindIt(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	privA, pidA := newReplayTestKey(t)
	privB, pidB := newReplayTestKey(t)
	space := newReplayTestSpace(t, pidA, pidB)
	space.c.SetStaticLookupOp(world_mock.LookupMockOp)

	// Write the block the second writer's edit reads.
	transformConf := space.genesis.GetHeadRef().GetTransformConf()
	xfrm, err := block_transform.NewTransformer(controller.ConstructOpts{}, space.c.sfs, transformConf)
	if err != nil {
		t.Fatal(err.Error())
	}
	btx, bcs := block.NewTransaction(space.so.blockStore, xfrm, nil, nil)
	bcs.SetBlock(&block_mock.Example{Msg: "payload"}, true)
	payload, _, err := btx.Write(ctx, true)
	if err != nil {
		t.Fatal(err.Error())
	}
	ghost, err := world_block_tx.NewTxCreateObject("ghost", &bucket.ObjectRef{RootRef: payload, TransformConf: transformConf})
	if err != nil {
		t.Fatal(err.Error())
	}
	edit, err := world_block_tx.NewTxApplyObjectOp(world_mock.MockObjectOpTypeID, world_mock.NewMockObjectOp("edited"), "ghost", pidB)
	if err != nil {
		t.Fatal(err.Error())
	}

	// The first writer's two operations are installed. The second writer's
	// concurrent chain ends at the edit, which needs the hidden block.
	checkpointData, err := space.genesis.MarshalVT()
	if err != nil {
		t.Fatal(err.Error())
	}
	checkpoint := &sobject.SOCheckpointInner{StateData: checkpointData}
	a1 := space.sign("a1", privA, "object-a1", &sobject.SOOperationLink{Nonce: 1})
	a2 := space.sign("a2", privA, "object-a2", &sobject.SOOperationLink{Nonce: 2, PrevOpHash: a1.Hash()})
	set := sobject.NewSOOperationSet(replayTestObjectID, checkpoint)
	for _, op := range []*sobject.SOOperation{a1, a2} {
		if _, err := set.Add(op); err != nil {
			t.Fatal(err.Error())
		}
	}

	// Sign the second writer's chain until it sorts before the installed
	// operations, as another writer's operations do half of the time.
	var b1, b2 *sobject.SOOperation
	for range 64 {
		b1 = space.signTx("b1", privB, ghost, &sobject.SOOperationLink{Nonce: 1})
		b2 = space.signTx("b2", privB, edit, &sobject.SOOperationLink{Nonce: 2, PrevOpHash: b1.Hash()})
		if bytes.Compare(b1.Hash(), a1.Hash()) < 0 {
			break
		}
	}
	if bytes.Compare(b1.Hash(), a1.Hash()) >= 0 {
		t.Fatal("could not sign an operation that sorts before the installed ones")
	}

	// Install the World of the first writer's operations.
	snap := &replayTestSnapshot{config: space.config, checkpoint: checkpoint, set: set}
	store := &hidingBlockStore{
		BlockStore: space.so.blockStore,
		ref:        payload,
		peer:       make(chan struct{}),
		asked:      make(chan struct{}, 1),
	}
	so := &testSharedObject{peerID: pidA, blockStore: store, localStore: store_kvtx_inmem.NewStore(), snapshot: snap}
	stateCtr := ccontainer.NewCContainer[sobject.SharedObjectStateSnapshot](snap)
	head, err := space.c.waitWorldInit(ctx, so, stateCtr, newReplayer(space.c, so))
	if err != nil {
		t.Fatal(err.Error())
	}
	blk, err := buildBlockEngine(ctx, space.c.le, space.c.bus, space.c.sfs, so, head.GetHeadRef(), transformConf, space.c.buildLookupWorldOp(space.c.le), false)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(blk.Release)
	engine := newSoEngine(ctx, space.c, so, blk, newReplayer(space.c, so))
	warnings := &warnLog{}
	space.c.le.Logger.AddHook(warnings)
	space.c.le.Logger.SetOutput(io.Discard)

	// advance replays the set as the watcher does, under the writer lock.
	advance := func() error {
		t.Helper()
		unlock, err := space.c.writeMtx.Lock(ctx)
		if err != nil {
			t.Fatal(err.Error())
		}
		defer unlock()
		_, err = engine.advance(ctx, snap, nil)
		return err
	}

	// readable fails the test unless the World holds exactly the keys that
	// want names among the keys the test writes.
	readable := func(when string, want ...string) {
		t.Helper()
		read, err := engine.NewTransaction(ctx, false)
		if err != nil {
			t.Fatal(err.Error())
		}
		defer read.Discard()
		for _, key := range []string{"object-a1", "object-a2", "ghost"} {
			objectState, found, err := read.GetObject(ctx, key)
			world.ReleaseObjectState(objectState)
			if err != nil {
				t.Fatal(err.Error())
			}
			if found != slices.Contains(want, key) {
				t.Fatalf("%s: object %s found = %v; want %v", when, key, found, !found)
			}
		}
	}
	if err := advance(); err != nil {
		t.Fatalf("install the first writer's operations: %v", err)
	}
	readable("installed", "object-a1", "object-a2")

	// The second writer's operations arrive and sort first, but the block
	// the edit reads is not available.
	for _, op := range []*sobject.SOOperation{b1, b2} {
		if _, err := set.Add(op); err != nil {
			t.Fatal(err.Error())
		}
	}
	store.hidden.Store(true)
	for range 2 {
		if err := advance(); !block.IsNotAvailable(err) {
			t.Fatalf("replay with a missing block = %v; want it unavailable", err)
		}
		readable("replay stopped behind the installed World", "object-a1", "object-a2")
		if !engine.replay.missing.EqualsRef(payload) {
			t.Fatalf("replay stopped at %v; want %s", engine.replay.missing, payload.MarshalString())
		}
	}
	if !warnings.has("replay stopped behind the served World, keeping the served World") {
		t.Fatalf("replay logged %q; want it to say it kept the served World", warnings.msgs)
	}

	// Once the block arrives, every operation is installed.
	store.serve()
	if err := advance(); err != nil {
		t.Fatalf("replay after the block arrived: %v", err)
	}
	readable("replayed", "object-a1", "object-a2", "ghost")
}
