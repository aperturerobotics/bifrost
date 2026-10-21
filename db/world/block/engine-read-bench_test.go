package world_block_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
	"github.com/s4wave/spacewave/db/testbed"
	"github.com/s4wave/spacewave/db/world"
	world_block "github.com/s4wave/spacewave/db/world/block"
	world_mock "github.com/s4wave/spacewave/db/world/mock"
	"github.com/sirupsen/logrus"
)

// BenchmarkEngineReadTxParallel opens a read transaction per operation and
// reads the bodies of a handful of objects, from many goroutines at once, as
// a server does when every connected client reads the same state after a
// revision. Run it with -mutexprofile to see the time spent waiting on the
// Engine lock.
func BenchmarkEngineReadTxParallel(b *testing.B) {
	// Open the testbed and its empty bucket cursor.
	ctx := context.Background()
	le := logrus.NewEntry(logrus.New())
	tb, err := testbed.NewTestbed(ctx, le)
	if err != nil {
		b.Fatal(err)
	}
	defer tb.Release()
	ocs, err := tb.BuildEmptyCursor(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer ocs.Release()

	// Open the Engine and write the objects every reader reads.
	eng, err := world_block.NewEngine(ctx, le, ocs, world_mock.LookupMockOp, nil, false)
	if err != nil {
		b.Fatal(err)
	}
	defer eng.Close()
	keys := writeBenchObjects(ctx, b, eng, 8)

	// Each operation opens a snapshot, reads every body and discards it.
	b.SetParallelism(16)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tx, err := eng.NewTransaction(ctx, false)
			if err != nil {
				b.Error(err)
				return
			}
			for _, key := range keys {
				_, _, err := world.AccessWorldObject(ctx, tx, key, false, func(cursor *block.Cursor) error {
					_, err := block_mock.UnmarshalExample(ctx, cursor)
					return err
				})
				if err != nil {
					b.Error(err)
				}
			}
			tx.Discard()
		}
	})
}

// writeBenchObjects writes n objects with example bodies in one transaction
// and returns their keys.
func writeBenchObjects(ctx context.Context, b *testing.B, eng *world_block.Engine, n int) []string {
	// Open the write transaction.
	wtx, err := eng.NewTransaction(ctx, true)
	if err != nil {
		b.Fatal(err)
	}

	// Create each object with its body.
	keys := make([]string, n)
	for i := range keys {
		keys[i] = "bench/read/" + strconv.Itoa(i)
		obj, _, err := world.CreateWorldObject(ctx, wtx, keys[i], func(cursor *block.Cursor) error {
			cursor.SetBlock(block_mock.NewExample("body "+keys[i]), true)
			return nil
		})
		if err != nil {
			b.Fatal(err)
		}
		world.ReleaseObjectState(obj)
	}

	// Commit them as one revision.
	if err := wtx.Commit(ctx); err != nil {
		b.Fatal(err)
	}
	return keys
}
