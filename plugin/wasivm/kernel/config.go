package kernel

import (
	"io"

	"github.com/go-git/go-billy/v6"
)

// Config configures a kernel.
type Config struct {
	// Stdout receives the console's output. It defaults to os.Stdout.
	Stdout io.Writer
	// Stderr receives the console's error output. It defaults to os.Stderr.
	Stderr io.Writer
	// Stdin is the console's input. It defaults to os.Stdin.
	Stdin io.Reader

	// DebugSyscalls logs every syscall to Stderr.
	DebugSyscalls bool

	// TerminalFD is the host terminal behind the console. A positive
	// descriptor makes the console that terminal, and -1 makes it a plain
	// stream. 0 is standard input, and also a plain stream when Stdin is nil.
	TerminalFD int
	// VirtualTerminal makes the console a terminal the kernel emulates in
	// place of a host terminal, applying the termios line discipline itself:
	// echo, line editing and signal characters. Resize sets its window size.
	// It overrides TerminalFD.
	VirtualTerminal bool

	// RootFS is the root filesystem. In-memory nodes, such as /dev, /proc and
	// /tmp, overlay it, and writes to its files reach it. When nil, the
	// filesystem is held only in memory. The kernel does not close it.
	RootFS billy.Filesystem
	// Files adds in-memory files, keyed by absolute path such as
	// "/etc/hostname", creating their parent directories. A WebAssembly
	// program is installed executable and any other file read-write.
	Files map[string][]byte

	// Env is the processes' environment. When nil, it holds PATH, HOME, TERM
	// and SHELL.
	Env map[string]string

	// InitPath is the program Boot runs. It defaults to "/init".
	InitPath string
	// InitArgs are the arguments Boot passes after the program's path.
	InitArgs []string
}
