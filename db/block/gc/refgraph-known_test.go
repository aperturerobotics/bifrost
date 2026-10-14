package block_gc

import (
	"context"
	"slices"
	"strconv"
	"testing"
)

// TestSeedRefsKnown checks that seeding marks every block an edge names, and
// nothing else, across several batches, and that a rerun changes nothing.
func TestSeedRefsKnown(t *testing.T) {
	// Write a chain of blocks longer than a seeding batch under a bucket.
	ctx := context.Background()
	rg := newTestRefGraph(t)
	const chain = seedBatchRecords + seedBatchRecords/2
	blocks := make([]string, chain)
	for i := range blocks {
		blocks[i] = BlockIRI(testBlockRef(t, "seed-"+strconv.Itoa(i)))
	}
	edges := []RefEdge{{Subject: NodeGCRoot, Object: BucketIRI("bucket")}, {Subject: BucketIRI("bucket"), Object: blocks[0]}}
	for i := 1; i < chain; i++ {
		edges = append(edges, RefEdge{Subject: blocks[i-1], Object: blocks[i]})
	}
	if err := rg.ApplyRefBatch(ctx, edges, nil); err != nil {
		t.Fatal(err)
	}

	// Seed twice: the rerun after a finished seeding finds every edge written.
	for range 2 {
		if err := rg.SeedRefsKnown(ctx); err != nil {
			t.Fatal(err)
		}
	}

	// Every block has its refs known, and the other nodes do not.
	for _, node := range slices.Concat([]string{NodeGCRoot, BucketIRI("bucket")}, blocks) {
		out, err := rg.GetOutgoingRefs(ctx, node)
		if err != nil {
			t.Fatal(err)
		}
		_, want := ParseBlockIRI(node)
		if got := slices.Contains(out, NodeRefsKnown); got != want {
			t.Fatalf("%v: refs known %v, want %v", FormatIRI(node), got, want)
		}
	}

	// The refs-known node keeps no incoming records.
	in, err := rg.GetIncomingRefs(ctx, NodeRefsKnown)
	if err != nil || len(in) != 0 {
		t.Fatalf("refs-known incoming = %v/%v, want none", in, err)
	}
}
