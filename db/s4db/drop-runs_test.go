package s4db

import (
	"slices"
	"testing"
)

// TestDropRunsLarge drops half a million runs from as many reserved runs,
// which a scan per run could not finish, and keeps the rest in order.
func TestDropRunsLarge(t *testing.T) {
	const n = 500_000
	reserved := make([]run, 0, n)
	var runs, kept []run
	for i := range uint64(n) {
		r := run{start: i * 2, n: 1}
		reserved = append(reserved, r)
		if i%2 == 0 {
			runs = append(runs, r)
		} else {
			kept = append(kept, r)
		}
	}
	got := dropRuns(reserved, runs)
	if !slices.Equal(got, kept) {
		t.Fatalf("kept %d runs, want %d in order", len(got), len(kept))
	}
	if got := dropRuns([]run{{1, 1}}, nil); len(got) != 1 {
		t.Fatalf("dropping nothing kept %d runs", len(got))
	}
}
