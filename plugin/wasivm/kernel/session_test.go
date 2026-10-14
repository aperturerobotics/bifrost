package kernel

import "testing"

// newSessionTestKernel returns a kernel whose init, PID 1, leads the
// session, with children 2 and 3 in init's process group.
func newSessionTestKernel() (*Kernel, []*Process) {
	// Create a kernel with a terminal console and its processes.
	k := &Kernel{procs: newProcessTable()}
	k.console = newTTY(k, true, true)
	procs := make([]*Process, 3)
	for i := range procs {
		p := newProcess(k, i+1, nil)
		p.cancel = func() {}
		if i != 0 {
			p.ppid, p.pgid, p.sid = initPID, initPID, initPID
		}
		k.procs.set(p.pid, p)
		procs[i] = p
	}
	return k, procs
}

// TestSetpgid checks which process group moves setpgid allows.
func TestSetpgid(t *testing.T) {
	_, procs := newSessionTestKernel()
	init, child, other := procs[0], procs[1], procs[2]
	tests := []struct {
		name       string
		caller     *Process
		pid, pgid  int32
		want       uint64
		wantTarget *Process
		wantPgid   int
	}{
		{name: "child starts a group", caller: child, want: 0, wantTarget: child, wantPgid: 2},
		{name: "parent moves a child", caller: init, pid: 3, pgid: 2, want: 0, wantTarget: other, wantPgid: 2},
		{name: "session leader stays", caller: init, want: negErrno(EPERM), wantTarget: init, wantPgid: 1},
		{name: "not a child", caller: child, pid: 3, pgid: 2, want: negErrno(ESRCH), wantTarget: other, wantPgid: 2},
		{name: "missing group", caller: init, pid: 3, pgid: 9, want: negErrno(EPERM), wantTarget: other, wantPgid: 2},
	}
	for _, tc := range tests {
		if got := tc.caller.sysSetpgid(tc.pid, tc.pgid); got != tc.want {
			t.Fatalf("%s: setpgid = %#x, want %#x", tc.name, got, tc.want)
		}
		if tc.wantTarget.pgid != tc.wantPgid {
			t.Fatalf("%s: pgid %d, want %d", tc.name, tc.wantTarget.pgid, tc.wantPgid)
		}
	}
}

// TestForegroundGroupSignals checks that the terminal's signals reach only
// the foreground process group once one is set, and every process before.
func TestForegroundGroupSignals(t *testing.T) {
	// Move child 2 into its own group.
	k, procs := newSessionTestKernel()
	if res := procs[1].sysSetpgid(0, 0); res != 0 {
		t.Fatalf("setpgid = %#x", res)
	}

	// Signal before and after making it the foreground group.
	for _, fg := range []int{0, 2} {
		k.console.fgPgid = fg
		for _, p := range procs {
			p.sigActions[SIGINT] = sigAction{handler: 0x1000}
			p.sigCaught = 0
		}
		k.console.signalForeground(SIGINT)
		for _, p := range procs {
			want := fg == 0 || p.pgid == fg
			if got := p.sigCaught != 0; got != want {
				t.Fatalf("foreground %d: pid %d caught %t, want %t", fg, p.pid, got, want)
			}
		}
	}
}

// TestKillAll checks that kill(-1) skips init and the caller.
func TestKillAll(t *testing.T) {
	_, procs := newSessionTestKernel()
	for _, p := range procs {
		p.sigActions[SIGTERM] = sigAction{handler: 0x1000}
	}
	if res := procs[1].sysKill(-1, SIGTERM); res != 0 {
		t.Fatalf("kill = %#x", res)
	}
	for i, want := range []bool{false, false, true} {
		if got := procs[i].sigCaught != 0; got != want {
			t.Fatalf("pid %d caught %t, want %t", procs[i].pid, got, want)
		}
	}
}

// TestExitReparentsToInit checks that init adopts the children of an
// exiting process and hears SIGCHLD for one that already exited.
func TestExitReparentsToInit(t *testing.T) {
	// Make 3 a child of 2, exited and not reaped.
	k, procs := newSessionTestKernel()
	init, parent, orphan := procs[0], procs[1], procs[2]
	orphan.ppid = parent.pid
	close(orphan.exitCh)
	init.sigActions[SIGCHLD] = sigAction{handler: 0x1000}

	// Exit the parent.
	k.exitProcess(parent)
	if orphan.ppid != initPID {
		t.Fatalf("orphan ppid %d, want %d", orphan.ppid, initPID)
	}
	if init.sigCaught&sigBit(SIGCHLD) == 0 {
		t.Fatal("init did not receive SIGCHLD")
	}
}
