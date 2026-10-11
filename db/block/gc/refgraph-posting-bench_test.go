package block_gc

import (
	"encoding/binary"
	"strconv"
	"testing"

	cayley_kv "github.com/aperturerobotics/cayley/graph/kv"
	cayley_flat "github.com/aperturerobotics/cayley/kv/flat"
	"github.com/aperturerobotics/cayley/quad"
	store_kvtx_inmem "github.com/s4wave/spacewave/db/store/kvtx/inmem"
)

// BenchmarkRefGraphPostingChurn moves one edge per batch while keeping a
// shared subject's posting size fixed. The untouched edges stay in the graph.
func BenchmarkRefGraphPostingChurn(b *testing.B) {
	for _, count := range []int{1_000, 100_000} {
		b.Run("edges="+strconv.Itoa(count), func(b *testing.B) {
			// Open an isolated reference graph for the posting workload.
			ctx := b.Context()
			rg, err := NewRefGraph(ctx, store_kvtx_inmem.NewStore(), []byte("gc/"))
			if err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() {
				if err := rg.Close(); err != nil {
					b.Error(err)
				}
			})

			// Prepare one bucket's edges and other owners for the moving targets.
			subject := BucketIRI("posting-benchmark")
			first := []RefEdge{{Subject: subject, Object: "posting/first"}}
			second := []RefEdge{{Subject: subject, Object: "posting/second"}}
			seed := make([]RefEdge, 0, count+2)
			seed = append(seed, first...)
			for i := range count - 1 {
				seed = append(seed, RefEdge{
					Subject: subject,
					Object:  "posting/static/" + strconv.Itoa(i),
				})
			}
			seed = append(seed,
				RefEdge{Subject: "posting/keeper", Object: first[0].Object},
				RefEdge{Subject: "posting/keeper", Object: second[0].Object},
			)

			// Seed the posting through the production graph API.
			if err := rg.ApplyRefBatch(ctx, seed, nil); err != nil {
				b.Fatal(err)
			}
			requireRefGraphPostingSize(b, rg, subject, count)

			// Alternate one addition and one removal without growing orphan markers.
			var graph RefGraphOps = rg
			adds, removes := second, first
			b.ReportAllocs()
			for b.Loop() {
				if err := graph.ApplyRefBatch(ctx, adds, removes); err != nil {
					b.Fatal(err)
				}
				adds, removes = removes, adds
			}

			// Require the stream to preserve its posting and final edge membership.
			requireRefGraphPostingSize(b, rg, subject, count)
			found, err := rg.hasRefs(ctx, append(adds, removes...))
			if err != nil {
				b.Fatal(err)
			}
			if found[0] || !found[1] {
				b.Fatalf("final absent/present membership = %v", found)
			}
		})
	}
}

// requireRefGraphPostingSize verifies that the fixture holds one large posting,
// rather than the same number of edges distributed across small posting keys.
func requireRefGraphPostingSize(b *testing.B, rg *RefGraph, subject string, want int) {
	// Resolve the subject through the graph naming API.
	b.Helper()
	ctx := b.Context()
	subjectRef, err := rg.handle.ValueOf(ctx, quad.IRI(subject))
	if err != nil {
		b.Fatal(err)
	}

	// Resolve the predicate through the graph naming API.
	predRef, err := rg.handle.ValueOf(ctx, quad.IRI(PredGCRef))
	if err != nil {
		b.Fatal(err)
	}

	// Read the exact shared forward posting under one storage transaction.
	subjectID := uint64(subjectRef.(cayley_kv.Int64Value))
	predID := uint64(predRef.(cayley_kv.Int64Value))
	key := cayley_flat.KeyEscape(cayley_kv.DefaultQuadIndexes[0].Key([]uint64{subjectID, predID}))
	tx, err := rg.store.NewTransaction(ctx, false)
	if err != nil {
		b.Fatal(err)
	}
	defer tx.Discard()

	// Count encoded entries without allocating a decoded posting list.
	data, found, err := tx.Get(ctx, key)
	if err != nil {
		b.Fatal(err)
	}
	if !found {
		b.Fatal("shared forward posting is missing")
	}
	var count int
	for len(data) != 0 {
		_, n := binary.Uvarint(data)
		if n <= 0 {
			b.Fatal("shared forward posting contains an invalid quad ID")
		}
		count++
		data = data[n:]
	}
	if count != want {
		b.Fatalf("shared forward posting holds %d edges, want %d", count, want)
	}
}
