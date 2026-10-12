//go:build js

package kernel

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-billy/v6/memfs"
	"github.com/go-git/go-billy/v6/util"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine/browser"
)

// testRuntime returns a browser runtime closed when the test ends.
func testRuntime(t *testing.T, ctx context.Context) engine.Runtime {
	t.Helper()
	rt := browser.New()
	t.Cleanup(func() { _ = rt.Close(ctx) })
	return rt
}

// hostFS returns an in-memory copy of the host directory dir. On js, go-billy's
// osfs is an in-memory filesystem that cannot read the host.
func hostFS(t *testing.T, dir string) billy.Filesystem {
	// Copy each entry under its path relative to dir.
	t.Helper()
	mfs := memfs.New()
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		// Stop at an entry the walk could not read.
		if err != nil {
			return err
		}

		// Find the entry's place and mode in the copy.
		rel, err := filepath.Rel(dir, path)
		if err != nil || rel == "." {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}

		// Recreate it as a directory, a symlink or a file.
		switch {
		case entry.IsDir():
			return mfs.MkdirAll(rel, info.Mode().Perm())
		case entry.Type()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return mfs.Symlink(target, rel)
		default:
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return util.WriteFile(mfs, rel, data, info.Mode().Perm())
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	return mfs
}
