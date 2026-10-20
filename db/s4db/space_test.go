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
			settleCut(sp)
		}
		least := 1 + rng.Uint64N(32)
		want := least + rng.Uint64N(64)
		got, gotOK := sp.choose(least, want)
		exp, expOK := scanChoose(sp, least, want)
		if got != exp || gotOK != expOK {
			t.Fatalf("step %d: choose(%d, %d) = %v %v, scan = %v %v", i, least, want, got, gotOK, exp, expOK)
		}
	}

	// The indexes hold exactly the free runs, and the count their pages.
	checkBelow(t, sp)
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

// settleCut moves the cut until it reaches a goal.
func settleCut(sp *space) {
	for sp.goal != 0 && sp.cut != sp.goal {
		sp.moveCut()
	}
}

// checkBelow checks that below holds exactly the free runs starting below
// the cut.
func checkBelow(t *testing.T, sp *space) {
	t.Helper()
	var n int
	sp.free.Scan(func(start, length uint64) bool {
		_, ok := sp.below.Get(run{start: start, n: length})
		if ok != (start < sp.cut) {
			t.Fatalf("cut %d: below has run %d+%d = %v", sp.cut, start, length, ok)
		}
		if ok {
			n++
		}
		return true
	})
	if sp.below.Len() != n {
		t.Fatalf("below holds %d runs, %d free runs start below the cut %d", sp.below.Len(), n, sp.cut)
	}
}

// TestSpaceChooseBelowManyLonger checks that a fit with a goal takes a short
// run below it when more than goalVisits longer runs lie past it, as after
// compaction frees the end of a fragmented file.
func TestSpaceChooseBelowManyLonger(t *testing.T) {
	// Free one page low in the file, and many long runs high in it.
	sp := newSpace()
	base := sp.alloc(4 * goalVisits * 16).start
	sp.addFree(run{start: base + 8, n: 1})
	high := base + goalVisits*16
	for i := range uint64(2 * goalVisits) {
		sp.addFree(run{start: high + i*16, n: 8})
	}

	// A value fits the page below the goal.
	sp.goal = high
	settleCut(sp)
	r, ok := sp.choose(1, extentPages)
	if !ok || r.start != base+8 {
		t.Fatalf("choose = %+v %v, want the page at %d", r, ok, base+8)
	}
}

// TestSpaceMoveCutBounded checks that the cut passes at most cutRuns free
// runs a move, raising and lowering, and that below follows it.
func TestSpaceMoveCutBounded(t *testing.T) {
	// Free every other page of a file holding more than two moves of runs.
	sp := newSpace()
	base := sp.alloc(6 * cutRuns).start
	for p := uint64(0); p < 6*cutRuns; p += 2 {
		sp.addFree(run{start: base + p, n: 1})
	}

	// Raising the cut to the end takes three moves.
	sp.goal = base + 6*cutRuns
	for i := range 3 {
		if sp.cut == sp.goal {
			t.Fatalf("cut reached the goal after %d moves, want 3", i)
		}
		sp.moveCut()
		checkBelow(t, sp)
	}
	if sp.cut != sp.goal {
		t.Fatalf("cut is %d after 3 moves, want the goal %d", sp.cut, sp.goal)
	}

	// Lowering it to the start takes three moves too.
	sp.goal = base
	for range 3 {
		sp.moveCut()
		checkBelow(t, sp)
	}
	if sp.cut != sp.goal || sp.below.Len() != 0 {
		t.Fatalf("cut is %d holding %d runs, want the goal %d holding none", sp.cut, sp.below.Len(), sp.goal)
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

// TestSpaceAllocRecordGrows checks that the space record's allocation ends
// when taking its pages lengthens the encoding past them: here the lowest
// free run starts at the last page whose number fits two varint bytes, so
// taking a page from it moves its start to three.
func TestSpaceAllocRecordGrows(t *testing.T) {
	// Fill the encoding to exactly one page with runs past the lowest one.
	sp := fullRecordSpace(t)
	r, b := sp.allocRecord()

	// The record fills the pages it took, and the saved space holds them
	// in use.
	if r.start != 1<<14-1 || len(b) != int(pageOff(r.n)) || r.n != 2 {
		t.Fatalf("record %+v in %d bytes, want 2 pages from %d", r, len(b), 1<<14-1)
	}
	got, err := decodeSpace(b)
	if err != nil {
		t.Fatal(err)
	}
	if hits := got.overlaps(r); len(hits) != 0 {
		t.Fatalf("saved space holds the record's pages free: %+v", hits)
	}
}

// fullRecordSpace returns a space whose encoding fills exactly one page and
// whose lowest free run starts at page 1<<14-1.
func fullRecordSpace(t *testing.T) *space {
	// Vary the filler count and the width of the last run's length.
	t.Helper()
	for k := range pageSize / 4 {
		for _, last := range []uint64{1, 200, 20000} {
			sp := newSpace()
			sp.end = 1 << 30
			sp.setFree(run{start: 1<<14 - 1, n: 1000})
			for i := range k {
				n := uint64(1)
				if i == k-1 {
					n = last
				}
				sp.setFree(run{start: 1<<15 + uint64(i)*1<<15, n: n})
			}
			if len(sp.encode()) == pageSize {
				return sp
			}
		}
	}
	t.Fatal("no filler fills the page")
	return nil
}
