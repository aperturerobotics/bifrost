package s4db

import (
	"math/rand/v2"
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

// scanChoose is choose by a scan of the free runs in address order: the
// lowest run holding want when it starts below the goal, else the longest
// run below the goal holding least, else the lowest run holding want, else
// the lowest of the longest runs holding least.
func scanChoose(s *space, least, want uint64) (run, bool) {
	// Find the lowest run holding want pages, the longest run holding least,
	// and the longest holding least below the goal, the highest of equals.
	var first, longest, below run
	var hasFirst, hasLongest, hasBelow bool
	s.free.Scan(func(start, n uint64) bool {
		if n >= want && !hasFirst {
			first, hasFirst = run{start: start, n: want}, true
		}
		if n >= least && n > longest.n {
			longest, hasLongest = run{start: start, n: n}, true
		}
		if n >= least && s.goal != 0 && start < s.goal && n >= below.n {
			below, hasBelow = run{start: start, n: min(n, want)}, true
		}
		return true
	})

	// Prefer the lowest run below the goal, then the longest below it.
	switch {
	case hasFirst && (s.goal == 0 || first.start < s.goal):
		return first, true
	case hasBelow:
		return below, true
	case hasFirst:
		return first, true
	}
	return longest, hasLongest
}

// TestSpaceChooseMatchesScan checks that the size index picks the run an
// address-order scan picks, across random allocations and frees.
func TestSpaceChooseMatchesScan(t *testing.T) {
	// Allocate and free random runs, keeping the allocated ones.
	rng := rand.New(rand.NewPCG(3, 4))
	sp := newSpace()
	var held []run
	for i := range 20000 {
		if len(held) != 0 && rng.IntN(2) == 0 {
			j := rng.IntN(len(held))
			sp.addFree(held[j])
			held = slices.Delete(held, j, j+1)
		} else {
			want := 1 + rng.Uint64N(64)
			held = append(held, sp.alloc(want))
		}

		// Compare a fit of each shape against the scan, with a goal end
		// half the time.
		sp.goal = 0
		if rng.IntN(2) == 0 {
			sp.goal = rng.Uint64N(sp.end + 1)
		}
		least := 1 + rng.Uint64N(32)
		want := least + rng.Uint64N(64)
		got, gotOK := sp.choose(least, want)
		exp, expOK := scanChoose(sp, least, want)
		if got != exp || gotOK != expOK {
			t.Fatalf("step %d: choose(%d, %d) = %v %v, scan = %v %v", i, least, want, got, gotOK, exp, expOK)
		}
	}

	// The index holds exactly the free runs, and the count their pages.
	if sp.sizes.Len() != sp.free.Len() {
		t.Fatalf("sizes holds %d runs, free holds %d", sp.sizes.Len(), sp.free.Len())
	}
	var pages uint64
	sp.free.Scan(func(start, n uint64) bool {
		if _, ok := sp.sizes.Get(run{start: start, n: n}); !ok {
			t.Fatalf("sizes lacks free run %d+%d", start, n)
		}
		pages += n
		return true
	})
	if pages != sp.freeCount {
		t.Fatalf("freeCount is %d, the free runs hold %d pages", sp.freeCount, pages)
	}
}

// BenchmarkSpaceReserveFragmented measures a checkpoint reservation in a
// file whose free space is a million single pages.
func BenchmarkSpaceReserveFragmented(b *testing.B) {
	// Free every other page of a large file.
	sp := newSpace()
	base := sp.alloc(2_000_000).start
	for p := uint64(0); p < 2_000_000; p += 2 {
		sp.addFree(run{start: base + p, n: 1})
	}

	// Ask for a reservation no free run holds.
	for b.Loop() {
		if _, ok := sp.choose(placeBatch, reserveBatch); ok {
			b.Fatal("a fragmented space held a reservation")
		}
	}
}
