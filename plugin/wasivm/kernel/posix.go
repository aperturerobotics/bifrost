package kernel

import (
	"context"
	"io"
	"strconv"
	"strings"
)

// posixFunctions returns the env module: the POSIX functions wasi-libc
// declares but does not define. A guest built against those headers imports
// each one it calls.
func posixFunctions() map[string]hostFunc {
	return map[string]hostFunc{
		// Identity.
		"getpid":  {"i", func(_ context.Context, p *Process, _ []uint64) uint64 { return uint64(p.pid) }},
		"getppid": {"i", func(_ context.Context, p *Process, _ []uint64) uint64 { return uint64(p.ppid) }},
		"getuid":  {"i", func(_ context.Context, p *Process, _ []uint64) uint64 { return uint64(p.uid) }},
		"geteuid": {"i", func(_ context.Context, p *Process, _ []uint64) uint64 { return uint64(p.euid) }},
		"getgid":  {"i", func(_ context.Context, p *Process, _ []uint64) uint64 { return uint64(p.gid) }},
		"getegid": {"i", func(_ context.Context, p *Process, _ []uint64) uint64 { return uint64(p.egid) }},
		"getpgrp": {"i", func(_ context.Context, p *Process, _ []uint64) uint64 { return uint64(p.pgid) }},
		"setsid":  {"i", func(_ context.Context, p *Process, _ []uint64) uint64 { return uint64(p.sid) }},
		"setpgid": returns("iii", 0),
		"setuid": {"ii", func(_ context.Context, p *Process, a []uint64) uint64 {
			return p.sysSetuid(uint32(a[0]))
		}},
		"setgid": {"ii", func(_ context.Context, p *Process, a []uint64) uint64 {
			return p.sysSetgid(uint32(a[0]))
		}},
		"umask": {"ii", func(_ context.Context, p *Process, a []uint64) uint64 {
			old := p.umask
			p.umask = uint32(a[0]) & 0o777
			return uint64(old)
		}},
		"getgroups": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysGetgroups(uint32(a[0]), uint32(a[1])))
		}},
		"getgrouplist": {"iiiii", func(_ context.Context, p *Process, a []uint64) uint64 {
			p.inst.Memory().WriteUint32Le(uint32(a[3]), 0)
			return 0
		}},
		"getlogin_r": {"iii", func(_ context.Context, p *Process, a []uint64) uint64 {
			p.inst.Memory().Write(uint32(a[0]), []byte("root\x00"))
			return 0
		}},

		// The user and group databases are empty.
		"getpwuid": returns("ii", 0),
		"getpwnam": returns("ii", 0),
		"getgrnam": returns("ii", 0),
		"getgrgid": returns("ii", 0),
		"setpwent": returns("v", 0),
		"getpwent": returns("i", 0),
		"endpwent": returns("v", 0),

		// Process control. vfork is fork, which Linux permits.
		"vfork":        {"i", forkFunc},
		"fork":         {"i", forkFunc},
		"_exit":        {"vi", exitFunc},
		"__posix_exit": {"vi", exitFunc},
		"abort": {"v", func(_ context.Context, p *Process, _ []uint64) uint64 {
			p.abort()
			return 0
		}},
		"execve": {"iiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysExecve(ctx, uint32(a[0]), uint32(a[1]), uint32(a[2])))
		}},
		"execvp": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			file := readCString(p.inst.Memory(), uint32(a[0]))
			return p.posixResult(ctx, p.posixExecvp(ctx, file, uint32(a[1])))
		}},
		"waitpid": {"iiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			// ECHILD with WNOHANG reports that no child changed state, so a
			// reap loop ends without an error.
			options := uint32(a[2])
			result := p.sysWait4(ctx, int32(a[0]), uint32(a[1]), options, 0)
			if int32(result) == -ECHILD && options&wnohang != 0 {
				return 0
			}
			return p.posixResult(ctx, result)
		}},
		"execl":  returns("iiii", fail),
		"system": returns("ii", fail),

		// The linker imports the destructor hook; exit flushes stdio itself.
		"__wasm_call_dtors": returns("v", 0),

		// Signals.
		"kill": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysKill(int32(a[0]), uint32(a[1])))
		}},
		"raise": {"ii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysKill(int32(p.pid), uint32(a[0])))
		}},
		"killpg": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysKill(-int32(a[0]), uint32(a[1])))
		}},
		"sigaction": {"iiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysRtSigaction(uint32(a[0]), uint32(a[1]), uint32(a[2])))
		}},
		"signal": {"iii", func(_ context.Context, p *Process, a []uint64) uint64 {
			signum := uint32(a[0])
			if signum == 0 || signum >= 32 {
				return fail
			}
			return uint64(p.setSigAction(signum, &sigAction{handler: uint32(a[1])}).handler)
		}},
		"sigprocmask": {"iiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysRtSigprocmask(uint32(a[0]), uint32(a[1]), uint32(a[2])))
		}},
		"sigsuspend": {"ii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysRtSigsuspend(ctx, uint32(a[0])))
		}},
		"sigpending": {"ii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysRtSigpending(uint32(a[0])))
		}},
		"sigaltstack": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysSigaltstack(uint32(a[0]), uint32(a[1])))
		}},
		"setitimer": {"iiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysSetitimer(uint32(a[0]), uint32(a[1]), uint32(a[2])))
		}},
		"alarm": returns("ii", 0),

		// A sigset_t is 64 bits, bit n-1 standing for signal n.
		"sigemptyset": {"ii", func(_ context.Context, p *Process, a []uint64) uint64 {
			p.inst.Memory().WriteUint64Le(uint32(a[0]), 0)
			return 0
		}},
		"sigfillset": {"ii", func(_ context.Context, p *Process, a []uint64) uint64 {
			p.inst.Memory().WriteUint64Le(uint32(a[0]), ^uint64(0))
			return 0
		}},
		"sigaddset": {"iii", func(_ context.Context, p *Process, a []uint64) uint64 {
			mem := p.inst.Memory()
			set, _ := mem.ReadUint64Le(uint32(a[0]))
			mem.WriteUint64Le(uint32(a[0]), set|1<<(uint32(a[1])-1))
			return 0
		}},
		"sigismember": {"iii", func(_ context.Context, p *Process, a []uint64) uint64 {
			set, _ := p.inst.Memory().ReadUint64Le(uint32(a[0]))
			return set >> (uint32(a[1]) - 1) & 1
		}},

		// The signal headers declare SIG_IGN and SIG_ERR as functions.
		"__SIG_IGN": returns("vi", 0),
		"__SIG_ERR": returns("vi", 0),

		// Non-local jumps.
		"setjmp": {"ii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			p.setjmp(ctx, uint32(a[0]))
			return 0
		}},
		"sigsetjmp": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			p.setjmp(ctx, uint32(a[0]))
			return 0
		}},
		"longjmp":    {"vii", longjmpFunc},
		"siglongjmp": {"vii", longjmpFunc},

		// Terminal control.
		"tcgetattr": {"iii", func(_ context.Context, p *Process, a []uint64) uint64 {
			if !p.kernel.tty || !p.isConsoleFD(uint32(a[0])) {
				return fail
			}
			p.kernel.writeWasiTermios(p.inst.Memory(), uint32(a[1]))
			return 0
		}},
		"tcsetattr": {"iiii", func(_ context.Context, p *Process, a []uint64) uint64 {
			if !p.kernel.tty || !p.isConsoleFD(uint32(a[0])) {
				return fail
			}
			p.kernel.readWasiTermios(p.inst.Memory(), uint32(a[2]))
			return 0
		}},
		"tcgetpgrp": {"ii", func(_ context.Context, p *Process, _ []uint64) uint64 { return uint64(p.pgid) }},
		"tcsetpgrp": returns("iii", 0),
		"tcflush":   returns("iii", 0),
		"cfgetispeed": {"ii", func(_ context.Context, p *Process, a []uint64) uint64 {
			speed, _ := p.inst.Memory().ReadUint32Le(uint32(a[0]) + termiosIspeed)
			return uint64(speed)
		}},
		"cfgetospeed": {"ii", func(_ context.Context, p *Process, a []uint64) uint64 {
			speed, _ := p.inst.Memory().ReadUint32Le(uint32(a[0]) + termiosOspeed)
			return uint64(speed)
		}},
		"cfsetispeed": {"iii", func(_ context.Context, p *Process, a []uint64) uint64 {
			p.inst.Memory().WriteUint32Le(uint32(a[0])+termiosIspeed, uint32(a[1]))
			return 0
		}},
		"cfsetospeed": {"iii", func(_ context.Context, p *Process, a []uint64) uint64 {
			p.inst.Memory().WriteUint32Le(uint32(a[0])+termiosOspeed, uint32(a[1]))
			return 0
		}},
		"ttyname": {"ii", func(ctx context.Context, p *Process, _ []uint64) uint64 {
			// The kernel owns no guest memory to hold the name.
			p.setErrno(ctx, wasiErrno(ENOSYS))
			return 0
		}},
		"ttyname_r": {"iiii", func(_ context.Context, p *Process, a []uint64) uint64 {
			p.inst.Memory().Write(uint32(a[1]), []byte("/dev/tty\x00"))
			return 0
		}},

		// Pseudo-terminals are unsupported.
		"posix_openpt": {"ii", func(ctx context.Context, p *Process, _ []uint64) uint64 {
			return p.posixResult(ctx, enosys)
		}},
		"ptsname":  returns("ii", 0),
		"grantpt":  returns("ii", fail),
		"unlockpt": returns("ii", fail),

		// Files. __posix_fcntl3 replaces wasi-libc's fcntl, which rejects
		// F_DUPFD.
		"dup": {"ii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysDup(uint32(a[0])))
		}},
		"dup2": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysDup2(uint32(a[0]), uint32(a[1])))
		}},
		"__posix_fcntl3": {"iiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysFcntl(uint32(a[0]), uint32(a[1]), uint32(a[2])))
		}},
		"pipe": {"ii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysPipe2(uint32(a[0])))
		}},
		"chdir": {"ii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysChdir(uint32(a[0])))
		}},
		"fchdir": {"ii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysFchdir(uint32(a[0])))
		}},
		"flock": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysFlock(uint32(a[0]), uint32(a[1])))
		}},
		"sendfile": {"iiiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysSendfile(ctx, uint32(a[0]), uint32(a[1]), uint32(a[2]), uint32(a[3])))
		}},
		"mkfifo": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysMknodat(atFDCWD, uint32(a[0]), uint32(a[1])&0o7777|sIFIFO, 0))
		}},
		"mknod": {"iiij", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysMknodat(atFDCWD, uint32(a[0]), uint32(a[1]), uint32(a[2])))
		}},
		"mkdtemp": {"ii", func(_ context.Context, p *Process, a []uint64) uint64 {
			return uint64(p.mkdtemp(uint32(a[0])))
		}},
		"getxattr": {"iiiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysGetxattr(uint32(a[0]), uint32(a[1]), uint32(a[2]), uint32(a[3])))
		}},
		"listxattr": {"iiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysListxattr(uint32(a[0]), uint32(a[1]), uint32(a[2])))
		}},
		"setxattr": {"iiiiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysSetxattr(uint32(a[0]), uint32(a[1]), uint32(a[2]), uint32(a[3]), uint32(a[4])))
		}},
		"mkstemp": returns("ii", fail),
		"fchown":  returns("iiii", 0),
		"chown":   returns("iiii", 0),
		"lchown":  returns("iiii", 0),
		"chroot":  returns("ii", 0),
		"sync":    returns("v", 0),

		// There is one filesystem and no mount table.
		"setmntent": returns("iii", 0),
		"getmntent": returns("ii", 0),
		"endmntent": returns("ii", 1),

		// Resources and time.
		"getrlimit": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysGetrlimit(uint32(a[0]), uint32(a[1])))
		}},
		"setrlimit": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysSetrlimit(uint32(a[0]), uint32(a[1])))
		}},
		"getpriority": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysGetpriority(uint32(a[0]), uint32(a[1])))
		}},
		"setpriority": {"iiii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysSetpriority(uint32(a[0]), uint32(a[1]), uint32(a[2])))
		}},
		"clock_settime": {"iii", func(ctx context.Context, p *Process, a []uint64) uint64 {
			return p.posixResult(ctx, p.sysClockSettime(uint32(a[0]), uint32(a[1])))
		}},
		"sysinfo": {"ii", func(_ context.Context, p *Process, a []uint64) uint64 {
			p.inst.Memory().Write(uint32(a[0]), make([]byte, sysinfoSize))
			return 0
		}},
		"tzset": returns("v", 0),

		// Networking is unsupported.
		"socket":       returns("iiii", fail),
		"bind":         returns("iiii", fail),
		"listen":       returns("iii", fail),
		"connect":      returns("iiii", fail),
		"getaddrinfo":  returns("iiiii", fail),
		"freeaddrinfo": returns("vi", 0),
		"gai_strerror": returns("ii", 0),

		// Dynamic loading is unsupported.
		"dlopen":  returns("iii", 0),
		"dlclose": returns("ii", 0),
		"dlsym":   returns("iii", 0),
		"dlerror": returns("i", 0),
	}
}

// sysinfoSize is the size of the wasi-libc sysinfo struct.
const sysinfoSize = 64

// forkFunc starts a child process that runs beside the caller.
func forkFunc(ctx context.Context, p *Process, _ []uint64) uint64 {
	return p.posixFork(ctx)
}

// exitFunc ends the process with the status in its argument.
func exitFunc(_ context.Context, p *Process, a []uint64) uint64 {
	p.posixExit(uint32(a[0]))
	return 0
}

// longjmpFunc jumps back to the setjmp call for the jmp_buf in its first
// argument, which then returns the second argument, or 1 when that is 0.
func longjmpFunc(ctx context.Context, p *Process, a []uint64) uint64 {
	p.longjmp(ctx, uint32(a[0]), max(uint32(a[1]), 1))
	return 0
}

// abort raises SIGABRT and ends the process as SIGABRT would.
func (p *Process) abort() {
	p.sysKill(int32(p.pid), SIGABRT)
	p.posixExit(128 + SIGABRT)
}

// setjmp records the execution state for the jmp_buf at buf: the engine
// snapshot and the C stack between the stack pointer and the heap base.
func (p *Process) setjmp(ctx context.Context, buf uint32) {
	// Without snapshots, longjmp cannot return here.
	snapshotter := p.kernel.runtime.Snapshotter(ctx)
	if snapshotter == nil {
		return
	}

	// Copy the C stack, which a longjmp must restore along with the engine
	// state.
	entry := jmpBufEntry{snapshot: snapshotter.Snapshot()}
	var heapBase uint32
	if g := p.inst.ExportedGlobal(guestStackPointer); g != nil {
		entry.sp = uint32(g.Get())
	}
	if g := p.inst.ExportedGlobal("__heap_base"); g != nil {
		heapBase = uint32(g.Get())
	}
	if entry.sp > 0 && heapBase > entry.sp {
		if data, ok := p.inst.Memory().Read(entry.sp, heapBase-entry.sp); ok {
			entry.cstack = append([]byte(nil), data...)
		}
	}
	if p.jmpBufs == nil {
		p.jmpBufs = make(map[uint32]jmpBufEntry)
	}
	p.jmpBufs[buf] = entry
}

// longjmp restores the state setjmp recorded for buf, making that setjmp
// return val. A jump without a matching setjmp aborts the process.
func (p *Process) longjmp(ctx context.Context, buf, val uint32) {
	// Find the state.
	entry, ok := p.jmpBufs[buf]
	if !ok {
		_, _ = io.WriteString(p.kernel.stderr, "longjmp: no matching setjmp\n")
		p.abort()
		return
	}

	// Restore the C stack and the engine state.
	if g := p.inst.ExportedGlobal(guestStackPointer); g != nil {
		g.Set(uint64(entry.sp))
	}
	p.inst.Memory().Write(entry.sp, entry.cstack)
	entry.snapshot.Restore(ctx, []uint64{uint64(val)})
}

// mkdtemp replaces the trailing XXXXXX of the template at ptr with a suffix
// that names a new directory, and creates it. It returns ptr, or 0 on failure.
func (p *Process) mkdtemp(ptr uint32) uint32 {
	// The template must end with the six characters to replace.
	mem := p.inst.Memory()
	tmpl := readCString(mem, ptr)
	if len(tmpl) < 6 {
		return 0
	}

	// Try fresh suffixes until one names a directory that does not exist
	// yet, writing each name back into the template.
	prefix := tmpl[:len(tmpl)-6]
	for range 100 {
		suffix := strconv.FormatUint(uint64(p.kernel.tempSeq.Add(1)), 36)
		suffix = strings.Repeat("0", max(0, 6-len(suffix))) + suffix
		mem.Write(ptr, append([]byte(prefix+suffix[len(suffix)-6:]), 0))
		result := p.sysMkdirat(atFDCWD, ptr, 0o700)
		if result == negErrno(EEXIST) {
			continue
		}
		if int32(result) < 0 {
			return 0
		}
		return ptr
	}
	return 0
}

// posixResult converts a Linux syscall result to the C convention: a
// negative errno becomes -1 with errno set in the guest.
func (p *Process) posixResult(ctx context.Context, result uint64) uint64 {
	if int32(result) >= 0 {
		return result
	}
	p.setErrno(ctx, wasiErrno(uint32(-int32(result))))
	return fail
}

// setErrno sets the guest's errno to the WASI errno value. A guest that does
// not export __errno_location keeps its errno.
func (p *Process) setErrno(ctx context.Context, errno uint32) {
	// Find errno, then set it.
	fn := p.inst.ExportedFunction("__errno_location")
	if fn == nil {
		return
	}
	results, err := fn.Call(ctx)
	if err != nil || len(results) == 0 {
		return
	}
	p.inst.Memory().WriteUint32Le(uint32(results[0]), errno)
}
