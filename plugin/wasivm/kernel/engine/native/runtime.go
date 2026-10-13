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

// Snapshotter returns the snapshotter of the host call running under ctx, or
// nil when its guest is not transformed with Asyncify.
func (r *Runtime) Snapshotter(ctx context.Context) engine.Snapshotter {
	fr := frameOf(ctx)
	if fr == nil || fr.inst.unwinder == nil {
		return nil
	}
	return snapshotter{fr: fr}
}

// Close releases the runtime and every module it compiled.
func (r *Runtime) Close(ctx context.Context) error {
	return r.r.Close(ctx)
}

// hostCall adapts a kernel host function to wazero. wazero turns the panic
// of a failed host function into the error of the guest call.
func (r *Runtime) hostCall(fn hostFunc) api.GoModuleFunc {
	return func(ctx context.Context, mod api.Module, stack []uint64) {
		// Run the call in the guest's innermost frame.
		inst := r.instance(mod)
		var err error
		if fr := inst.frame(); fr != nil {
			err = fr.hostCall(ctx, fn, stack)
		} else {
			err = fn(ctx, inst, stack)
		}
		if err != nil {
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
