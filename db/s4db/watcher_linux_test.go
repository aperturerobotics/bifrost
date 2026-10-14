package s4db

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// TestWatchersShareInotify checks that open databases share one inotify
// instance, since the kernel limits instances per user across the host.
func TestWatchersShareInotify(t *testing.T) {
	// Open more databases than one.
	dir := t.TempDir()
	for i := range 4 {
		db := openTest(t, filepath.Join(dir, strconv.Itoa(i)+".s4wave"), Options{})
		defer db.Close()
	}

	// Count the process's inotify descriptors.
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	var instances int
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err == nil && target == "anon_inode:inotify" {
			instances++
		}
	}
	if instances != 1 {
		t.Fatalf("%d inotify instances, want 1", instances)
	}
}
