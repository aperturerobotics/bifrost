// Package native runs kernel guests on wazero.
package native

import (
	"context"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// Runtime is a wazero runtime with threads enabled. Canceling the context of
// a running guest stops it, even while it computes.
type Runtime struct {
	// r is the wazero runtime.
	r wazero.Runtime

	// mtx guards instances.
	mtx sync.Mutex
	// instances maps each wazero module to its Instance, so a host call can
	// find the instance that made it.
	instances map[api.Module]*Instance
}

// New creates a Runtime.
func New(ctx context.Context) *Runtime {
	cfg := wazero.NewRuntimeConfig().
		WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesThreads).
		WithCloseOnContextDone(true)
	return &Runtime{
		r:         wazero.NewRuntimeWithConfig(ctx, cfg),
		instances: make(map[api.Module]*Instance),
	}
}

// Define provides fns to guests as the imports of module.
func (r *Runtime) Define(ctx context.Context, module string, fns map[string]engine.HostFunction) error {
	builder := r.r.NewHostModuleBuilder(module)
	for name, fn := range fns {
		builder.NewFunctionBuilder().
			WithGoModuleFunction(r.hostCall(fn.Call), fn.Params, fn.Results).
			Export(name)
	}
	_, err := builder.Instantiate(ctx)
	return err
}

// Compile compiles a WebAssembly binary into a reusable module.
func (r *Runtime) Compile(ctx context.Context, wasm []byte) (engine.Module, error) {
	compiled, err := r.r.CompileModule(ctx, wasm)
	if err != nil {
		return nil, err
	}
	return &Module{compiled: compiled, runtime: r}, nil
}

// WithSnapshots returns a context whose guest calls can capture their
// execution state.
func (r *Runtime) WithSnapshots(ctx context.Context) context.Context {
	return experimental.WithSnapshotter(context.WithValue(ctx, snapshotsKey{}, struct{}{}))
}

// Snapshotter returns the snapshotter of the host call running under ctx, or
// nil when the guest call was not started under WithSnapshots.
func (r *Runtime) Snapshotter(ctx context.Context) engine.Snapshotter {
	if ctx.Value(snapshotsKey{}) == nil {
		return nil
	}
	return &snapshotter{s: experimental.GetSnapshotter(ctx)}
}

// Close releases the runtime and every module it compiled.
func (r *Runtime) Close(ctx context.Context) error {
	return r.r.Close(ctx)
}

// hostCall adapts a kernel host function to wazero. wazero turns the panic
// of a failed host function into the error of the guest call.
func (r *Runtime) hostCall(fn func(context.Context, engine.Instance, []uint64) error) api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		if err := fn(ctx, r.instance(mod), stack); err != nil {
			panic(err)
		}
	}
}

// instance returns the Instance of a wazero module.
func (r *Runtime) instance(mod api.Module) *Instance {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.instances[mod]
}

// addInstance records the Instance of a wazero module.
func (r *Runtime) addInstance(mod api.Module, inst *Instance) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.instances[mod] = inst
}

// removeInstance forgets the Instance of a wazero module.
func (r *Runtime) removeInstance(mod api.Module) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	delete(r.instances, mod)
}

// _ is a type assertion
var _ engine.Runtime = (*Runtime)(nil)
