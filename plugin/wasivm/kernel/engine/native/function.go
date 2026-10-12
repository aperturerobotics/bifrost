package native

import (
	"context"

	"github.com/tetratelabs/wazero/api"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// Function is an exported wazero guest function.
type Function struct {
	// fn is the wazero function.
	fn api.Function
}

// Call invokes the function.
func (f *Function) Call(ctx context.Context, args ...uint64) ([]uint64, error) {
	return f.fn.Call(ctx, args...)
}

// _ is a type assertion
var _ engine.Function = (*Function)(nil)
