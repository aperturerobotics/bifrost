//go:build !js && !tinygo

package s4wave_vm_world

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/aperturerobotics/starpc/srpc"
	"github.com/s4wave/spacewave/db/unixfs"
	unixfs_world "github.com/s4wave/spacewave/db/unixfs/world"
	"github.com/s4wave/spacewave/db/world"
	db_world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	s4wave_process "github.com/s4wave/spacewave/sdk/process"
	s4wave_terminal "github.com/s4wave/spacewave/sdk/terminal"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// TestWasiVmRunsInitOnConsole boots a WasiVM whose init echoes one console
// line, types the line through a viewer console, and checks that the exit of
// init stops the VM.
func TestWasiVmRunsInitOnConsole(t *testing.T) {
	// Start a World testbed.
	ctx := t.Context()
	wtb, err := db_world_testbed.Default(ctx, db_world_testbed.WithWorldVerbose(false))
	if err != nil {
		t.Fatal(err)
	}
	defer wtb.Release()
	ws := world.NewEngineWorldState(wtb.Engine, true)

	// Create a rootfs holding the init program.
	initWasm, err := os.ReadFile("testdata/wasivm-echo.wasm")
	if err != nil {
		t.Fatal(err)
	}
	const rootfsKey, vmKey = "fs/rootfs", "vm/wasivm/test"
	writeRootfsFile(t, ws, rootfsKey, "init", initWasm)

	// Create a VM on the rootfs and start it.
	create := s4wave_vm.NewCreateWasiVmOp(vmKey, "test", rootfsKey, time.Unix(1, 0))
	create.Config = &s4wave_vm.WasiVmConfig{InitPath: "/init"}
	if _, _, err := ws.ApplyWorldOp(ctx, create, ""); err != nil {
		t.Fatal(err)
	}
	start := s4wave_vm.NewSetWasiVmStateOp(vmKey, s4wave_vm.VmState_VmState_RUNNING)
	if _, _, err := ws.ApplyWorldOp(ctx, start, ""); err != nil {
		t.Fatal(err)
	}

	// Serve the VM resource and run its execution until it is running.
	inv, release, err := wasiVmFactory(ctx, wtb.Logger, wtb.Bus, wtb.Engine, ws, vmKey)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	client := srpc.NewClient(srpc.NewServerPipe(srpc.NewServer(inv)))
	exec, err := s4wave_process.NewSRPCPersistentExecutionServiceClient(client).Execute(ctx, &s4wave_process.ExecuteRequest{})
	if err != nil {
		t.Fatal(err)
	}
	defer exec.Close()
	waitExecutionState(t, exec, s4wave_process.ExecutionState_ExecutionState_RUNNING)

	// Type a line on the console and read until init exits.
	console, err := s4wave_vm.NewSRPCWasiVmConsoleServiceClient(client).Console(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer console.Close()
	if err := console.Send(&s4wave_terminal.TerminalFrame{
		Kind: s4wave_terminal.TerminalFrameKind_TERMINAL_FRAME_KIND_INPUT,
		Data: []byte("hi\r"),
	}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	for {
		frame, err := console.Recv()
		if err != nil {
			t.Fatalf("console ended before exit; output = %q: %v", out.String(), err)
		}
		out.Write(frame.GetData())
		if frame.GetKind() == s4wave_terminal.TerminalFrameKind_TERMINAL_FRAME_KIND_EXIT {
			if frame.GetExitCode() != 0 || frame.GetError() != "" {
				t.Fatalf("exit = %d %q, want a clean exit", frame.GetExitCode(), frame.GetError())
			}
			break
		}
	}
	if !bytes.Contains(out.Bytes(), []byte("got:hi\r\n")) {
		t.Fatalf("console output = %q, want got:hi", out.String())
	}

	// The exit of init stops the VM.
	waitExecutionState(t, exec, s4wave_process.ExecutionState_ExecutionState_STOPPED)
	vm := readWasiVmObject(t, ws, vmKey)
	if vm.GetState() != s4wave_vm.VmState_VmState_STOPPED || vm.GetObservedState() != s4wave_vm.VmState_VmState_STOPPED {
		t.Fatalf("vm state = %s observed %s, want STOPPED", vm.GetState(), vm.GetObservedState())
	}
}

// writeRootfsFile creates the UnixFS object objKey holding one executable file.
func writeRootfsFile(t *testing.T, ws world.WorldState, objKey, name string, content []byte) {
	// Create the rootfs object.
	t.Helper()
	ctx := t.Context()
	op := unixfs_world.NewFsInitOp(objKey, unixfs_world.FSType_FSType_FS_NODE, nil, false, time.Unix(1, 0))
	if _, _, err := ws.ApplyWorldOp(ctx, op, ""); err != nil {
		t.Fatal(err)
	}
	obj, _, err := ws.GetObject(ctx, objKey)
	defer world.ReleaseObjectState(obj)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := unixfs_world.FsMknodWithContent(
		ctx,
		ws,
		obj,
		"",
		unixfs_world.FSType_FSType_FS_NODE,
		[]string{name},
		unixfs.NewFSCursorNodeType_File(),
		int64(len(content)),
		bytes.NewReader(content),
		0o755,
		time.Unix(2, 0),
	); err != nil {
		t.Fatal(err)
	}
}

// waitExecutionState reads execution statuses until one reports state.
func waitExecutionState(
	t *testing.T,
	exec s4wave_process.SRPCPersistentExecutionService_ExecuteClient,
	state s4wave_process.ExecutionState,
) {
	t.Helper()
	for {
		status, err := exec.Recv()
		if err != nil {
			t.Fatalf("execution ended before %s: %v", state, err)
		}
		switch status.GetState() {
		case state:
			return
		case s4wave_process.ExecutionState_ExecutionState_ERROR:
			t.Fatalf("execution failed: %s", status.GetError())
		}
	}
}

// readWasiVmObject reads the WasiVm block of objKey.
func readWasiVmObject(t *testing.T, ws world.WorldState, objKey string) *s4wave_vm.WasiVm {
	// Read the VM block of the object.
	t.Helper()
	ctx := t.Context()
	obj, _, err := ws.GetObject(ctx, objKey)
	defer world.ReleaseObjectState(obj)
	if err != nil {
		t.Fatal(err)
	}
	vm, err := readWasiVm(ctx, obj)
	if err != nil {
		t.Fatal(err)
	}
	return vm
}
