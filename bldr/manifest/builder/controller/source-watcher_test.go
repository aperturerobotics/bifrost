//go:build !js

package bldr_manifest_builder_controller

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	bldr_manifest_builder "github.com/s4wave/spacewave/bldr/manifest/builder"
)

// TestSourceWatcher checks that changes made before Wait are kept, that touch
// counts as a change, and that unwatched files and bare permission changes do
// not.
func TestSourceWatcher(t *testing.T) {
	// Bound the test so a missed event fails instead of hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Watch one input file in a temporary directory.
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.ts")
	otherPath := filepath.Join(dir, "other.ts")
	for _, p := range []string{inputPath, otherPath} {
		if err := os.WriteFile(p, []byte("a"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	inputFile := &bldr_manifest_builder.InputManifest_File{Path: "input.ts"}
	w, err := NewSourceWatcher(ctx)
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]*bldr_manifest_builder.InputManifest_File{inputPath: inputFile}
	if err := w.SetPaths(paths); err != nil {
		t.Fatal(err)
	}

	// waitChanged waits for the next change and checks it names the input.
	waitChanged := func(step string) {
		t.Helper()
		files, events, err := w.Wait(ctx, 50*time.Millisecond)
		if err != nil {
			t.Fatalf("%s: %v", step, err)
		}
		if len(files) != 1 || files[0] != inputFile || events == 0 {
			t.Fatalf("%s: expected the input file, got %v after %d events", step, files, events)
		}
	}

	// A write made while no one waits, as during a build, is kept.
	if err := os.WriteFile(inputPath, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitChanged("write before wait")

	// A write to an unwatched file is dropped, and a bare permission change
	// leaves the modification time alone, so neither wakes the waiter.
	if err := os.WriteFile(otherPath, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(inputPath, 0o600); err != nil {
		t.Fatal(err)
	}
	quietCtx, quietCancel := context.WithTimeout(ctx, 300*time.Millisecond)
	files, _, err := w.Wait(quietCtx, 50*time.Millisecond)
	quietCancel()
	if err != context.DeadlineExceeded {
		t.Fatalf("expected no change, got %v and %v", files, err)
	}

	// Touching the input moves its modification time, which counts.
	now := time.Now().Add(time.Second)
	if err := os.Chtimes(inputPath, now, now); err != nil {
		t.Fatal(err)
	}
	waitChanged("touch")
}
