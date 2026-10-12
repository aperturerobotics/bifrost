//go:build !js

package wasivm_plugin

import (
	"context"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine/native"
)

// newRuntime creates the engine that runs the guests, wazero natively.
func newRuntime(ctx context.Context) engine.Runtime {
	return native.New(ctx)
}
