package world_block

import (
	"context"
	"fmt"
	"testing"

	"github.com/s4wave/spacewave/db/testbed"
	"github.com/s4wave/spacewave/db/world"
	world_mock "github.com/s4wave/spacewave/db/world/mock"
	"github.com/sirupsen/logrus"
)

// BenchmarkInsertGraphQuadsBatch commits 64 writes per World transaction, each
// creating an object with three links, like a write queue batching results.
// per-write inserts each write's links in its own call, per-write-lookup reads
// the graph before each insert, and one-call inserts all links at the end.
func BenchmarkInsertGraphQuadsBatch(b *testing.B) {
	for _, mode := range []string{"per-write", "per-write-lookup", "one-call"} {
		b.Run(mode, func(b *testing.B) {
			// Seed a World and commit batches of writes into it.
			ctx := context.Background()
			eng := newInsertBatchEngine(b, ctx)
			n := 0
			for b.Loop() {
				// Make the writes in one transaction.
				etx, err := eng.NewBlockEngineTransaction(ctx, true)
				if err != nil {
					b.Fatal(err)
				}
				err = etx.performOp(ctx, func(t *Tx) error {
					var all []world.GraphQuad
					for range 64 {
						// Create the object and build its links.
						n++
						key := fmt.Sprintf("r/%d", n)
						obj, err := t.state.CreateObject(ctx, key, nil)
						world.ReleaseObjectState(obj)
						if err != nil {
							return err
						}
						host := fmt.Sprintf("p/%d", n%2000)
						qs := []world.GraphQuad{
							world.NewGraphQuadWithKeys(key, "<type>", "p/0", ""),
							world.NewGraphQuadWithKeys(host, "<host-result>", key, ""),
							world.NewGraphQuadWithKeys(key, "<result-player>", fmt.Sprintf("p/%d", (n*7)%2000), ""),
						}
						if mode == "one-call" {
							all = append(all, qs...)
							continue
						}

						// Read the graph when asked, then insert the links.
						if mode == "per-write-lookup" {
							if _, err := t.state.LookupGraphQuads(ctx, world.NewGraphQuadWithKeys(host, "<host-result>", "", ""), 0); err != nil {
								return err
							}
						}
						if err := t.state.InsertGraphQuads(ctx, qs); err != nil {
							return err
						}
					}
					if len(all) == 0 {
						return nil
					}
					return t.state.InsertGraphQuads(ctx, all)
				})
				if err != nil {
					b.Fatal(err)
				}

				// Commit the transaction.
				if err := etx.Commit(ctx); err != nil {
					b.Fatal(err)
				}
				etx.Discard()
			}
		})
	}
}

// newInsertBatchEngine opens an engine on a World seeded with 2000 linked
// objects.
func newInsertBatchEngine(b *testing.B, ctx context.Context) *Engine {
	// Open an engine on an empty World.
	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	le := logrus.NewEntry(log)
	tb, err := testbed.NewTestbed(ctx, le, testbed.WithVerbose(false))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(tb.Release)

	// Build the engine on an empty cursor.
	base, err := tb.BuildEmptyCursor(ctx)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(base.Release)
	eng, err := NewEngine(ctx, le, base, world_mock.LookupMockOp, nil, false)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = eng.Close() })

	// Seed the objects, each linked to the one before it.
	tx, err := eng.NewTransaction(ctx, true)
	if err != nil {
		b.Fatal(err)
	}
	defer tx.Discard()
	for i := range 2000 {
		obj, err := tx.CreateObject(ctx, fmt.Sprintf("p/%d", i), nil)
		world.ReleaseObjectState(obj)
		if err != nil {
			b.Fatal(err)
		}
		if i == 0 {
			continue
		}
		if err := tx.SetGraphQuad(ctx, world.NewGraphQuadWithKeys(fmt.Sprintf("p/%d", i), "<friend>", fmt.Sprintf("p/%d", i-1), "")); err != nil {
			b.Fatal(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		b.Fatal(err)
	}
	return eng
}
