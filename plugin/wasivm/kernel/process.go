package kernel

import (
	"context"
	"maps"
	"slices"
	"sync"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// Process is a running guest program. Its syscall methods run on the
// program's goroutine and reach the shared kernel state through kernel.
type Process struct {
	procState
	sigState

	kernel *Kernel
	inst   engine.Instance
	// cancel ends the program. Exit and a terminating signal call it.
	cancel context.CancelFunc

	// exitCh closes when the program has exited and exitCode is final.
	exitCh   chan struct{}
	exitCode int

	// vfork is the parent state saved while a vfork child runs in this
	// process, or nil. sigMu guards it, since signal delivery reads it.
	vfork *vforkState
	// selfExecProc is the process that replaced this one through an exec
	// outside a vfork. The caller waits on it.
	selfExecProc *Process
	// jmpBufs holds the state setjmp recorded, by jmp_buf address.
	jmpBufs map[uint32]jmpBufEntry

	// sigMu guards sigState, vfork and the signal fields below.
	sigMu sync.Mutex
	// sigTerm is the terminating signal the process received. A vfork child
	// exits at its next safe point, leaving the parent running. Any other
	// process is canceled, at once or when it starts.
	sigTerm uint32
	// sigInterrupt ends the blocking call in progress so that a caught
	// signal's handler runs promptly.
	sigInterrupt context.CancelCauseFunc
	// sigWaitCh closes on every delivery to wake rt_sigtimedwait and
	// rt_sigsuspend. It is created again on demand after a wake.
	sigWaitCh chan struct{}
	// sigUnblockedDeliveries counts deliveries that took their disposition
	// rather than staying pending, so rt_sigsuspend can tell a wake that
	// ran a handler from a signal held by the mask.
	sigUnblockedDeliveries uint64
}

// procState is the state of a process that a vfork child changes and its
// parent gets back when the child execs or exits.
type procState struct {
	pid  int
	ppid int
	pgid int
	sid  int

	// The credentials: the real, effective, saved and filesystem ids, and
	// the supplementary groups. Every process runs as root by default, and
	// the kernel stores the ids without checking file permissions.
	uid    uint32
	gid    uint32
	euid   uint32
	egid   uint32
	suid   uint32
	sgid   uint32
	fsuid  uint32
	fsgid  uint32
	groups []uint32

	// fds maps each descriptor number to its open file.
	fds   map[uint32]*OpenFile
	cwd   string
	umask uint32
	// rlimits holds the resource limits by RLIMIT number.
	rlimits [rlimitNlimits]rlimit
	// nice is the scheduling priority getpriority returns.
	nice int32

	env  map[string]string
	args []string
}

// clone returns a copy of s that shares no maps or slices with it.
func (s procState) clone() procState {
	// Copy the maps and slices; the struct copy holds the rest.
	s.groups = slices.Clone(s.groups)
	s.fds = maps.Clone(s.fds)
	s.env = maps.Clone(s.env)
	s.args = slices.Clone(s.args)
	return s
}

// jmpBufEntry is the execution state setjmp records for longjmp: the engine
// snapshot and the C stack, which the engine snapshot does not hold.
type jmpBufEntry struct {
	snapshot engine.Snapshot
	// sp is the C stack pointer at the setjmp.
	sp uint32
	// cstack holds the C stack from sp to the heap base.
	cstack []byte
}

// newProcess creates a process with the kernel's environment, its standard
// descriptors on the console, and the root directory preopened at fd 3.
func newProcess(k *Kernel, pid int, args []string) *Process {
	// Start from the defaults.
	p := &Process{
		procState: procState{
			pid:     pid,
			pgid:    pid,
			sid:     pid,
			fds:     make(map[uint32]*OpenFile),
			cwd:     "/",
			umask:   0o022,
			rlimits: defaultRlimits(),
			env:     maps.Clone(k.env),
			args:    args,
		},
		kernel: k,
		exitCh: make(chan struct{}),
	}

	// Open the standard descriptors and the root preopen.
	if k.console != nil {
		p.fds[0] = &OpenFile{node: k.console, path: "/dev/console", flags: oRDONLY}
		p.fds[1] = &OpenFile{node: k.console, path: "/dev/console", flags: oWRONLY}
		p.fds[2] = &OpenFile{node: k.console, path: "/dev/console", flags: oWRONLY, stderr: true}
	}
	if k.vfs != nil {
		p.fds[3] = &OpenFile{node: k.vfs.root, path: "/", flags: oDIRECTORY}
	}
	return p
}

// allocFD returns the lowest unused descriptor number, as POSIX requires of
// open, dup and pipe.
func (p *Process) allocFD() uint32 {
	fd := uint32(0)
	for p.fds[fd] != nil {
		fd++
	}
	return fd
}

// start runs the guest's _start entry point until the program exits.
func (p *Process) start(ctx context.Context) error {
	start := p.inst.ExportedFunction("_start")
	if start == nil {
		return ErrNoStart
	}
	_, err := start.Call(ctx)
	return err
}

// readCString reads a null-terminated string from WASM memory.
func readCString(mem engine.Memory, ptr uint32) string {
	var buf []byte
	for i := range uint32(4096) {
		b, ok := mem.ReadByteAt(ptr + i)
		if !ok || b == 0 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf)
}
