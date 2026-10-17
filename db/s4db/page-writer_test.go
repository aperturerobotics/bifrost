package s4db

import (
	"testing"
)

// TestPageWriterTurns checks that the page writer writes each reserved run
// once it is full and reserves the next run in the same turn, so a
// background checkpoint takes the writer lock once per run.
func TestPageWriterTurns(t *testing.T) {
	// Count the turns and the pages each writes, handing out runs of
	// reserveBatch pages.
	const pages = 2*reserveBatch + 100
	var turns, written int
	var end uint64
	w := &pageWriter{cache: newCache(64 << 20)}
	w.turn = func(buf []byte, at, n uint64) (run, error) {
		// Count the turn and check a reserving turn wrote a full run.
		turns++
		written += len(buf) / pageSize
		if len(buf) != 0 && n != 0 && at+uint64(len(buf)/pageSize) != end {
			t.Fatalf("turn wrote pages %d to %d before the run ended at %d", at, at+uint64(len(buf)/pageSize), end)
		}

		// Hand out the next run.
		if n == 0 {
			return run{}, nil
		}
		r := run{start: end + 1, n: n}
		end = r.end()
		return r, nil
	}

	// Add the pages and flush the last run.
	d := &draft{leaf: true, keys: [][]byte{[]byte("k")}, vals: []value{{inline: []byte("v")}}}
	for range pages {
		if _, err := w.add(d); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.flush(); err != nil {
		t.Fatal(err)
	}

	// Three runs take three reserving turns and one final write.
	if turns != 4 || written != pages {
		t.Fatalf("%d turns wrote %d pages, want 4 turns writing %d", turns, written, pages)
	}
}
