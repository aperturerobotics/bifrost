package kernel

import (
	"context"

	"github.com/aperturerobotics/util/broadcast"
	"github.com/pkg/errors"
)

// processTable holds the kernel's processes by PID and counts the child
// processes still running.
type processTable struct {
	// bcast guards the fields below and wakes waiters when running drops.
	bcast broadcast.Broadcast
	// nextPID is the PID the next process receives.
	nextPID int
	// procs maps each PID to its process, including exited children not yet
	// reaped by wait4.
	procs map[int]*Process
	// running counts the child process goroutines that have not returned.
	running int
}

// newProcessTable creates an empty table whose first PID is 1.
func newProcessTable() *processTable {
	return &processTable{nextPID: 1, procs: make(map[int]*Process)}
}

// allocPID reserves the next PID.
func (t *processTable) allocPID() int {
	// Take the next PID.
	l := t.bcast.Lock()
	defer l.Unlock()
	pid := t.nextPID
	t.nextPID++
	return pid
}

// lastPID returns the PID allocated last, or 0 before the first.
func (t *processTable) lastPID() int {
	l := t.bcast.Lock()
	defer l.Unlock()
	return t.nextPID - 1
}

// set records p under pid, replacing any process there.
func (t *processTable) set(pid int, p *Process) {
	l := t.bcast.Lock()
	defer l.Unlock()
	t.procs[pid] = p
}

// get returns the process with pid, or nil.
func (t *processTable) get(pid int) *Process {
	l := t.bcast.Lock()
	defer l.Unlock()
	return t.procs[pid]
}

// remove forgets the process with pid.
func (t *processTable) remove(pid int) {
	l := t.bcast.Lock()
	defer l.Unlock()
	delete(t.procs, pid)
}

// removeProcess forgets p when its PID still maps to it. An exec replaces
// the process under the same PID, and the replacement stays.
func (t *processTable) removeProcess(p *Process) {
	l := t.bcast.Lock()
	defer l.Unlock()
	if t.procs[p.pid] == p {
		delete(t.procs, p.pid)
	}
}

// filter returns the processes match accepts, or every process when match is
// nil.
func (t *processTable) filter(match func(p *Process) bool) []*Process {
	// Collect the matching processes.
	l := t.bcast.Lock()
	defer l.Unlock()
	var out []*Process
	for _, p := range t.procs {
		if match == nil || match(p) {
			out = append(out, p)
		}
	}
	return out
}

// errNoChild is the error of waiting for a child when none matches.
var errNoChild = errors.New("no matching child process")

// waitChild waits for a child of the process ppid that match accepts to exit,
// and returns it. With nohang it returns nil at once while every matching
// child runs. It returns errNoChild when no child matches.
func (t *processTable) waitChild(ctx context.Context, ppid int, match func(p *Process) bool, nohang bool) (*Process, error) {
	for {
		// Find an exited child.
		l := t.bcast.Lock()
		found := false
		for _, p := range t.procs {
			if p.ppid != ppid || !match(p) {
				continue
			}
			found = true
			if p.exited() {
				l.Unlock()
				return p, nil
			}
		}
		waitCh := l.WaitCh()
		l.Unlock()
		switch {
		case !found:
			return nil, errNoChild
		case nohang:
			return nil, nil
		}

		// Wait for an exit.
		if !wait(ctx, waitCh) {
			return nil, context.Cause(ctx)
		}
	}
}

// exited wakes waitChild after a process exits.
func (t *processTable) exited() {
	l := t.bcast.Lock()
	l.Broadcast()
	l.Unlock()
}

// start runs fn in a goroutine counted as a running child process.
func (t *processTable) start(fn func()) {
	l := t.bcast.Lock()
	t.running++
	l.Unlock()

	go func() {
		defer t.childDone()
		fn()
	}()
}

// childDone marks one child process goroutine as returned.
func (t *processTable) childDone() {
	l := t.bcast.Lock()
	defer l.Unlock()
	t.running--
	if t.running == 0 {
		l.Broadcast()
	}
}

// waitIdle waits until no child process goroutine is running.
func (t *processTable) waitIdle(ctx context.Context) error {
	for {
		l := t.bcast.Lock()
		if t.running == 0 {
			l.Unlock()
			return nil
		}
		waitCh := l.WaitCh()
		l.Unlock()

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-waitCh:
		}
	}
}
