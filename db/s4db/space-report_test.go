//go:build darwin || linux || windows

package s4db

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSpaceReport checks that the report's free runs match the usage, that
// the live values end inside the file, and that the handle's own slot shows.
func TestSpaceReport(t *testing.T) {
	// Fill a file with free space below live values.
	ctx := context.Background()
	opts := Options{RelocateBudget: -1, SingleProcess: true}
	db := openTest(t, filepath.Join(t.TempDir(), "report.s4wave"), opts)
	defer db.Close()
	fillSlack(t, ctx, db)

	// Read the usage and the report of the same state.
	u, err := db.Usage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	r, err := db.SpaceReport(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// The free run classes add up to the free bytes, each run in its class.
	var runs, free int64
	for _, c := range r.FreeRuns {
		if c.Runs == 0 || c.Bytes < c.Runs*c.MinBytes || c.Bytes >= 2*c.Runs*c.MinBytes {
			t.Fatalf("class %+v holds runs outside it", c)
		}
		runs += c.Runs
		free += c.Bytes
	}
	if runs == 0 || free != u.FreeBytes {
		t.Fatalf("free runs %d sum to %d, want FreeBytes %d", runs, free, u.FreeBytes)
	}
	if r.LargestFreeBytes > free || r.LargestFreeBytes < r.FreeRuns[len(r.FreeRuns)-1].MinBytes {
		t.Fatalf("largest free run %d does not fit the classes %+v", r.LargestFreeBytes, r.FreeRuns)
	}

	// The live values end inside the file.
	if r.ValueEndBytes == 0 || r.ValueEndBytes > u.FileBytes {
		t.Fatalf("ValueEndBytes = %d, file %d", r.ValueEndBytes, u.FileBytes)
	}

	// The handle's own slot is listed as live.
	var own bool
	for _, s := range r.Slots {
		own = own || (s.Own && s.Live && s.Slot == db.slot)
	}
	if !own {
		t.Fatalf("slots %+v lack the handle's own", r.Slots)
	}

	// On Unix the file system reports what it allocates, and only free runs
	// not punched yet hold any of it.
	if runtime.GOOS == "windows" {
		return
	}
	if r.DiskBytes <= 0 || r.DiskBytes > u.FileBytes {
		t.Fatalf("DiskBytes = %d, file %d", r.DiskBytes, u.FileBytes)
	}
	if r.FreeDiskBytes < 0 || r.FreeDiskBytes > r.UnpunchedBytes {
		t.Fatalf("FreeDiskBytes = %d, unpunched %d", r.FreeDiskBytes, r.UnpunchedBytes)
	}
}
