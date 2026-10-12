package s4wave_vm_world

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/starpc/srpc"
	"github.com/pkg/errors"
	resource_server "github.com/s4wave/spacewave/bldr/resource/server"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	bifrost_rpc "github.com/s4wave/spacewave/net/rpc"
	s4wave_process "github.com/s4wave/spacewave/sdk/process"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	"github.com/s4wave/spacewave/sdk/world/objecttype"
	"github.com/sirupsen/logrus"
)

// WasiVmType is the ObjectType for vm/wasivm objects.
var WasiVmType = objecttype.NewObjectType(s4wave_vm.WasiVmTypeID, wasiVmFactory)

// wasiVmFactory creates a WasiVM resource with PersistentExecutionService and
// WasiVmConsoleService. The Space runs Execute on one instance, which boots
// the kernel; the console of every instance reaches that kernel over the bus.
func wasiVmFactory(
	ctx context.Context,
	le *logrus.Entry,
	b bus.Bus,
	engine world.Engine,
	ws world.WorldState,
	objectKey string,
) (srpc.Invoker, func(), error) {
	// Read and write the VM through the engine when there is one.
	if ws == nil {
		return nil, nil, objecttype.ErrWorldStateRequired
	}
	resourceWS := ws
	if engine != nil {
		resourceWS = world.NewEngineWorldState(engine, true)
	}

	// Serve execution and the console of the VM. The running instance routes
	// its console under the process prefix of the VM, which the Space forwards
	// from the bus of every viewer.
	route := objecttype.ProcessServicePrefix(objecttype.EngineIDFromContext(ctx), objectKey)
	resource := &wasiVmResource{le: le, objectKey: objectKey, ws: resourceWS, b: b, route: route}
	console := newWasiVmConsoleForwarder(b, route)
	mux := resource_server.NewResourceMux(func(mux srpc.Mux) error {
		if err := s4wave_process.SRPCRegisterPersistentExecutionService(mux, resource); err != nil {
			return err
		}
		return s4wave_vm.SRPCRegisterWasiVmConsoleService(mux, console)
	})
	return mux, func() {}, nil
}

// wasiVmResource implements PersistentExecutionService for a WasiVm object.
type wasiVmResource struct {
	le        *logrus.Entry
	objectKey string
	ws        world.WorldState
	// b serves the console route of the running VM.
	b bus.Bus
	// route prefixes the services of the running VM on b.
	route string
}

// wasiVmRun is one boot of a WasiVM.
type wasiVmRun struct {
	// generation is the run generation the boot belongs to.
	generation uint64
	// cancel stops the kernel.
	cancel context.CancelFunc
	// releaseRoute removes the console route.
	releaseRoute func()
}

// stop removes the console route and stops the kernel.
func (r *wasiVmRun) stop() {
	r.releaseRoute()
	r.cancel()
}

// Execute reconciles the desired VM state with a kernel booted in this
// process.
//
// Starting a run boots the kernel on the linked rootfs and exposes its console
// on the bus, then commits RUNNING. A run that cannot start commits ERROR.
// When init exits, the run commits STOPPED as both the desired and observed
// state; when it fails, it commits ERROR. Every commit is fenced by the run
// generation, so a late report from an earlier run changes nothing.
func (r *wasiVmResource) Execute(req *s4wave_process.ExecuteRequest, stream s4wave_process.SRPCPersistentExecutionService_ExecuteStream) error {
	// Stop the current run when its generation ends or execution ends.
	var run *wasiVmRun
	stopRun := func() {
		if run != nil {
			run.stop()
			run = nil
		}
	}
	defer stopRun()

	// Suppress repeated execution states while retaining error reports.
	lastEmitted := s4wave_process.ExecutionState(-1)
	emit := func(s s4wave_process.ExecutionState, errMsg string) error {
		if s == lastEmitted && errMsg == "" {
			return nil
		}
		if err := stream.Send(&s4wave_process.ExecuteStatus{State: s, Error: errMsg}); err != nil {
			return err
		}
		lastEmitted = s
		return nil
	}

	// Open the VM object and report a missing object as stopped.
	ctx := stream.Context()
	objState, found, err := r.ws.GetObject(ctx, r.objectKey)
	defer world.ReleaseObjectState(objState)
	if err != nil {
		return err
	}
	if !found {
		return emit(s4wave_process.ExecutionState_ExecutionState_STOPPED, "")
	}

	// Reconcile the run against each VM object revision.
	for {
		_, rev, err := objState.GetRootRef(ctx)
		if err != nil {
			return err
		}
		vm, err := readWasiVm(ctx, objState)
		if err != nil {
			return err
		}
		generation := vm.GetRunGeneration()
		observed := vm.GetObservedState()
		if run != nil && run.generation != generation {
			stopRun()
		}

		switch vm.GetState() {
		case s4wave_vm.VmState_VmState_RUNNING, s4wave_vm.VmState_VmState_STARTING:
			if observed == s4wave_vm.VmState_VmState_ERROR {
				stopRun()
				if err := emit(s4wave_process.ExecutionState_ExecutionState_ERROR, vm.GetErrorMessage()); err != nil {
					return err
				}
				break
			}
			if run == nil {
				run, err = r.startRun(ctx, generation, vm.GetConfig())
				if err != nil {
					if _, _, err := r.setObservedState(ctx, generation, s4wave_vm.VmState_VmState_ERROR, err.Error()); err != nil {
						return err
					}
					continue
				}
			}
			changed, accepted, err := r.setObservedState(ctx, generation, s4wave_vm.VmState_VmState_RUNNING, "")
			if err != nil {
				return err
			}
			if changed || !accepted {
				continue
			}
			if err := emit(s4wave_process.ExecutionState_ExecutionState_RUNNING, ""); err != nil {
				return err
			}
		default:
			stopRun()
			changed, accepted, err := r.setObservedState(ctx, generation, s4wave_vm.VmState_VmState_STOPPED, "")
			if err != nil {
				return err
			}
			if changed || !accepted {
				continue
			}
			if err := emit(s4wave_process.ExecutionState_ExecutionState_STOPPED, ""); err != nil {
				return err
			}
		}

		// Wait for the VM object revision to change before reconciling again.
		if _, err := objState.WaitRev(ctx, rev+1, false); err != nil {
			if errors.Is(err, context.Canceled) {
				return ctx.Err()
			}
			if errors.Is(err, world.ErrObjectNotFound) {
				return nil
			}
			return err
		}
	}
}

// readWasiVm reads the WasiVm block of a VM object.
func readWasiVm(ctx context.Context, objState world.ObjectState) (*s4wave_vm.WasiVm, error) {
	var vm *s4wave_vm.WasiVm
	_, _, err := world.AccessObjectState(ctx, objState, false, func(bcs *block.Cursor) error {
		var err error
		vm, err = block.UnmarshalBlock[*s4wave_vm.WasiVm](ctx, bcs, func() block.Block {
			return &s4wave_vm.WasiVm{}
		})
		return err
	})
	if err == nil && vm == nil {
		err = errors.New("wasivm block missing on object")
	}
	return vm, err
}

// startRun boots the kernel for a run generation and exposes its console.
func (r *wasiVmResource) startRun(ctx context.Context, generation uint64, config *s4wave_vm.WasiVmConfig) (*wasiVmRun, error) {
	// Find the root filesystem of the VM.
	rootfsKey, found, err := lookupSingleEdge(ctx, r.ws, r.objectKey, string(s4wave_vm.PredWasiVmRootfs))
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("wasivm has no root filesystem")
	}

	// Boot the kernel; a run that ends on its own commits how it ended.
	runCtx, cancel := context.WithCancel(ctx)
	console, err := startWasiVmKernel(runCtx, r.le, r.ws, rootfsKey, config, func(_ int32, runErr error) {
		if runCtx.Err() == nil {
			r.finishRun(runCtx, generation, runErr)
		}
	})
	if err != nil {
		cancel()
		return nil, err
	}

	// Expose the console to viewers.
	releaseRoute, err := r.exposeConsole(ctx, console)
	if err != nil {
		cancel()
		return nil, err
	}
	return &wasiVmRun{generation: generation, cancel: cancel, releaseRoute: releaseRoute}, nil
}

// finishRun commits the end of a run: STOPPED as both the desired and
// observed state when init exited, or ERROR when the run failed.
func (r *wasiVmResource) finishRun(ctx context.Context, generation uint64, runErr error) {
	// Commit the end of the run under its generation.
	_, _, err := r.updateRun(ctx, generation, func(vm *s4wave_vm.WasiVm) bool {
		// A failed run keeps the desired state for a retry.
		if runErr != nil {
			vm.ObservedState = s4wave_vm.VmState_VmState_ERROR
			vm.ErrorMessage = runErr.Error()
			return true
		}

		// An exited init stops the VM.
		vm.State = s4wave_vm.VmState_VmState_STOPPED
		vm.ObservedState = s4wave_vm.VmState_VmState_STOPPED
		vm.ErrorMessage = ""
		return true
	})
	if err != nil && ctx.Err() == nil {
		r.le.WithError(err).Warn("unable to record the end of a wasivm run")
	}
}

// setObservedState commits the observed state of a run generation.
func (r *wasiVmResource) setObservedState(
	ctx context.Context,
	generation uint64,
	state s4wave_vm.VmState,
	errorMessage string,
) (changed bool, accepted bool, err error) {
	if state != s4wave_vm.VmState_VmState_ERROR {
		errorMessage = ""
	}
	return r.updateRun(ctx, generation, func(vm *s4wave_vm.WasiVm) bool {
		if vm.GetObservedState() == state && vm.GetErrorMessage() == errorMessage {
			return false
		}
		vm.ObservedState = state
		vm.ErrorMessage = errorMessage
		return true
	})
}

// updateRun applies update to the VM block when its run generation matches
// generation. update reports whether it changed the block. accepted reports
// whether the generation matched.
func (r *wasiVmResource) updateRun(
	ctx context.Context,
	generation uint64,
	update func(vm *s4wave_vm.WasiVm) bool,
) (changed bool, accepted bool, err error) {
	// Open the VM object.
	objState, found, err := r.ws.GetObject(ctx, r.objectKey)
	defer world.ReleaseObjectState(objState)
	if err != nil {
		return false, false, err
	}
	if !found {
		return false, false, world.ErrObjectNotFound
	}

	// Update the VM block when the run generation matches.
	_, _, err = world.AccessObjectState(ctx, objState, true, func(bcs *block.Cursor) error {
		// Read the block and skip it unless the generation matches.
		vm, err := block.UnmarshalBlock[*s4wave_vm.WasiVm](ctx, bcs, func() block.Block {
			return &s4wave_vm.WasiVm{}
		})
		if err != nil || vm == nil || vm.GetRunGeneration() != generation {
			return err
		}
		accepted = true
		if changed = update(vm); changed {
			bcs.SetBlock(vm, true)
		}
		return nil
	})
	return changed, accepted, err
}

// exposeConsole routes the console service of a running VM on the bus, where
// the console forwarder of each viewer reaches it.
func (r *wasiVmResource) exposeConsole(ctx context.Context, console *wasiVmConsole) (func(), error) {
	mux := srpc.NewMux(nil)
	if err := s4wave_vm.SRPCRegisterWasiVmConsoleService(mux, console); err != nil {
		return nil, err
	}
	ctrl := bifrost_rpc.NewRpcServiceController(
		controller.NewInfo(
			"sdk/vm/world/wasivm/"+r.objectKey+"/console-route",
			controller.MustParseVersion("0.0.1"),
			"console route of a running WasiVM",
		),
		func(context.Context, func()) (srpc.Invoker, func(), error) {
			return mux, nil, nil
		},
		[]string{r.route},
		true,
		nil,
		nil,
		nil,
	)
	return r.b.AddController(ctx, ctrl, nil)
}

// _ is a type assertion
var _ s4wave_process.SRPCPersistentExecutionServiceServer = (*wasiVmResource)(nil)
