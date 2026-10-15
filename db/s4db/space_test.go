package s4db

import (
	"slices"
	"testing"
)

// TestSpaceDrainSkipsRetaken checks that drain returns only the released
// pages that are still free, so a page taken again is never punched.
func TestSpaceDrainSkipsRetaken(t *testing.T) {
	// Allocate pages, then free a run in the middle.
	sp := newSpace()
	sp.alloc(32)
	start := uint64(firstPage + 8)
	sp.addFree(run{start: start, n: 8})

	// Take the middle of the freed run again.
	sp.take(run{start: start + 2, n: 3})

	// Only the parts on either side may be punched.
	got := sp.drain()
	slices.SortFunc(got, func(a, b run) int { return int(a.start) - int(b.start) })
	want := []run{{start: start, n: 2}, {start: start + 5, n: 3}}
	if !slices.Equal(got, want) {
		t.Fatalf("drain = %v, want %v", got, want)
	}
	if rest := sp.drain(); len(rest) != 0 {
		t.Fatalf("second drain = %v, want none", rest)
	}
}

// TestSpaceDrainMergesRuns checks that drain returns runs released page by
// page as one run, so each free stretch costs one punch.
func TestSpaceDrainMergesRuns(t *testing.T) {
	// Allocate pages, then free eight of them one at a time, out of order.
	sp := newSpace()
	sp.alloc(32)
	start := uint64(firstPage + 8)
	for _, i := range []uint64{3, 0, 7, 1, 2, 6, 4, 5} {
		sp.addFree(run{start: start + i, n: 1})
	}

	// The released pages drain as one run.
	got := sp.drain()
	want := []run{{start: start, n: 8}}
	if !slices.Equal(got, want) {
		t.Fatalf("drain = %v, want %v", got, want)
	}
	if sp.releasedPages != 0 {
		t.Fatalf("releasedPages = %d after drain, want 0", sp.releasedPages)
	}
}
