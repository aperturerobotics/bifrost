package kernel

import (
	"bytes"
	"cmp"
	"context"
	"strings"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// vforkState is the parent's state saved while a vfork child runs in the
// parent's instance. The parent resumes from it when the child execs or exits.
//
// The child runs on the parent's goroutine and linear memory, so its own
// bookkeeping and allocations change the parent's memory. memImage restores
// the memory, and snapshot the engine's call stack.
type vforkState struct {
	// outer is the state of the vfork the parent itself runs in, since a
	// vfork child may vfork again, as a subshell does for a pipeline.
	outer *vforkState

	snapshot     engine.Snapshot
	stackPointer uint32
	memImage     []byte

	// childPID is the PID of the child running in the parent's instance.
	childPID int
	// parent and parentSig are the parent's process and signal state.
	parent    procState
	parentSig sigState
}

// posixVfork saves the parent's state and turns the process into the vfork
// child, which runs until it execs or exits. The parent then resumes here
// with the child's PID as the result.
func (p *Process) posixVfork(ctx context.Context) uint64 {
	// Snapshot the engine state.
	snapshotter := p.kernel.runtime.Snapshotter(ctx)
	if snapshotter == nil {
		return negErrno(ENOSYS)
	}
	vf := &vforkState{outer: p.vfork, snapshot: snapshotter.SnapshotOnce()}

	// Capture the stack pointer and all of linear memory. The child runs on
	// through the callers of vfork and overwrites their locals on the C
	// stack, which the engine snapshot does not hold.
	if g := p.inst.ExportedGlobal(guestStackPointer); g != nil {
		vf.stackPointer = uint32(g.Get())
	}
	mem := p.inst.Memory()
	if view, ok := mem.Read(0, mem.Size()); ok {
		vf.memImage = bytes.Clone(view)
	}

	// Save the process state, holding the parent's pipe ends open while the
	// child changes the descriptor table.
	vf.parent = p.clone()
	for _, of := range vf.parent.fds {
		of.hold()
	}

	// Become the child.
	vf.childPID = p.kernel.procs.allocPID()
	p.sigMu.Lock()
	vf.parentSig = p.sigState
	p.vfork = vf
	p.sigMu.Unlock()
	p.ppid = p.pid
	p.pid = vf.childPID
	return 0
}

// sysExecve runs the program at pathPtr in place of the calling program.
func (p *Process) sysExecve(ctx context.Context, pathPtr, argvPtr, envpPtr uint32) uint64 {
	// Read the path, arguments and environment, then run the program.
	mem := p.inst.Memory()
	path := readCString(mem, pathPtr)
	args := readArgvFromMemory(mem, argvPtr)
	env := readEnvpFromMemory(mem, envpPtr)
	return p.exec(ctx, path, args, env)
}

// exec runs the program at path in place of the calling program, with env
// as its environment when env is set. It returns only an errno: a vfork
// child resumes its parent, and any other process exits for its
// replacement.
func (p *Process) exec(ctx context.Context, path string, args []string, env map[string]string) uint64 {
	// Read the program.
	node := p.kernel.vfs.lookup(resolvePath(p.cwd, p.fds, atFDCWD, path))
	if node == nil {
		return negErrno(ENOENT)
	}
	wasm, err := p.kernel.vfs.readFile(node)
	if err != nil {
		return negErrno(EIO)
	}
	if len(wasm) == 0 {
		return negErrno(ENOENT)
	}

	// Outside a vfork, the program replaces this process, as a shell's exec
	// does. The caller waits on the replacement.
	if p.vfork == nil {
		p.selfExecProc = p.execChild(p.pid, p.ppid, args, env)
		_ = p.inst.Close(ctx)
		p.kernel.startChildProcess(context.Background(), p.selfExecProc, wasm)
		panic(&ProcessExitError{})
	}

	// In a vfork child, start the program as the child process, drop the
	// child's descriptors and resume the parent.
	vf := p.vfork
	child := p.execChild(vf.childPID, vf.parent.pid, args, env)
	p.kernel.startChildProcess(ctx, child, wasm)
	for _, of := range p.fds {
		of.unref()
	}
	p.resumeParentAfterChild(vf)
	return 0
}

// execChild registers the process that runs an executed program. It shares
// p's descriptors and working directory, and takes env when it is set.
func (p *Process) execChild(pid, ppid int, args []string, env map[string]string) *Process {
	// Share the descriptors, directory and environment.
	child := newProcess(p.kernel, pid, args)
	child.ppid = ppid
	child.cwd = p.cwd
	child.fds = make(map[uint32]*OpenFile, len(p.fds))
	for fd, of := range p.fds {
		of.ref()
		child.fds[fd] = of
	}
	if env != nil {
		child.env = env
	}

	// Keep the signal mask, pending signals and ignored signals, as exec
	// does. A terminating signal that reached p during the exec ends the
	// program when it starts.
	p.sigMu.Lock()
	child.sigMask = p.sigMask
	child.sigPending = p.sigPending
	child.sigTerm = p.sigTerm
	p.sigTerm = 0
	for sig, act := range p.sigActions {
		if act.handler == sigIgn {
			child.sigActions[sig].handler = sigIgn
		}
	}
	p.sigMu.Unlock()

	// Register the program under its PID.
	p.kernel.procs.set(pid, child)
	return child
}

// resumeParentAfterChild restores the parent's state and memory, then
// resumes the parent at its vfork call with the child's PID as the result.
// The signal state returns together with the end of the vfork, so a signal
// never meets the parent with the child's dispositions. It does not return:
// restoring the snapshot unwinds the engine stack back to the vfork.
func (p *Process) resumeParentAfterChild(vf *vforkState) {
	// Restore the process and signal state.
	for _, of := range vf.parent.fds {
		of.restoreHeld()
	}
	p.procState = vf.parent
	p.sigMu.Lock()
	p.sigState = vf.parentSig
	p.vfork = vf.outer
	p.sigTerm = 0
	p.sigMu.Unlock()

	// Restore the memory and resume.
	p.inst.Memory().Write(0, vf.memImage)
	if g := p.inst.ExportedGlobal(guestStackPointer); g != nil {
		g.Set(uint64(vf.stackPointer))
	}
	vf.snapshot.Restore([]uint64{uint64(uint32(vf.childPID))})
}

// exitVforkChild ends the vfork child with exitCode and resumes the parent.
// The child's process stays in the table for the parent's wait4.
func (p *Process) exitVforkChild(exitCode uint32) {
	// Record the exited child.
	vf := p.vfork
	child := newProcess(p.kernel, vf.childPID, nil)
	child.ppid = vf.parent.pid
	child.exitCode = int(exitCode)
	close(child.exitCh)
	p.kernel.procs.set(vf.childPID, child)

	// Drop the child's descriptors and resume the parent.
	for _, of := range p.fds {
		of.unref()
	}
	p.resumeParentAfterChild(vf)
}

// sysWait4 waits for a child to exit, writes its wait status, and reaps it.
// A pid of zero or less waits for any child. With WNOHANG it returns zero
// while the child runs.
func (p *Process) sysWait4(ctx context.Context, pid int32, statusPtr, options, rusagePtr uint32) uint64 {
	// Find the child.
	var child *Process
	if pid > 0 {
		child = p.kernel.procs.get(int(pid))
	} else {
		child = p.kernel.procs.childOf(p.pid)
	}
	if child == nil {
		return negErrno(ECHILD)
	}

	// Wait for it to exit. A caught signal interrupts the wait.
	if options&wnohang != 0 {
		select {
		case <-child.exitCh:
		default:
			return 0
		}
	}
	ctx, release := p.interruptible(ctx)
	defer release()
	select {
	case <-ctx.Done():
		return negErrno(EINTR)
	case <-child.exitCh:
	}

	// Report a normal exit with the exit code in bits 8 to 15, and empty
	// resource usage.
	mem := p.inst.Memory()
	if statusPtr != 0 {
		mem.WriteUint32Le(statusPtr, uint32(child.exitCode&0xff)<<8)
	}
	if rusagePtr != 0 {
		mem.Write(rusagePtr, make([]byte, rusageSize))
	}

	// Reap it.
	p.kernel.procs.remove(child.pid)
	return uint64(uint32(child.pid))
}

// rusageSize is the size of a wasm32 struct rusage.
const rusageSize = 72

// posixExecvp runs a program found by name in the PATH directories, or at
// file when it names a path.
func (p *Process) posixExecvp(ctx context.Context, file string, argvPtr uint32) uint64 {
	// Run a path directly.
	args := readArgvFromMemory(p.inst.Memory(), argvPtr)
	if strings.Contains(file, "/") {
		return p.exec(ctx, file, args, nil)
	}

	// Search the PATH. An empty entry means the working directory.
	for dir := range strings.SplitSeq(cmp.Or(p.env["PATH"], "/bin:/usr/bin"), ":") {
		path := cmp.Or(dir, ".") + "/" + file
		if p.kernel.vfs.lookup(resolvePath(p.cwd, p.fds, atFDCWD, path)) != nil {
			return p.exec(ctx, path, args, nil)
		}
	}
	return negErrno(ENOENT)
}

// posixExit ends the program with code. In a vfork child, it ends the child
// and resumes the parent instead.
func (p *Process) posixExit(code uint32) {
	if p.vfork != nil {
		p.exitVforkChild(code)
	}
	panic(&ProcessExitError{Code: int(code)})
}

// readArgvFromMemory reads a null-terminated argv array from guest memory.
func readArgvFromMemory(mem engine.Memory, ptr uint32) []string {
	if ptr == 0 {
		return nil
	}
	var args []string
	for i := uint32(0); ; i++ {
		argPtr, ok := mem.ReadUint32Le(ptr + i*4)
		if !ok || argPtr == 0 {
			break
		}
		args = append(args, readCString(mem, argPtr))
	}
	return args
}

// readEnvpFromMemory reads a null-terminated envp array from guest memory
// into a map.
func readEnvpFromMemory(mem engine.Memory, ptr uint32) map[string]string {
	if ptr == 0 {
		return nil
	}
	env := make(map[string]string)
	for i := uint32(0); ; i++ {
		envPtr, ok := mem.ReadUint32Le(ptr + i*4)
		if !ok || envPtr == 0 {
			break
		}
		if key, value, ok := strings.Cut(readCString(mem, envPtr), "="); ok {
			env[key] = value
		}
	}
	return env
}
