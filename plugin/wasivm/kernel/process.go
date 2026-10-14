package kernel

import (
	"context"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// Process is a running guest program. Its syscall methods run on the
// program's goroutine and reach the shared kernel state through kernel.
type Process struct {
	procState
	sigState

	kernel *Kernel
	inst   engine.Instance
	// module is the compiled program inst runs, and globals are the export
	// names of its mutable globals. A fork copies them into a new instance.
	module  engine.Module
	globals []string
	// cancel ends the program. Exit and a terminating signal call it.
	cancel context.CancelFunc
	// started is when the process started, kept across exec.
	started time.Time
	// memSize is the size of the program's memory at its last host call. The
	// kernel lock guards it.
	memSize uint32

	// exitCh closes when the program has exited and exitCode and exitSignal
	// are final. exitSignal is the signal that ended the program, or 0.
	exitCh     chan struct{}
	exitCode   int
	exitSignal uint32

	// selfExecProc is the process that replaced this one through exec. The
	// caller waits on it. It is set under sigMu, which hands later signals
	// to it.
	selfExecProc *Process
	// jmpBufs holds the state setjmp recorded, by jmp_buf address.
	jmpBufs map[uint32]jmpBufEntry
	// handlerDepth counts the signal handlers running.
	handlerDepth int

	// sigMu guards sigState and the signal fields below.
	sigMu sync.Mutex
	// sigTerm is the terminating signal the process received. The process
	// is canceled, at once or when it starts.
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

// procState is the state a fork child copies from its parent.
type procState struct {
	pid  int
	ppid int
	pgid int
	sid  int

	// The credentials: the user and group ids and the supplementary groups.
	// Every process runs as root by default, and the kernel stores the ids
	// without checking file permissions.
	uids   ids
	gids   ids
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
	// exe is the path of the program the process runs.
	exe string
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
		kernel:  k,
		exitCh:  make(chan struct{}),
		started: time.Now(),
	}

	// Open the standard descriptors and the root preopen.
	if k.console != nil {
		p.setFD(0, &OpenFile{node: k.console, path: "/dev/console", flags: oRDONLY})
		p.setFD(1, &OpenFile{node: k.console, path: "/dev/console", flags: oWRONLY})
		p.setFD(2, &OpenFile{node: k.console, path: "/dev/console", flags: oWRONLY, stderr: true})
	}
	if k.vfs != nil {
		p.setFD(3, &OpenFile{node: k.vfs.root, path: "/", flags: oDIRECTORY})
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

// exited reports whether the program has exited.
func (p *Process) exited() bool {
	select {
	case <-p.exitCh:
		return true
	default:
		return false
	}
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
