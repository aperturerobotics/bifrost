package kvtx_deferred

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/s4wave/spacewave/db/kvtx"
	kvtx_block "github.com/s4wave/spacewave/db/kvtx/block"
	kvtx_kvtest "github.com/s4wave/spacewave/db/kvtx/kvtest"
	"github.com/s4wave/spacewave/db/testbed"
	"github.com/sirupsen/logrus"
)

// buildOkraOps opens deferred ops over an empty graph okra tree.
func buildOkraOps(t *testing.T, maxEntries int) (*Ops, kvtx.BlockTx) {
	// Start a testbed with an empty block cursor.
	ctx := t.Context()
	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	tb, err := testbed.NewTestbed(ctx, logrus.NewEntry(log), testbed.WithVerbose(false))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tb.Release)

	// Build an empty cursor to hold the tree.
	bcs, err := tb.BuildEmptyCursor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bcs.Release)

	// Open a write transaction on an empty graph tree.
	_, cursor := bcs.BuildTransaction(nil)
	cursor.SetBlock(kvtx_block.NewKeyValueStoreForWorkload(kvtx_block.WorkloadClassGraphPrefixRead), true)
	ktx, err := kvtx_block.BuildKvTransaction(ctx, cursor, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ktx.Discard)

	// Hold the writes over the tree's write batches.
	batch, ok := ktx.(kvtx.WriteBatchTxOps)
	if !ok {
		t.Fatalf("%T does not support write batches", ktx)
	}
	return NewOps(ktx, batch, maxEntries, 0), ktx
}

// TestOpsKVTest runs the shared store checks through held writes.
func TestOpsKVTest(t *testing.T) {
	// Run the checks against a store over the deferred ops.
	ops, _ := buildOkraOps(t, 0)
	if err := kvtx_kvtest.TestAll(t.Context(), kvtx.NewTxStore(ops)); err != nil {
		t.Fatal(err)
	}
}

// TestOpsMergedReads checks reads through held writes against a map, across
// random writes, deletes and flushes.
func TestOpsMergedReads(t *testing.T) {
	// Open the ops and a seeded key source.
	ctx := t.Context()
	ops, ktx := buildOkraOps(t, 64)
	rng := rand.New(rand.NewPCG(1, 2))
	want := make(map[string]string)
	key := func() []byte { return fmt.Appendf(nil, "%c/%03d", 'a'+rng.IntN(3), rng.IntN(200)) }
	for round := range 2000 {
		// Write, delete, or flush.
		switch r := rng.IntN(20); {
		case r == 0:
			if err := ops.Flush(ctx); err != nil {
				t.Fatal(err)
			}
		case r < 6:
			k := key()
			if err := ops.Delete(ctx, k); err != nil {
				t.Fatal(err)
			}
			delete(want, string(k))
		default:
			k, v := key(), fmt.Sprintf("v%d", round)
			if err := ops.Set(ctx, k, []byte(v)); err != nil {
				t.Fatal(err)
			}
			want[string(k)] = v
		}
		if round%50 != 0 {
			continue
		}

		// Read one key, a batch, and a prefix iteration.
		k := key()
		v, found, err := ops.Get(ctx, k)
		if err != nil || found != (want[string(k)] != "") || string(v) != want[string(k)] {
			t.Fatalf("round %d: get %s = %q %v %v, want %q", round, k, v, found, err, want[string(k)])
		}
		keys := [][]byte{key(), key(), key()}
		vals, founds, err := ops.GetBatch(ctx, keys)
		if err != nil {
			t.Fatal(err)
		}
		for i, k := range keys {
			if founds[i] != (want[string(k)] != "") || string(vals[i]) != want[string(k)] {
				t.Fatalf("round %d: batch %s = %q %v, want %q", round, k, vals[i], founds[i], want[string(k)])
			}
		}
		prefix := fmt.Appendf(nil, "%c/", 'a'+rng.IntN(3))
		checkIterate(t, ctx, ops, prefix, nil, want)
		checkIterate(t, ctx, ops, prefix, fmt.Appendf(prefix, "%03d", rng.IntN(200)), want)
	}

	// The flushed tree matches the map.
	if err := ops.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if ops.Held() != 0 {
		t.Fatalf("held %d after flush", ops.Held())
	}
	got := make(map[string]string)
	err := ktx.ScanPrefix(ctx, nil, func(k, v []byte) error {
		got[string(k)] = string(v)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(got, want) {
		t.Fatalf("tree has %d keys, want %d", len(got), len(want))
	}
}

// checkIterate checks a sorted forward iteration from seek against want.
func checkIterate(t *testing.T, ctx context.Context, ops *Ops, prefix, seek []byte, want map[string]string) {
	// List the expected entries in key order.
	t.Helper()
	var exp []string
	for _, k := range slices.Sorted(maps.Keys(want)) {
		if bytes.HasPrefix([]byte(k), prefix) && k >= string(seek) {
			exp = append(exp, k+"="+want[k])
		}
	}

	// Position the iterator at the start or at seek.
	it := ops.Iterate(ctx, prefix, true, false)
	defer it.Close()
	var got []string
	if seek != nil {
		if err := it.Seek(seek); err != nil {
			t.Fatal(err)
		}
	} else {
		it.Next()
	}

	// Collect the entries and compare them.
	for ; it.Valid(); it.Next() {
		v, err := it.Value()
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, string(it.Key())+"="+string(v))
	}
	if err := it.Err(); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, exp) {
		t.Fatalf("iterate %q from %q:\n got %v\nwant %v", prefix, seek, got, exp)
	}
}

// TestOpsFlushOnBound checks that held writes apply once they pass the bound.
func TestOpsFlushOnBound(t *testing.T) {
	// Hold three writes, one under the bound.
	ctx := t.Context()
	ops, ktx := buildOkraOps(t, 4)
	for i := range 3 {
		if err := ops.Set(ctx, fmt.Appendf(nil, "k%d", i), []byte("v")); err != nil {
			t.Fatal(err)
		}
	}
	if _, found, _ := ktx.Get(ctx, []byte("k0")); found || ops.Held() != 3 {
		t.Fatalf("applied before the bound: found %v held %d", found, ops.Held())
	}
	if err := ops.Set(ctx, []byte("k3"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := ktx.Get(ctx, []byte("k0")); !found || ops.Held() != 0 {
		t.Fatalf("not applied at the bound: found %v held %d", found, ops.Held())
	}
}
