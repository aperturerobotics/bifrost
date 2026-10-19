//go:build !js

package web_pkg_vite

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestLinkProjectNodeModulesCleansCreatedLink checks that one-shot Vite state
// does not retain a link to a project module tree across runs.
func TestLinkProjectNodeModulesCleansCreatedLink(t *testing.T) {
	// Create persistent script state, an old project tree, and the current tree.
	root := t.TempDir()
	scriptDir := filepath.Join(root, "state", "vite")
	oldModules := filepath.Join(root, "old", "node_modules")
	currentModules := filepath.Join(root, "current", "node_modules")
	for _, dir := range []string{scriptDir, oldModules, currentModules} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Repoint a stale link left by an earlier run at the current project tree.
	link := filepath.Join(scriptDir, "node_modules")
	if err := os.Symlink(oldModules, link); err != nil {
		t.Fatal(err)
	}
	cleanup, err := linkProjectNodeModules(scriptDir, filepath.Dir(currentModules), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if got, err := os.Readlink(link); err != nil || got != currentModules {
		t.Fatalf("node_modules link = %q, %v; want %q", got, err, currentModules)
	}

	// Remove the temporary link so persistent Vite state cannot retain it.
	cleanup()
	if _, err := os.Lstat(link); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("node_modules link remains after cleanup: %v", err)
	}
}
