//go:build !js

package world_block_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aperturerobotics/util/broadcast"
	resource_world "github.com/s4wave/spacewave/core/resource/world"
	"github.com/s4wave/spacewave/db/world"
)

// viewCostEngine counts the transactions one view opens and records the
// highest seqno the view waits for.
type viewCostEngine struct {
	world.Engine

	// txs counts the transactions the view opened.
	txs atomic.Int64

	// bcast guards waitFor and fires when it rises.
	bcast broadcast.Broadcast
	// waitFor is the highest seqno the view has waited for.
	waitFor uint64
}

// NewTransaction counts the transaction and opens it on the Engine.
func (e *viewCostEngine) NewTransaction(ctx context.Context, write bool) (world.Tx, error) {
	e.txs.Add(1)
	return e.Engine.NewTransaction(ctx, write)
}

// WaitSeqno records value and waits on the Engine.
func (e *viewCostEngine) WaitSeqno(ctx context.Context, value uint64) (uint64, error) {
	e.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		if value > e.waitFor {
			e.waitFor = value
			broadcast()
		}
	})
	return e.Engine.WaitSeqno(ctx, value)
}

// waitIdle waits until the view waits for a commit after seqno.
func (e *viewCostEngine) waitIdle(ctx context.Context, t *testing.T, name string, seqno uint64) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		var waitFor uint64
		var waitCh <-chan struct{}
		e.bcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			waitFor, waitCh = e.waitFor, getWaitCh()
		})
		if waitFor > seqno {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-timeout:
			t.Fatalf("%s did not wait for a commit after seqno %d", name, seqno)
		case <-waitCh:
		}
	}
}

// viewCostCase is one kind of World watch and the transactions it may open
// on a commit to an object it did not read.
type viewCostCase struct {
	name string
	// want is the transaction count on the unrelated commit.
	want int64
	// run runs the view on eng until ctx ends.
	run func(ctx context.Context, eng world.Engine) error
}

// TestWorldViewCostOnUnrelatedCommit opens one view per kind of World watch
// on the seeded World, commits a change to an object only one of them read,
// and checks the transactions each view opens in response.
func TestWorldViewCostOnUnrelatedCommit(t *testing.T) {
	// Open an Engine on the seeded World.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	eng, cleanup := setupWorldSeededBench(ctx, t)
	defer cleanup()

	// Each view reads its own seed object, except the one that reads the
	// changed object.
	changedKey := worldSeedObjectKey(1)
	cases := []viewCostCase{{
		name: "reread-per-commit",
		want: 1,
		run: func(ctx context.Context, eng world.Engine) error {
			return rereadViewPerCommit(ctx, eng, worldSeedObjectKey(2))
		},
	}, {
		name: "object-rev-wait",
		want: 1,
		run: func(ctx context.Context, eng world.Engine) error {
			return waitObjectRevView(ctx, eng, worldSeedObjectKey(3))
		},
	}, {
		name: "tracked-object",
		want: 1,
		run: func(ctx context.Context, eng world.Engine) error {
			return runTrackedView(ctx, eng, func(ws world.WorldState) error {
				_, _, err := world.GetObjectRev(ctx, ws, worldSeedObjectKey(4))
				return err
			})
		},
	}, {
		name: "tracked-iterate",
		want: 1,
		run: func(ctx context.Context, eng world.Engine) error {
			return runTrackedView(ctx, eng, func(ws world.WorldState) error {
				iter := ws.IterateObjects(ctx, worldSeedTypeKey(0), false)
				defer iter.Close()
				for iter.Next() {
				}
				return iter.Err()
			})
		},
	}, {
		name: "tracked-graph",
		want: 1,
		run: func(ctx context.Context, eng world.Engine) error {
			return runTrackedView(ctx, eng, func(ws world.WorldState) error {
				filter := world.NewGraphQuadWithKeys(worldSeedObjectKey(5), worldSeedParentPred, "", "")
				_, err := ws.LookupGraphQuads(ctx, filter, 0)
				return err
			})
		},
	}, {
		name: "tracked-changed-object",
		want: 2,
		run: func(ctx context.Context, eng world.Engine) error {
			return runTrackedView(ctx, eng, func(ws world.WorldState) error {
				_, _, err := world.GetObjectRev(ctx, ws, changedKey)
				return err
			})
		},
	}}

	// Start every view and wait for it to settle at the current seqno.
	seqno, err := eng.GetSeqno(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	engs := make([]*viewCostEngine, len(cases))
	errs := make(chan error, len(cases))
	for i, c := range cases {
		engs[i] = &viewCostEngine{Engine: eng}
		go func() {
			if err := c.run(ctx, engs[i]); ctx.Err() == nil {
				errs <- err
			}
		}()
	}
	for i, c := range cases {
		engs[i].waitIdle(ctx, t, c.name, seqno)
	}

	// Commit a change to the changed object, then wait for every view to
	// settle after it.
	before := make([]int64, len(cases))
	for i := range engs {
		before[i] = engs[i].txs.Load()
	}
	if err := seedWorldTx(ctx, eng, func(ws world.WorldState) error {
		// Bump the changed object's revision.
		obj, err := world.MustGetObject(ctx, ws, changedKey)
		if err != nil {
			return err
		}
		defer world.ReleaseObjectState(obj)
		_, err = obj.IncrementRev(ctx)
		return err
	}); err != nil {
		t.Fatal(err.Error())
	}
	for i, c := range cases {
		engs[i].waitIdle(ctx, t, c.name, seqno+1)
	}

	// Compare the transactions each view opened for the commit.
	select {
	case err := <-errs:
		t.Fatalf("view ended: %v", err)
	default:
	}
	for i, c := range cases {
		got := engs[i].txs.Load() - before[i]
		t.Logf("%s: %d transactions", c.name, got)
		if got != c.want {
			t.Errorf("%s: %d transactions on the commit, want %d", c.name, got, c.want)
		}
	}
}

// rereadViewPerCommit rereads the object at key after every commit, as a
// watch that waits on the World seqno does.
func rereadViewPerCommit(ctx context.Context, eng world.Engine, key string) error {
	for {
		// Read the object and the seqno it was read at.
		tx, err := eng.NewTransaction(ctx, false)
		if err != nil {
			return err
		}
		seqno, err := tx.GetSeqno(ctx)
		if err == nil {
			_, _, err = world.GetObjectRev(ctx, tx, key)
		}
		tx.Discard()
		if err != nil {
			return err
		}

		// Wait for the next commit.
		if _, err := eng.WaitSeqno(ctx, seqno+1); err != nil {
			return err
		}
	}
}

// waitObjectRevView waits for each next revision of the object at key, as a
// watch over one object's revision does.
func waitObjectRevView(ctx context.Context, eng world.Engine, key string) error {
	// Read the current revision.
	rev, _, err := world.GetObjectRev(ctx, world.NewEngineWorldState(eng, false), key)
	if err != nil {
		return err
	}

	// Wait for each later revision.
	for {
		rev, err = world.WaitObjectRevBySeqno(ctx, eng, key, rev+1, false)
		if err != nil {
			return err
		}
	}
}

// runTrackedView runs read through a TrackedWorldState and runs it again
// each time a read changes, as the World state watch does.
func runTrackedView(ctx context.Context, eng world.Engine, read func(ws world.WorldState) error) error {
	for {
		if err := runTrackedViewOnce(ctx, eng, read); err != nil {
			return err
		}
	}
}

// runTrackedViewOnce runs read once and waits for one of its reads to change.
func runTrackedViewOnce(ctx context.Context, eng world.Engine, read func(ws world.WorldState) error) error {
	// Open the read transaction the view reads.
	tx, err := eng.NewTransaction(ctx, false)
	if err != nil {
		return err
	}
	defer tx.Discard()
	seqno, err := tx.GetSeqno(ctx)
	if err != nil {
		return err
	}

	// Read through the tracked state and wait for a change.
	tracked := resource_world.NewTrackedWorldState(tx, world.NewEngineWorldState(eng, false), seqno, ctx)
	defer tracked.Close()
	if err := read(tracked); err != nil {
		return err
	}
	return tracked.WaitForChanges(ctx)
}
