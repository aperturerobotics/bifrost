//go:build !js

package bldr_util_buildcache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestCacheDo walks one key through a build, a replay into an emptied output
// root, an edit, a revert of the edit, and the eviction of the oldest state.
func TestCacheDo(t *testing.T) {
	// Build into a temporary cache, source, and output root.
	le := logrus.NewEntry(logrus.New())
	cache := Open(t.TempDir())
	src := filepath.Join(t.TempDir(), "main.ts")
	outputRoot := filepath.Join(t.TempDir(), "dist")
	key := []byte("bundle request")

	// The step copies its input into two outputs and counts its runs.
	var runs int
	do := func(content string) string {
		// Write the input and run the step through the cache.
		t.Helper()
		if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		value, err := cache.Do(le, key, outputRoot, func() (*Step, error) {
			runs++
			for _, out := range []string{"main.mjs", "chunks/dep.mjs"} {
				path := filepath.Join(outputRoot, filepath.FromSlash(out))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					return nil, err
				}
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					return nil, err
				}
			}
			return &Step{
				Inputs:  []string{src},
				Outputs: []string{"main.mjs", "chunks/dep.mjs"},
				Value:   []byte("built " + content),
			}, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		// Require the output root to hold this input's outputs.
		data, err := os.ReadFile(filepath.Join(outputRoot, "chunks", "dep.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != content {
			t.Fatalf("output holds %q, want %q", data, content)
		}
		return string(value)
	}

	// Build, replay into an emptied root, edit, revert, and evict.
	for i, step := range []struct {
		content string
		wipe    bool
		runs    int
	}{
		{content: "a", runs: 1},
		{content: "a", wipe: true, runs: 1},
		{content: "b", runs: 2},
		{content: "a", runs: 2},
		{content: "c", runs: 3},
		{content: "b", runs: 4},
	} {
		if step.wipe {
			if err := os.RemoveAll(outputRoot); err != nil {
				t.Fatal(err)
			}
		}
		if value := do(step.content); value != "built "+step.content {
			t.Fatalf("step %d: value %q", i, value)
		}
		if runs != step.runs {
			t.Fatalf("step %d: %d runs, want %d", i, runs, step.runs)
		}
	}
}
