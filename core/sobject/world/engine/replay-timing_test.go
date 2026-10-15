//go:build !js && !tinygo

package sobject_world_engine

import (
	"bytes"
	"io"
	"runtime/trace"
	"strconv"
	"testing"
	"time"

	"github.com/s4wave/spacewave/core/sobject"
	store_kvtx_inmem "github.com/s4wave/spacewave/db/store/kvtx/inmem"
	"github.com/sirupsen/logrus"
	exptrace "golang.org/x/exp/trace"
)

// TestReplayLateOperationStageTimes measures a full replay above a nonzero
// checkpoint when a late concurrent operation sorts before the cached prefix.
// Run with -v to print inclusive stage times; nested stages are not additive.
func TestReplayLateOperationStageTimes(t *testing.T) {
	// Build an isolated World with three writers and real block transforms.
	ctx := t.Context()
	seedKey, seedPeer := newReplayTestKey(t)
	keyA, peerA := newReplayTestKey(t)
	keyB, peerB := newReplayTestKey(t)
	space := newReplayTestSpace(t, seedPeer, peerA, peerB)
	space.c.le.Logger.SetLevel(logrus.InfoLevel)
	space.so.localStore = store_kvtx_inmem.NewStore()
	space.so.blockStore = &replayTimingBlockStore{BlockStore: space.so.blockStore}
	seed := space.sign("checkpoint", seedKey, "checkpoint-object", &sobject.SOOperationLink{Nonce: 1})
	seedSnap := newHeldTestSnapshot(t, space, seed)
	base, err := ReplayWorld(ctx, space.c.le, space.c.bus, space.c.sfs, space.so, "test-replay-timing", nil, seedSnap)
	if err != nil {
		t.Fatal(err)
	}

	// Cover the seed operation with a checkpoint at height one.
	data, err := base.MarshalVT()
	if err != nil {
		t.Fatal(err)
	}
	seedPosition := &sobject.SOOperationPosition{PeerId: seedPeer.String(), Nonce: 1, OpHash: seed.Hash()}
	checkpoint := &sobject.SOCheckpointInner{
		Height:    1,
		StateData: data,
		Authors:   []*sobject.SOOperationPosition{seedPosition},
	}
	snap := &replayTestSnapshot{
		config:     space.config,
		checkpoint: checkpoint,
		set:        sobject.NewSOOperationSet(replayTestObjectID, checkpoint),
	}

	// Choose the lower of two concurrent hashes as the late operation. Both
	// link only to the covered seed, so the late one must sort first.
	opA := space.sign("a", keyA, "object-a", &sobject.SOOperationLink{Nonce: 1, Parents: []*sobject.SOOperationPosition{seedPosition}})
	opB := space.sign("b", keyB, "object-b", &sobject.SOOperationLink{Nonce: 1, Parents: []*sobject.SOOperationPosition{seedPosition}})
	late, first, tailKey := opA, opB, keyB
	if bytes.Compare(opB.Hash(), opA.Hash()) < 0 {
		late, first, tailKey = opB, opA, keyA
	}
	if _, err := snap.set.Add(first); err != nil {
		t.Fatal(err)
	}

	// Chain distinct-object writes after the higher hash, all above the
	// checkpoint. A modest tail exposes per-operation cost within two minutes.
	const tailOperations = 128
	prev := first
	ops := []*sobject.SOOperation{first}
	keys := []string{"checkpoint-object", "object-a", "object-b"}
	for i := 1; i < tailOperations; i++ {
		// Sign each successor against the preceding operation.
		key := "tail-" + strconv.Itoa(i)
		op := space.sign(key, tailKey, key, &sobject.SOOperationLink{Nonce: uint64(i + 1), PrevOpHash: prev.Hash()})
		if _, err := snap.set.Add(op); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, key)
		ops = append(ops, op)
		prev = op
	}

	// Fill the persistent replay's prefix through the production public API.
	replay, err := NewWorldReplay(space.c.le, space.c.bus, space.c.sfs, space.so, "test-replay-timing", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := replay.Replay(ctx, snap); err != nil {
		t.Fatal(err)
	}

	// Deliver the late operation in a new held set and prove it invalidates
	// the entire cached prefix, excluding the covered seed.
	lateSnap := &replayTestSnapshot{config: space.config, checkpoint: checkpoint, set: sobject.NewSOOperationSet(replayTestObjectID, checkpoint)}
	for _, op := range ops {
		if _, err := lateSnap.set.Add(op); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := lateSnap.set.Add(late); err != nil {
		t.Fatal(err)
	}
	order := lateSnap.set.Order()
	if len(order) != tailOperations+1 || !bytes.Equal(order[0], late.Hash()) {
		t.Fatalf("late operation did not sort first above checkpoint: count=%d", len(order))
	}

	// Trace only the second replay, including its final durability fence.
	var buf bytes.Buffer
	if err := trace.Start(&buf); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	state, err := replay.Replay(ctx, lateSnap)
	elapsed := time.Since(start)
	trace.Stop()
	if err != nil {
		t.Fatal(err)
	}
	count := tailOperations + 1
	t.Logf("late replay: checkpoint=%d operations=%d total=%s per-op=%s", checkpoint.GetHeight(), count, elapsed, elapsed/time.Duration(count))
	if elapsed >= 2*time.Minute {
		t.Fatalf("late replay exceeded two minutes: %s", elapsed)
	}

	// Report nested stages separately and require a complete replay, including
	// every operation decode and one fence for the pass.
	stages := replayTimingStages(t, &buf)
	for _, stage := range []struct {
		// name is the runtime task type.
		name string
		// want is the exact sample count, or -1 for a positive count.
		want int
	}{
		{"alpha/so-engine/replay/prepare-op", count},
		{"alpha/so-engine/process-op/decode", count},
		{"alpha/so-engine/process-op", count},
		{"alpha/so-engine/process-op/bind-world", count},
		{"alpha/so-engine/process-apply-tx-op/new-engine-tx", count},
		{"alpha/so-engine/process-apply-tx-op/execute-tx", count},
		{"alpha/so-engine/process-apply-tx-op/commit", count},
		{"alpha/so-engine/replay/payload-refs", count},
		{"test/replay/block-read", -1},
		{"db/block/cursor/fetch/decode", -1},
		{"alpha/so-engine/replay/fence", 1},
	} {
		// Check coverage before using the stage's mean.
		got := stages[stage.name]
		if (stage.want >= 0 && got.count != stage.want) || got.count == 0 {
			t.Fatalf("%s: got %d samples, want %d", stage.name, got.count, stage.want)
		}
		t.Logf("stage=%s calls=%d total=%s per-op=%s per-call=%s (inclusive)", stage.name, got.count, got.elapsed, got.elapsed/time.Duration(count), got.elapsed/time.Duration(got.count))
	}

	// Compare the warm replay with a cold replay of exactly the same held set.
	cold, err := ReplayWorld(ctx, space.c.le, space.c.bus, space.c.sfs, space.so, "test-replay-timing-cold", nil, lateSnap)
	if err != nil {
		t.Fatal(err)
	}
	if !state.EqualVT(cold) {
		t.Fatal("late replay diverged from cold replay")
	}

	// Read through the public checkpoint engine to prove every operation
	// applied, including the operation already covered by the checkpoint.
	engine, release, err := OpenReadCheckpoint(ctx, space.c.le, space.c.bus, space.so, "test-replay-timing-read", nil, lateSnap)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	read, err := engine.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Discard()
	for _, key := range keys {
		if found, err := read.HasObject(ctx, key); err != nil || !found {
			t.Fatalf("replayed object %q: found=%v err=%v", key, found, err)
		}
	}
}

// replayTimingStage is the inclusive elapsed time and sample count of a task.
type replayTimingStage struct {
	// elapsed sums completed task durations.
	elapsed time.Duration
	// count is the number of completed tasks.
	count int
}

// replayTimingStages extracts inclusive task durations from a runtime trace.
func replayTimingStages(t *testing.T, r io.Reader) map[string]replayTimingStage {
	// Open the completed trace and correlate task starts by runtime task ID.
	t.Helper()
	reader, err := exptrace.NewReader(r)
	if err != nil {
		t.Fatal(err)
	}
	starts := make(map[exptrace.TaskID]exptrace.Event)
	stages := make(map[string]replayTimingStage)
	for {
		// Read every event so malformed or truncated traces fail the test.
		ev, err := reader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		// Sum each completed task under its original type.
		switch ev.Kind() {
		case exptrace.EventTaskBegin:
			starts[ev.Task().ID] = ev
		case exptrace.EventTaskEnd:
			if begin, ok := starts[ev.Task().ID]; ok {
				name := begin.Task().Type
				stage := stages[name]
				stage.elapsed += time.Duration(ev.Time() - begin.Time())
				stage.count++
				stages[name] = stage
				delete(starts, ev.Task().ID)
			}
		}
	}

	// Tasks opened during the measured pass must all finish before trace.Stop.
	if len(starts) != 0 {
		t.Fatalf("trace contains %d unfinished tasks", len(starts))
	}
	return stages
}
