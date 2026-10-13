package kernel

import (
	"context"
	"slices"
	"testing"
)

// testHandlerWasm is this module, built with wat2wasm --debug-names:
//
//	(module
//	  (type (func (param i32)))
//	  (table 1 funcref)
//	  (global $__stack_pointer (mut i32) (i32.const 65536))
//	  (global (export "caught") (mut i32) (i32.const 0))
//	  (func $handler (param i32) (global.set 1 (local.get 0)))
//	  (elem (i32.const 0) $handler))
var testHandlerWasm = []byte{
	0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x05, 0x01, 0x60,
	0x01, 0x7f, 0x00, 0x03, 0x02, 0x01, 0x00, 0x04, 0x04, 0x01, 0x70, 0x00,
	0x01, 0x06, 0x0d, 0x02, 0x7f, 0x01, 0x41, 0x80, 0x80, 0x04, 0x0b, 0x7f,
	0x01, 0x41, 0x00, 0x0b, 0x07, 0x0a, 0x01, 0x06, 0x63, 0x61, 0x75, 0x67,
	0x68, 0x74, 0x03, 0x01, 0x09, 0x07, 0x01, 0x00, 0x41, 0x00, 0x0b, 0x01,
	0x00, 0x0a, 0x08, 0x01, 0x06, 0x00, 0x20, 0x00, 0x24, 0x01, 0x0b, 0x00,
	0x2a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x01, 0x0a, 0x01, 0x00, 0x07, 0x68,
	0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x02, 0x03, 0x01, 0x00, 0x00, 0x07,
	0x12, 0x01, 0x00, 0x0f, 0x5f, 0x5f, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x5f,
	0x70, 0x6f, 0x69, 0x6e, 0x74, 0x65, 0x72,
}

// TestPrepareGuest checks that a prepared guest exports its stack pointer and
// mutable globals and runs a handler from its table through the signal
// trampoline.
func TestPrepareGuest(t *testing.T) {
	// Prepare the guest.
	ctx := context.Background()
	wasm, globals, err := prepareGuest(testHandlerWasm)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(globals, []string{guestStackPointer, "caught"}) {
		t.Fatalf("mutable globals %q, want the stack pointer and caught", globals)
	}

	// Instantiate it.
	mod, err := testRuntime(t, ctx).Compile(ctx, wasm)
	if err != nil {
		t.Fatal(err)
	}
	inst, err := mod.Instantiate(ctx, "guest")
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close(ctx)

	// The stack pointer is exported.
	sp := inst.ExportedGlobal(guestStackPointer)
	if sp == nil || sp.Get() != 65536 {
		t.Fatalf("stack pointer export: %v", sp)
	}

	// The trampoline calls handler 0 with the signal number.
	if _, err := inst.ExportedFunction(guestSignalTrampoline).Call(ctx, 0, 7); err != nil {
		t.Fatal(err)
	}
	if got := inst.ExportedGlobal("caught").Get(); got != 7 {
		t.Fatalf("handler got signal %d, want 7", got)
	}
}

// TestPrepareGuestIdempotent checks that preparing a prepared guest changes
// nothing.
func TestPrepareGuestIdempotent(t *testing.T) {
	// Prepare the guest twice.
	once, _, err := prepareGuest(testHandlerWasm)
	if err != nil {
		t.Fatal(err)
	}
	twice, _, err := prepareGuest(once)
	if err != nil {
		t.Fatal(err)
	}

	// Compare the results.
	if string(once) != string(twice) {
		t.Fatal("second prepare changed the guest")
	}
}
