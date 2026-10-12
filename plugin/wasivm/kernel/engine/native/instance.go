package native

import (
	"context"

	"github.com/tetratelabs/wazero/api"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// Instance is a running wazero guest.
type Instance struct {
	// mod is the wazero module instance.
	mod api.Module
	// runtime is the runtime that instantiated it.
	runtime *Runtime
}

// Memory returns the instance's linear memory, or nil when the guest has
// none.
func (i *Instance) Memory() engine.Memory {
	mem := i.mod.Memory()
	if mem == nil {
		return nil
	}
	return &Memory{mem: mem}
}

// ExportedFunction returns an exported function, or nil when the guest does
// not export it.
func (i *Instance) ExportedFunction(name string) engine.Function {
	fn := i.mod.ExportedFunction(name)
	if fn == nil {
		return nil
	}
	return &Function{fn: fn}
}

// ExportedGlobal returns an exported global, or nil when the guest does not
// export it.
func (i *Instance) ExportedGlobal(name string) engine.Global {
	g := i.mod.ExportedGlobal(name)
	if g == nil {
		return nil
	}
	return &Global{g: g}
}

// Close stops the instance and releases it.
func (i *Instance) Close(ctx context.Context) error {
	i.runtime.removeInstance(i.mod)
	return i.mod.Close(ctx)
}

// _ is a type assertion
var _ engine.Instance = (*Instance)(nil)
