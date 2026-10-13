// Package kernel runs Linux programs compiled to WebAssembly. Its host
// functions implement the Linux syscalls and WASI imports those programs
// call: processes, signals, pipes, a terminal and a filesystem over an
// optional root. No Linux kernel runs; the Go host is the kernel.
package kernel

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

type processContextKey struct{}

// Kernel is one running system: its WebAssembly runtime, filesystem,
// process table and console.
type Kernel struct {
	runtime engine.Runtime
	vfs     *VFS

	// mu is the kernel lock. Each host call holds it except while it
	// blocks, so processes change the kernel's state one at a time.
	mu sync.Mutex

	// procs is the process table.
	procs *processTable
	// tempSeq numbers the directories mkdtemp creates.
	tempSeq atomic.Uint32

	// tty reports whether the console is a terminal, host or virtual.
	tty bool
	// termFD is the host terminal's descriptor, or -1 for none.
	termFD int
	// virtTTY reports whether the kernel emulates the terminal's line
	// discipline.
	virtTTY bool
	// termRaw reports whether the host terminal is in raw mode.
	termRaw atomic.Bool
	// termMtx guards termIOS, which the guest sets while the input
	// goroutine reads it.
	termMtx sync.Mutex
	// termIOS is the terminal's termios state.
	termIOS termios

	// winCols and winRows are the virtual terminal's window size.
	winCols atomic.Uint32
	winRows atomic.Uint32

	// termMakeRaw, termRestore and termGetSize operate the host terminal.
	// The platform's initTermOps sets them.
	termMakeRaw func(fd int) (any, error)
	termRestore func(fd int, state any) error
	termGetSize func(fd int) (int, int, error)
	// termOldState is the host terminal's state before raw mode, which Close
	// restores.
	termOldState any

	// console is the /dev/console node each process starts with as its
	// standard input, output and error.
	console *VNode
	stdout  io.Writer
	stderr  io.Writer
	stdin   *stdinReader

	// debugSyscalls logs every syscall to stderr.
	debugSyscalls bool

	// bootTime is when the kernel started. The monotonic clocks count from
	// it, unaffected by setting the realtime clock.
	bootTime time.Time
	// timeMu guards realtimeOffset.
	timeMu sync.Mutex
	// realtimeOffset is added to the host time to give the kernel's realtime
	// clock.
	realtimeOffset time.Duration

	// initPath and initArgs are the program Boot runs.
	initPath string
	initArgs []string
	// env is the processes' environment.
	env map[string]string

	// hostname and domainname are the names uname reports.
	hostname   string
	domainname string
}

// New creates a kernel that runs its processes on rt. The kernel defines its
// host modules in rt, so rt serves this kernel alone.
func New(ctx context.Context, rt engine.Runtime, cfg *Config) (*Kernel, error) {
	// Choose the console's terminal. A virtual terminal replaces the host's.
	if cfg == nil {
		return nil, ErrNilConfig
	}
	termFD := cfg.TerminalFD
	if (termFD == 0 && cfg.Stdin == nil) || cfg.VirtualTerminal {
		termFD = -1
	}

	// Fill in the configuration defaults.
	env := cfg.Env
	if env == nil {
		env = map[string]string{
			"PATH":  "/bin:/usr/bin:/sbin:/usr/sbin",
			"HOME":  "/root",
			"TERM":  "xterm",
			"SHELL": "/bin/sh",
		}
	}
	k := &Kernel{
		runtime:       rt,
		procs:         newProcessTable(),
		debugSyscalls: cfg.DebugSyscalls,
		tty:           termFD >= 0 || cfg.VirtualTerminal,
		termFD:        termFD,
		virtTTY:       cfg.VirtualTerminal,
		bootTime:      time.Now(),
		initPath:      cmp.Or(cfg.InitPath, "/init"),
		initArgs:      cfg.InitArgs,
		env:           env,
		hostname:      "spacewave",
		domainname:    "(none)",
	}

	// Route every output path through ONLCR processing, which turns a
	// newline into a carriage return and newline while the terminal asks
	// for it.
	k.stdout = &onlcrWriter{kernel: k, inner: cmp.Or[io.Writer](cfg.Stdout, os.Stdout)}
	k.stderr = &onlcrWriter{kernel: k, inner: cmp.Or[io.Writer](cfg.Stderr, os.Stderr)}

	// Set up the terminal, the filesystem and the host modules.
	k.winCols.Store(80)
	k.winRows.Store(24)
	k.stdin = newStdinReader(k)
	initTermOps(k)
	k.initTermios()
	k.initVFS(cfg)
	if err := k.defineHostModules(ctx); err != nil {
		return nil, err
	}

	// Start reading the console input.
	k.stdin.start(ctx, cmp.Or[io.Reader](cfg.Stdin, os.Stdin))
	return k, nil
}

// Boot runs the program at InitPath as PID 1 and waits for it.
func (k *Kernel) Boot(ctx context.Context) error {
	// Read the init program.
	node := k.vfs.lookup(k.initPath)
	if node == nil {
		return ErrNoInit
	}
	wasm, err := k.vfs.readFile(node)
	if err != nil || len(wasm) == 0 {
		return ErrNoInit
	}

	// Run it with its arguments.
	args := append([]string{k.initPath}, k.initArgs...)
	return k.RunProcess(ctx, wasm, args...)
}

// RunProcess runs a guest program as a new process and waits for it and every
// child it started. A nonzero exit returns a ProcessExitError.
func (k *Kernel) RunProcess(ctx context.Context, wasm []byte, args ...string) error {
	// Register the process under a fresh PID.
	pid := k.procs.allocPID()
	p := newProcess(k, pid, args)
	k.procs.set(pid, p)

	// Run it, then wait for the children it left running.
	err := k.runProcess(ctx, p, func(ctx context.Context) error {
		return k.runProgram(ctx, p, wasm)
	})
	if err != nil {
		return err
	}
	if err := k.procs.waitIdle(ctx); err != nil {
		return err
	}
	if p.exitCode != 0 {
		return &ProcessExitError{Code: p.exitCode}
	}
	return nil
}

// startChildProcess runs a registered child process in the background. run
// runs its guest, as runProcess describes.
func (k *Kernel) startChildProcess(ctx context.Context, p *Process, run func(ctx context.Context) error) {
	k.procs.start(func() {
		_ = k.runProcess(ctx, p, run)
	})
}

// runProcess runs a registered process to completion with run, records its
// exit and releases it. run runs the guest until it returns or exits. A
// program that cannot run exits with 126. A child reports that error on the
// kernel's stderr, since its parent sees only the exit code.
func (k *Kernel) runProcess(ctx context.Context, p *Process, run func(ctx context.Context) error) error {
	// Give the process a context its kill and exit cancel. A terminating
	// signal that arrived before the start cancels it at once.
	procCtx, cancel := context.WithCancel(context.WithValue(ctx, processContextKey{}, p))
	p.sigMu.Lock()
	p.cancel = cancel
	killed := p.sigTerm != 0
	p.sigMu.Unlock()
	if killed {
		cancel()
	}
	defer k.exitProcess(p)

	// Run the program. A terminating signal ends it with status 128 plus the
	// signal number. Report any other failure before the parent can see it.
	err := k.runGuest(procCtx, p, run)
	p.sigMu.Lock()
	sig := p.sigTerm
	p.sigMu.Unlock()
	switch {
	case sig != 0:
		p.exitCode = 128 + int(sig)
		p.exitSignal = sig
		return nil
	case err != nil:
		p.exitCode = exitCannotRun
		if p.ppid != 0 {
			_, _ = fmt.Fprintf(k.stderr, "pid %d: %v\n", p.pid, err)
		}
	}
	return err
}

// runGuest runs the process's guest with run and records its exit code. When
// the program replaced itself with exec, the exit code is the replacement's.
// It returns an error when the program could not run or trapped.
func (k *Kernel) runGuest(ctx context.Context, p *Process, run func(ctx context.Context) error) error {
	// Run the guest; it ends by panicking a ProcessExitError from exit.
	err := run(ctx)
	if p.inst != nil {
		_ = p.inst.Close(ctx)
	}
	var exitErr *ProcessExitError
	if !errors.As(err, &exitErr) {
		return err
	}
	p.exitCode = exitErr.Code

	// After exec, the replacement process carries on under this PID.
	if child := p.selfExecProc; child != nil {
		<-child.exitCh
		p.exitCode = child.exitCode
		p.exitSignal = child.exitSignal
	}
	return nil
}

// runProgram compiles wasm as the process's program and runs it from its
// entry point.
func (k *Kernel) runProgram(ctx context.Context, p *Process, wasm []byte) error {
	// Compile and instantiate the program.
	wasm, globals, err := prepareGuest(wasm)
	if err != nil {
		return err
	}
	compiled, err := k.runtime.Compile(ctx, wasm)
	if err != nil {
		return errors.Wrap(err, "compile")
	}
	inst, err := compiled.Instantiate(ctx, "proc-"+strconv.Itoa(p.pid))
	if err != nil {
		return errors.Wrap(err, "instantiate")
	}
	p.inst, p.module, p.globals = inst, compiled, globals
	return p.start(ctx)
}

// wait waits for ch to close and reports true, or for ctx to end and reports
// false, without the kernel lock.
func wait(ctx context.Context, ch <-chan struct{}) bool {
	done := false
	blocking(ctx, func() {
		select {
		case <-ctx.Done():
		case <-ch:
			done = true
		}
	})
	return done
}

// blocking runs wait, which blocks, without the kernel lock the host call
// running under ctx holds, so other processes run in the meantime. Outside a
// process's host call it runs wait as is.
func blocking(ctx context.Context, wait func()) {
	// Find the process of the host call.
	p, ok := ctx.Value(processContextKey{}).(*Process)
	if !ok {
		wait()
		return
	}

	// Release the lock while waiting.
	p.kernel.mu.Unlock()
	defer p.kernel.mu.Lock()
	wait()
}

// exitCannotRun is the exit code of a program that cannot run, as a shell
// reports a command it cannot execute.
const exitCannotRun = 126

// exitProcess releases a process after its program returns: it closes the
// process's descriptors, wakes its waiters and sends its parent SIGCHLD. When
// no parent will reap it, it removes it from the process table.
func (k *Kernel) exitProcess(p *Process) {
	// Close the descriptors and record the exit.
	for fd, of := range p.fds {
		of.unref()
		delete(p.fds, fd)
	}
	p.cancel()
	close(p.exitCh)
	k.procs.exited()

	// Tell the parent, unless an exec replaced the process and its
	// replacement tells it later.
	replaced := k.procs.get(p.pid) != p
	if p.ppid == 0 {
		k.procs.removeProcess(p)
	} else if parent := k.procs.get(p.ppid); parent != nil && !replaced {
		parent.deliverSignal(SIGCHLD)
	}
}

// Close stops reading the console, restores the host terminal and closes the
// runtime.
func (k *Kernel) Close(ctx context.Context) error {
	// Stop the input and restore the terminal.
	k.stdin.close()
	if k.termOldState != nil && k.termRestore != nil {
		_ = k.termRestore(k.termFD, k.termOldState)
		k.termOldState = nil
		k.termRaw.Store(false)
	}

	// Release the runtime.
	return k.runtime.Close(ctx)
}

// Resize sets the virtual terminal's window size and sends SIGWINCH to the
// foreground processes.
func (k *Kernel) Resize(cols, rows uint32) {
	k.winCols.Store(cols)
	k.winRows.Store(rows)
	k.deliverSignalToForeground(SIGWINCH)
}

// Signal sends sig to the foreground processes, as the host forwards a
// signal it received, such as SIGINT from a terminal in cooked mode.
func (k *Kernel) Signal(sig uint32) { k.deliverSignalToForeground(sig) }

// deliverSignalToForeground sends sig to every process, as a terminal sends
// the signal of a typed character whether or not anything reads it. The
// terminal has no foreground process group, so every process is in the
// foreground, and each one's disposition decides what the signal does.
func (k *Kernel) deliverSignalToForeground(sig uint32) {
	for _, t := range k.procs.filter(nil) {
		t.deliverSignal(sig)
	}
}

// initVFS creates the filesystem: the devices, the in-memory directories, the
// root filesystem and the configured files.
func (k *Kernel) initVFS(cfg *Config) {
	// Create a VFS over the root filesystem.
	k.vfs = newVFS()
	v := k.vfs
	v.host = cfg.RootFS

	// Add the devices.
	dev := v.mkDir(v.root, "dev")
	v.mkCharDev(dev, "null", devNull)
	v.mkCharDev(dev, "zero", devZero)
	k.console = v.mkCharDev(dev, "console", devConsole)
	v.mkCharDev(dev, "tty", devTTY)
	v.mkCharDev(dev, "urandom", devURandom)

	// Add the in-memory directories that overlay the root filesystem.
	v.mkDir(v.root, "tmp")
	proc := v.mkDir(v.root, "proc")
	v.mkDir(proc, "self")

	// List the root filesystem's top level beside the overlays, or create
	// the standard directories of an in-memory root.
	if v.host != nil {
		v.loadHostChildren(v.root, "/")
	} else {
		v.mkDir(v.root, "etc")
		v.mkDir(v.root, "bin")
		usr := v.mkDir(v.root, "usr")
		v.mkDir(usr, "bin")
		v.mkDir(v.root, "root")
	}

	// Add the configured files.
	for path, content := range cfg.Files {
		v.addFile(path, content, 0o644)
	}
}
