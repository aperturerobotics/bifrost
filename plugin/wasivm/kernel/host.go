package kernel

import (
	"context"
	"fmt"
	"slices"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// fail is -1 as a 32-bit C result.
const fail = 0xFFFFFFFF

// hostFunc is a function the kernel exports to guests.
type hostFunc struct {
	// sig is the signature in Emscripten notation: the result type, then
	// each parameter type, where v is none, i is i32 and j is i64. "iij"
	// is int32 f(int32, int64).
	sig string
	// call runs the function for the calling process and returns its
	// result, which is dropped when the signature has none.
	call func(ctx context.Context, p *Process, args []uint64) uint64
}

// returns builds a host function that does nothing and returns result.
func returns(sig string, result uint64) hostFunc {
	return hostFunc{sig, func(context.Context, *Process, []uint64) uint64 { return result }}
}

// sigTypes maps each signature letter to its value type.
var sigTypes = map[byte]engine.ValueType{'i': engine.I32, 'j': engine.I64}

// parseSig parses an Emscripten signature into result and parameter types.
func parseSig(sig string) (results, params []engine.ValueType, err error) {
	// Read each letter's type. A leading v is a function without a result.
	if sig == "" {
		return nil, nil, errors.New("empty signature")
	}
	types := make([]engine.ValueType, 0, len(sig))
	for i, c := range []byte(sig) {
		t, ok := sigTypes[c]
		switch {
		case ok:
			types = append(types, t)
		case i != 0 || c != 'v':
			return nil, nil, errors.Errorf("signature %q: unknown type %q", sig, c)
		}
	}

	// Split the result from the parameters.
	if sig[0] == 'v' {
		return nil, types, nil
	}
	return types[:1:1], types[1:], nil
}

// defineHostModules defines the kernel's host modules in its runtime: env
// holds the POSIX functions and wasi_snapshot_preview1 the WASI ones.
func (k *Kernel) defineHostModules(ctx context.Context) error {
	modules := map[string]map[string]hostFunc{
		"env":                    posixFunctions(),
		"wasi_snapshot_preview1": wasiFunctions(),
	}
	for module, fns := range modules {
		defs := make(map[string]engine.HostFunction, len(fns))
		for name, fn := range fns {
			def, err := k.bindHostFunc(module+"."+name, fn)
			if err != nil {
				return err
			}
			defs[name] = def
		}
		if err := k.runtime.Define(ctx, module, defs); err != nil {
			return errors.Wrap(err, module)
		}
	}
	return nil
}

// callLocked runs the host function fn for p holding the kernel lock, with
// p as the caller and its memory size recorded.
func (k *Kernel) callLocked(ctx context.Context, p *Process, fn hostFunc, args []uint64) uint64 {
	// Run fn as p under the kernel lock.
	k.mu.Lock()
	defer k.mu.Unlock()
	k.caller = p
	p.memSize = p.inst.Memory().Size()
	return fn.call(ctx, p, args)
}

// bindHostFunc binds fn to the process making each call. After the call it
// runs the handlers of signals caught before or during it.
func (k *Kernel) bindHostFunc(name string, fn hostFunc) (engine.HostFunction, error) {
	results, params, err := parseSig(fn.sig)
	if err != nil {
		return engine.HostFunction{}, errors.Wrap(err, name)
	}
	call := func(ctx context.Context, _ engine.Instance, stack []uint64) error {
		// Find the calling process.
		p, ok := ctx.Value(processContextKey{}).(*Process)
		if !ok {
			return errNoProcessContext
		}

		// Run the function, keeping the arguments for the trace.
		args := stack[:len(params)]
		if k.debugSyscalls {
			args = slices.Clone(args)
		}
		result := k.callLocked(ctx, p, fn, args)
		if len(results) != 0 {
			stack[0] = result
		}
		if k.debugSyscalls {
			k.traceHostCall(p, name, args, results, result)
		}
		return p.runCaughtSignals(ctx)
	}
	return engine.HostFunction{Params: params, Results: results, Call: call}, nil
}

// traceHostCall writes a host call and its result to the kernel's stderr.
func (k *Kernel) traceHostCall(p *Process, name string, args []uint64, results []engine.ValueType, result uint64) {
	line := fmt.Sprintf("[%d] %s%v", p.pid, name, args)
	switch {
	case len(results) == 0:
	case results[0] == engine.I32:
		line += fmt.Sprintf(" = %d", int32(result))
	default:
		line += fmt.Sprintf(" = %d", int64(result))
	}
	_, _ = fmt.Fprintln(k.console.stderr, line)
}
