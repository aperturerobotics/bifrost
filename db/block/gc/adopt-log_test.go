package block_gc

import (
	"context"
	"strconv"
	"testing"

	"github.com/s4wave/spacewave/db/kvtx"
	store_kvtx_inmem "github.com/s4wave/spacewave/db/store/kvtx/inmem"
)

// scanCountTx counts the scans and iterators a transaction opens.
type scanCountTx struct {
	kvtx.Tx
	// scans counts ScanPrefix, ScanPrefixKeys and Iterate calls.
	scans int
}

// ScanPrefix counts the scan.
func (t *scanCountTx) ScanPrefix(ctx context.Context, prefix []byte, cb func(key, value []byte) error) error {
	t.scans++
	return t.Tx.ScanPrefix(ctx, prefix, cb)
}

// ScanPrefixKeys counts the scan.
func (t *scanCountTx) ScanPrefixKeys(ctx context.Context, prefix []byte, cb func(key []byte) error) error {
	t.scans++
	return t.Tx.ScanPrefixKeys(ctx, prefix, cb)
}

// Iterate counts the iterator.
func (t *scanCountTx) Iterate(ctx context.Context, prefix []byte, sort, reverse bool) kvtx.Iterator {
	t.scans++
	return t.Tx.Iterate(ctx, prefix, sort, reverse)
}

// logAdopted appends one adopt log record for blocks in its own commit and
// returns the scans it opened.
func logAdopted(t *testing.T, store kvtx.Store, blocks ...string) int {
	// Open a write transaction.
	t.Helper()
	ctx := context.Background()
	inner, err := store.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer inner.Discard()

	// Log through a counting wrapper and commit.
	tx := &scanCountTx{Tx: inner}
	if err := logAdoptedTx(ctx, tx, blocks); err != nil {
		t.Fatal(err)
	}
	if err := inner.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	return tx.scans
}

// TestAdoptLogAppendsWithoutScans checks the stall of October 22: each write
// during a pass scanned the adopt log for its last record, and in s4db that
// scan stepped past every record an earlier rescue deleted while it held the
// writer. An append now opens no scan, and a rescue consumes only the records
// logged since the last one.
func TestAdoptLogAppendsWithoutScans(t *testing.T) {
	// Open a pass and log many writes.
	ctx := context.Background()
	store := store_kvtx_inmem.NewStore()
	rg := NewRefGraph(store)
	t.Cleanup(func() { rg.Close() })
	if err := rg.BeginBlockPass(ctx); err != nil {
		t.Fatal(err)
	}
	const writes = 500
	for i := range writes {
		block := BlockIRI(testBlockRef(t, "adopt-"+strconv.Itoa(i)))
		if n := logAdopted(t, store, block); n != 0 {
			t.Fatalf("append %d opened %d scans", i, n)
		}
	}

	// A rescue consumes every record.
	mark := &BlockMark{Live: NewNodeSet()}
	rescued, err := rg.RescueAdopted(ctx, mark)
	if err != nil || rescued != writes {
		t.Fatalf("rescued %d err=%v, want %d", rescued, err, writes)
	}
	if n := countKeys(t, store, []byte{kindAdopt}); n != 0 {
		t.Fatalf("rescue left %d log records", n)
	}

	// Writes after the rescue continue the sequence without a scan, and the
	// next rescue reads only them.
	late := BlockIRI(testBlockRef(t, "adopt-late"))
	if n := logAdopted(t, store, late); n != 0 {
		t.Fatalf("append after rescue opened %d scans", n)
	}
	if n := countKeys(t, store, adoptKey(writes)); n != 1 {
		t.Fatalf("append after rescue wrote %d records at seq %d, want 1", n, writes)
	}
	rescued, err = rg.RescueAdopted(ctx, mark)
	if err != nil || rescued != 1 || !mark.Live.Has(late) {
		t.Fatalf("second rescue %d err=%v", rescued, err)
	}

	// Ending the pass leaves no record, and writes after it log nothing.
	logAdopted(t, store, BlockIRI(testBlockRef(t, "adopt-unrescued")))
	if err := rg.EndBlockPass(ctx); err != nil {
		t.Fatal(err)
	}
	logAdopted(t, store, BlockIRI(testBlockRef(t, "adopt-after")))
	if n := countKeys(t, store, []byte{kindAdopt}) + countKeys(t, store, passKey); n != 0 {
		t.Fatalf("ended pass left %d records", n)
	}
}

// TestAdoptLogLegacyPass checks a pass opened by a release whose pass record
// holds no range: the first append finds the log once and continues it, a
// rescue reads every record, and ending the pass deletes them.
func TestAdoptLogLegacyPass(t *testing.T) {
	// Open a graph over an empty store.
	ctx := context.Background()
	store := store_kvtx_inmem.NewStore()
	rg := NewRefGraph(store)
	t.Cleanup(func() { rg.Close() })

	// Write a legacy pass record and log records 3 and 4.
	old := []string{BlockIRI(testBlockRef(t, "legacy-3")), BlockIRI(testBlockRef(t, "legacy-4"))}
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Set(ctx, passKey, nil); err != nil {
		t.Fatal(err)
	}
	for i, block := range old {
		if err := tx.Set(ctx, adoptKey(uint64(3+i)), encodeChildren([]string{block})); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// The first append continues after record 4, and later ones scan nothing.
	logAdopted(t, store, BlockIRI(testBlockRef(t, "legacy-5")))
	if n := countKeys(t, store, adoptKey(5)); n != 1 {
		t.Fatalf("legacy append wrote %d records at seq 5, want 1", n)
	}
	if n := logAdopted(t, store, BlockIRI(testBlockRef(t, "legacy-6"))); n != 0 {
		t.Fatalf("append after the legacy one opened %d scans", n)
	}

	// The rescue reads all four, and ending the pass deletes the rest.
	mark := &BlockMark{Live: NewNodeSet()}
	rescued, err := rg.RescueAdopted(ctx, mark)
	if err != nil || rescued != 4 || !mark.Live.Has(old[0]) {
		t.Fatalf("rescued %d err=%v, want 4", rescued, err)
	}
	logAdopted(t, store, BlockIRI(testBlockRef(t, "legacy-7")))
	if err := rg.EndBlockPass(ctx); err != nil {
		t.Fatal(err)
	}
	if n := countKeys(t, store, []byte{kindAdopt}) + countKeys(t, store, passKey); n != 0 {
		t.Fatalf("ended legacy pass left %d records", n)
	}
}
