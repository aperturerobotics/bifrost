package kernel

import "testing"

// TestSignalCtrlC checks that Ctrl+C at the prompt leaves the shell running.
func TestSignalCtrlC(t *testing.T) {
	// Wait for the shell, then press Ctrl+C.
	tt := newTestTerminal(t)
	tt.typeText("echo ready-$((1+1))\n")
	tt.waitFor("ready-2")
	tt.typeText("\x03")

	// Run another command.
	tt.typeText("echo after-$((2+2))\n")
	tt.waitFor("after-4")
	tt.exit()
}

// TestSignalDisposition checks how each disposition of SIGINT handles its
// delivery: the default ends the process, a handler marks it caught and an
// ignored signal does nothing.
func TestSignalDisposition(t *testing.T) {
	tests := []struct {
		name       string
		handler    uint32
		wantCancel bool
		wantCaught uint64
	}{
		{name: "default", handler: 0, wantCancel: true},
		{name: "ignore", handler: 1},
		{name: "handler", handler: 0x1000, wantCaught: sigBit(SIGINT)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set the disposition.
			p := newSigTestProcess()
			p.sigActions[SIGINT] = sigAction{handler: tc.handler}
			cancelled := false
			p.cancel = func() { cancelled = true }

			// Deliver the signal.
			if res := p.sysKill(sigTestPID, SIGINT); res != 0 {
				t.Fatalf("kill = %#x, want 0", res)
			}
			if cancelled != tc.wantCancel || p.sigCaught != tc.wantCaught {
				t.Fatalf("cancelled %t, caught %#x; want %t, %#x", cancelled, p.sigCaught, tc.wantCancel, tc.wantCaught)
			}
		})
	}
}

// TestSigkillIgnoresDisposition checks SIGKILL ends a process that ignores
// it.
func TestSigkillIgnoresDisposition(t *testing.T) {
	// Ignore SIGKILL, then send it.
	p := newSigTestProcess()
	p.sigActions[SIGKILL] = sigAction{handler: 1}
	cancelled := false
	p.cancel = func() { cancelled = true }
	if res := p.sysKill(sigTestPID, SIGKILL); res != 0 || !cancelled {
		t.Fatalf("kill = %#x, cancelled %t; want 0, true", res, cancelled)
	}
}

// TestInitDropsUnhandledSignals checks init receives only the signals it
// handles or blocks, as on Linux.
func TestInitDropsUnhandledSignals(t *testing.T) {
	// Send init a signal it has no handler for, even with every signal
	// blocked, then one it handles.
	p, _ := newMemProcess(wasmPageSize)
	cancelled := false
	p.cancel = func() { cancelled = true }
	p.sigActions[SIGINT] = sigAction{handler: 0x1000}
	p.sigMask = ^uint64(0)
	if res := p.sysKill(initPID, SIGKILL); res != 0 || cancelled {
		t.Fatalf("kill SIGKILL = %#x, cancelled %t; want 0, false", res, cancelled)
	}
	p.sigMask = 0
	if res := p.sysKill(initPID, SIGINT); res != 0 || p.sigCaught != sigBit(SIGINT) {
		t.Fatalf("kill SIGINT = %#x, caught %#x; want 0, %#x", res, p.sigCaught, sigBit(SIGINT))
	}

	// Send init a blocked signal it has no handler for, which BusyBox init
	// takes with sigtimedwait to reboot.
	p.sigMask = sigBit(SIGTERM)
	if res := p.sysKill(initPID, SIGTERM); res != 0 || p.sigPending != sigBit(SIGTERM) {
		t.Fatalf("kill SIGTERM = %#x, pending %#x; want 0, %#x", res, p.sigPending, sigBit(SIGTERM))
	}
}

// TestSignalTrapHandler checks that Ctrl+C runs the shell's INT trap.
func TestSignalTrapHandler(t *testing.T) {
	// Set the trap.
	tt := newTestTerminal(t)
	tt.typeText("trap 'echo caught-$((3+3))' INT; echo ready-$((1+1))\n")
	tt.waitFor("ready-2")

	// Wait for the prompt, so the line editor reads ^C as a character. A
	// SIGINT that lands between commands only marks the trap pending until
	// the next command, as on Linux.
	tt.waitFor("# ")
	tt.typeText("\x03")
	tt.waitFor("caught-6")
	tt.exit()
}

// TestCtrlCDuringSleep checks that Ctrl+C ends a foreground command and
// returns to a working shell.
func TestCtrlCDuringSleep(t *testing.T) {
	// Start a long sleep in a child that prints first, so the child exists
	// when Ctrl+C reaches the terminal, then press Ctrl+C.
	tt := newTestTerminal(t)
	tt.typeText("sh -c 'echo sleeping-$((5*5)); exec sleep 60'\n")
	tt.waitFor("sleeping-25")
	tt.typeText("\x03")

	// Run another command.
	tt.typeText("echo after-$((2+2))\n")
	tt.waitFor("after-4")
	tt.exit()
}

// TestCtrlCSparesBackgroundJob checks that Ctrl+C reaches only the shell's
// foreground job, leaving a background job running.
func TestCtrlCSparesBackgroundJob(t *testing.T) {
	// Start a background job and a foreground one, then press Ctrl+C.
	tt := newTestTerminal(t)
	tt.typeText("sh -c 'exec sleep 60' & echo started-$((3*3))\n")
	tt.waitFor("started-9")
	tt.typeText("sh -c 'echo fg-$((4*4)); exec sleep 60'\n")
	tt.waitFor("fg-16")
	tt.typeText("\x03")

	// The background job is still running.
	tt.typeText("jobs; echo listed-$((6*6))\n")
	tt.waitFor("Running")
	tt.waitFor("listed-36")
	tt.typeText("kill %1\n")
	tt.exit()
}
