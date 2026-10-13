package native

import (
	"context"

	"github.com/tetratelabs/wazero"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// Module is a compiled wazero module.
type Module struct {
	// compiled is the wazero module.
	compiled wazero.CompiledModule
	// runtime is the runtime that compiled it.
	runtime *Runtime
}

// Instantiate creates an instance of the module named name. The kernel calls
// the guest's entry itself, so instantiation starts nothing.
func (m *Module) Instantiate(ctx context.Context, name string) (engine.Instance, error) {
	// Instantiate the guest without running its start functions.
	cfg := wazero.NewModuleConfig().WithName(name).WithStartFunctions()
	mod, err := m.runtime.r.InstantiateModule(ctx, m.compiled, cfg)
	if err != nil {
		return nil, err
	}

	// Reserve the unwind region and record the instance for its host calls.
	unwinder, err := attachUnwinder(mod)
	if err != nil {
		_ = mod.Close(ctx)
		return nil, err
	}
	inst := &Instance{mod: mod, runtime: m.runtime, unwinder: unwinder}
	m.runtime.addInstance(mod, inst)
	return inst, nil
}

// Close releases the compiled module.
func (m *Module) Close(ctx context.Context) error {
	return m.compiled.Close(ctx)
}

// _ is a type assertion
var _ engine.Module = (*Module)(nil)
