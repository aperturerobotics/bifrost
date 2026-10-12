//go:build !js

package kernel

import (
	"context"
	"testing"

	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-billy/v6/osfs"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine/native"
)

// testRuntime returns a native runtime closed when the test ends.
func testRuntime(t *testing.T, ctx context.Context) engine.Runtime {
	t.Helper()
	rt := native.New(ctx)
	t.Cleanup(func() { _ = rt.Close(ctx) })
	return rt
}

// hostFS returns a filesystem over the host directory dir.
func hostFS(t *testing.T, dir string) billy.Filesystem {
	t.Helper()
	return osfs.New(dir)
}
