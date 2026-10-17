//go:build js && goscript

package wayland

import (
	"context"
	"testing"

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
