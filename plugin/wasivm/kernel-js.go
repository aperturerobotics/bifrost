//go:build js

package wasivm_plugin

import (
	"context"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine/browser"
)

// newRuntime creates the engine that runs the guests, the browser's
// WebAssembly API.
func newRuntime(context.Context) engine.Runtime {
	return browser.New()
}
