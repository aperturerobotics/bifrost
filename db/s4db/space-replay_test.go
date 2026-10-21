package s4db

import (
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"
)

// replayTestSpace returns a space of end pages whose queues hold disjoint
// pending runs in tag order, about one in three pages, and a free run past
// them.
func replayTestSpace(rnd *rand.Rand, end uint64) *space {
	// Lay the pages out as used, pending value, or pending index pages.
	sp := newSpace()
	sp.end = end
	var seq, ckpt uint64
	for p := uint64(firstPage); p < end; {
		n := 1 + rnd.Uint64N(3)
		n = min(n, end-p)
		switch rnd.IntN(3) {
		case 0:
			seq += rnd.Uint64N(2)
			sp.pend(&sp.values, pending{run: run{start: p, n: n}, tag: seq})
		case 1:
			ckpt += rnd.Uint64N(2)
			sp.pend(&sp.pages, pending{run: run{start: p, n: n}, tag: ckpt})
		}
		p += n
	}

	// Leave room to grow.
	sp.addFree(run{start: end, n: 64})
	sp.end = end + 64
	return sp
}

// TestSpaceReplayClaimMatchesCut checks that claims through the replay index
// leave the same queues and free runs as claims that scan the queues, across
// pends and releases between them.
func TestSpaceReplayClaimMatchesCut(t *testing.T) {
	for seed := range uint64(200) {
		t.Run(strconv.FormatUint(seed, 10), func(t *testing.T) {
			// Build two equal spaces and index one.
			end := firstPage + 40 + seed%200
			scan := replayTestSpace(rand.New(rand.NewPCG(seed, 1)), end)
			indexed := replayTestSpace(rand.New(rand.NewPCG(seed, 1)), end)
			indexed.indexPending()

			// Apply the same random claims, pends and releases to both.
			// A few seeds also pend a run over an already pending page,
			// which must drop the index and keep the scan's result.
			rnd := rand.New(rand.NewPCG(seed, 2))
			overlapAt := -1
			if seed%10 == 0 {
				overlapAt = rnd.IntN(200)
			}
			next, tag := scan.end, uint64(1000)
			for step := range 200 {
				if step == overlapAt && len(scan.values) != 0 {
					x := scan.values[rnd.IntN(len(scan.values))]
					tag++
					p := pending{run: run{start: x.start, n: 1}, tag: tag}
					scan.pend(&scan.pages, p)
					indexed.pend(&indexed.pages, p)
					if indexed.pendingAt != nil {
						t.Fatal("an overlapping run kept the index")
					}
				}
				switch op := rnd.IntN(10); {
				case op < 7:
					r := run{start: firstPage + rnd.Uint64N(end-firstPage), n: 1 + rnd.Uint64N(4)}
					scan.claim(r)
					indexed.claim(r)
				case op < 9:
					tag++
					p := pending{run: run{start: next, n: 1 + rnd.Uint64N(2)}, tag: tag}
					next = p.end() + rnd.Uint64N(2)
					q, iq := &scan.values, &indexed.values
					if op == 8 {
						q, iq = &scan.pages, &indexed.pages
					}
					scan.pend(q, p)
					indexed.pend(iq, p)
				default:
					seq, ckpt := rnd.Uint64N(tag+1), rnd.Uint64N(tag+1)
					scan.release(seq, ckpt)
					indexed.release(seq, ckpt)
				}
				assertSameReplaySpace(t, scan, indexed)
			}
		})
	}
}

// assertSameReplaySpace fails unless the spaces hold the same queues and free
// runs, and the index of indexed names exactly its queued runs.
func assertSameReplaySpace(t *testing.T, scan, indexed *space) {
	// Compare the queues and the free runs.
	t.Helper()
	if !slices.Equal(scan.values, indexed.values) || !slices.Equal(scan.pages, indexed.pages) {
		t.Fatalf("queues differ:\nscan    %v %v\nindexed %v %v", scan.values, scan.pages, indexed.values, indexed.pages)
	}
	if got, want := spaceFreeRuns(indexed), spaceFreeRuns(scan); !slices.Equal(got, want) {
		t.Fatalf("free runs %v, want %v", got, want)
	}

	// Compare the index with the queues, unless an overlap dropped it.
	if indexed.pendingAt == nil {
		return
	}
	want := make(map[uint64]pendingRef)
	for _, p := range indexed.values {
		want[p.start] = pendingRef{n: p.n, tag: p.tag}
	}
	for _, p := range indexed.pages {
		want[p.start] = pendingRef{n: p.n, tag: p.tag, page: true}
	}
	if indexed.pendingAt.Len() != len(want) {
		t.Fatalf("index holds %d runs, queues %d", indexed.pendingAt.Len(), len(want))
	}
	indexed.pendingAt.Scan(func(start uint64, x pendingRef) bool {
		if want[start] != x {
			t.Fatalf("index run %d %+v, queued %+v", start, x, want[start])
		}
		return true
	})
}

// spaceFreeRuns lists the free runs of sp.
func spaceFreeRuns(sp *space) []run {
	var runs []run
	sp.free.Scan(func(start, n uint64) bool {
		runs = append(runs, run{start: start, n: n})
		return true
	})
	return runs
}

// BenchmarkSpaceReplayClaim measures one replay claim of a page that was
// never pending, against queues of about as many runs as Hyperline's 3.2 GB
// of pending space at shutdown.
func BenchmarkSpaceReplayClaim(b *testing.B) {
	for _, indexed := range []bool{false, true} {
		b.Run("indexed="+strconv.FormatBool(indexed), func(b *testing.B) {
			// Pend every other page of 1.6 million, one run each.
			const runs = 800_000
			sp := newSpace()
			for i := range uint64(runs) {
				sp.pend(&sp.values, pending{run: run{start: firstPage + 2*i, n: 1}, tag: i})
			}
			sp.end = firstPage + 2*runs
			if indexed {
				sp.indexPending()
			}

			// Claim the pages between them.
			b.ResetTimer()
			for i := range b.N {
				sp.claim(run{start: firstPage + 2*uint64(i%runs) + 1, n: 1}) //nolint:gosec
			}
		})
	}
}

// BenchmarkSpaceIndexPending measures indexing the pending runs at the start
// of a replay, at the size of BenchmarkSpaceReplayClaim.
func BenchmarkSpaceIndexPending(b *testing.B) {
	// Pend every other page of 1.6 million, one run each.
	sp := newSpace()
	for i := range uint64(800_000) {
		sp.pend(&sp.values, pending{run: run{start: firstPage + 2*i, n: 1}, tag: i})
	}

	// Index them.
	b.ResetTimer()
	for range b.N {
		sp.indexPending()
	}
}
