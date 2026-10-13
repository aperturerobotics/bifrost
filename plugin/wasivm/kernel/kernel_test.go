package kernel

import (
	"bytes"
	"context"
	"encoding/binary"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

// rootfsDir returns the built guest root. It skips the test when the root is
// not built.
func rootfsDir(t *testing.T) string {
	// Tests run in the package directory, beside the root's directory.
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "rootfs", "out"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("root not built: %v", err)
	}
	return dir
}

// busyboxWasm returns the root's BusyBox binary. It skips the test when the
// root is not built.
func busyboxWasm(t *testing.T) []byte {
	t.Helper()
	wasm, err := os.ReadFile(filepath.Join(rootfsDir(t), "usr", "bin", "busybox"))
	if err != nil {
		t.Skipf("busybox not built: %v", err)
	}
	return wasm
}

// shellTest is a BusyBox run and the output it must produce.
type shellTest struct {
	name string
	// root boots on the built root. Otherwise the VFS holds the BusyBox
	// applets and files.
	root bool
	// files adds files to the VFS.
	files map[string][]byte
	// terminal runs the shell as if on a host terminal.
	terminal bool
	// script is run with sh -c. When empty, args run with stdin as input.
	script string
	// args is the BusyBox command line. It defaults to sh.
	args  []string
	stdin string
	// exit is the expected exit code.
	exit int
	// want lists text stdout contains, and wantStderr text stderr contains.
	want, wantStderr []string
	// notWant lists text stdout lacks, and notStderr text stderr lacks.
	notWant, notStderr []string
}

// run boots a kernel for the test and runs its command.
func (tc *shellTest) run(t *testing.T) (stdout, stderr string) {
	// Configure the kernel's I/O and environment.
	bb := busyboxWasm(t)
	var outBuf, errBuf bytes.Buffer
	cfg := &Config{
		Stdout:     &outBuf,
		Stderr:     &errBuf,
		Stdin:      strings.NewReader(tc.stdin),
		TerminalFD: -1,
		Files:      tc.files,
		Env: map[string]string{
			"PATH":     "/bin:/usr/bin:/sbin:/usr/sbin",
			"HOME":     "/root",
			"TERM":     "xterm",
			"GREETING": "hello-env",
		},
	}

	// Choose its filesystem and terminal.
	if tc.root {
		cfg.RootFS = hostFS(t, rootfsDir(t))
	} else {
		cfg.Files = BusyboxApplets(bb)
		maps.Copy(cfg.Files, tc.files)
	}
	if tc.terminal {
		cfg.TerminalFD = 99
	}

	// Start the kernel.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	k, err := New(ctx, testRuntime(t, ctx), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer k.Close(ctx)

	// Run the command and check its exit code.
	args := tc.args
	switch {
	case tc.script != "":
		args = []string{"sh", "-c", tc.script}
	case args == nil:
		args = []string{"sh"}
	}
	err = k.RunProcess(ctx, bb, args...)
	stdout, stderr = outBuf.String(), errBuf.String()
	var exitErr *ProcessExitError
	switch {
	case err == nil && tc.exit == 0:
	case errors.As(err, &exitErr) && exitErr.Code == tc.exit:
	default:
		t.Fatalf("exit %v, want code %d, stdout: %q, stderr: %q", err, tc.exit, stdout, stderr)
	}
	return stdout, stderr
}

// check reports the output the test wants and does not get.
func (tc *shellTest) check(t *testing.T, stdout, stderr string) {
	for _, s := range tc.want {
		if !strings.Contains(stdout, s) {
			t.Errorf("stdout lacks %q: %q, stderr: %q", s, stdout, stderr)
		}
	}
	for _, s := range tc.notWant {
		if strings.Contains(stdout, s) {
			t.Errorf("stdout has %q: %q", s, stdout)
		}
	}
	for _, s := range tc.wantStderr {
		if !strings.Contains(stderr, s) {
			t.Errorf("stderr lacks %q: %q", s, stderr)
		}
	}
	for _, s := range tc.notStderr {
		if strings.Contains(stderr, s) {
			t.Errorf("stderr has %q: %q", s, stderr)
		}
	}
}

func TestShell(t *testing.T) {
	year := strconv.Itoa(time.Now().UTC().Year())
	tests := []shellTest{{
		name: "appletArgv",
		args: []string{"echo", "hello-applet"},
		want: []string{"hello-applet"},
	}, {
		name:   "environment",
		script: "echo $GREETING",
		want:   []string{"hello-env"},
	}, {
		name:  "interactive",
		stdin: "echo interactive\npwd\nexit\n",
		want:  []string{"interactive", "/"},
	}, {
		name:      "terminal",
		terminal:  true,
		stdin:     "echo on-terminal\nexit\n",
		want:      []string{"on-terminal"},
		notStderr: []string{"fcntl"},
	}, {
		name:  "vfsFiles",
		files: map[string][]byte{"/tmp/greeting.txt": []byte("hello from the VFS\n")},
		stdin: "cd /tmp\npwd\necho written > out.txt\ncat greeting.txt out.txt\nexit\n",
		want:  []string{"/tmp", "hello from the VFS", "written"},
	}, {
		name:      "forkExec",
		script:    "/bin/busybox echo fork-exec",
		want:      []string{"fork-exec"},
		notStderr: []string{"sh: vfork", "sh: waitpid"},
	}, {
		name:    "childInheritsCwd",
		script:  "cd /dev && ls",
		want:    []string{"null", "urandom"},
		notWant: []string{"proc"},
	}, {
		name:   "pipeline",
		script: "echo HELLO | tr A-Z a-z | cat",
		want:   []string{"hello"},
	}, {
		// A fork child runs beside its parent: yes ends when head closes the
		// pipe, and the shell runs on while its background job sleeps.
		name:   "rootForkRunsBeside",
		root:   true,
		script: "yes | head -2; sleep 0 & echo started; wait; echo waited; (echo sub) | cat",
		want:   []string{"y\ny\nstarted\nwaited\nsub"},
	}, {
		name:       "dupStderr",
		script:     "exec 3>&2; echo via-three >&3",
		wantStderr: []string{"via-three"},
		notWant:    []string{"via-three"},
	}, {
		name:   "clock",
		script: "date -u +%Y",
		want:   []string{year},
	}, {
		name:   "rename",
		files:  map[string][]byte{"/tmp/before.txt": []byte("renamed\n")},
		script: "mv /tmp/before.txt /tmp/after.txt && cat /tmp/after.txt && test ! -e /tmp/before.txt && echo gone",
		want:   []string{"renamed", "gone"},
	}, {
		name:   "sleep",
		script: "sleep 0 && echo slept",
		want:   []string{"slept"},
	}, {
		// ash recovers from errors with longjmp, and the message reads errno.
		name:       "cdError",
		stdin:      "cd /nonexistent\necho survived\nexit\n",
		want:       []string{"survived"},
		wantStderr: []string{"No such file or directory"},
		notStderr:  []string{"Success"},
	}, {
		// The shell saves its history on exit after a failed lookup.
		name:       "commandNotFound",
		stdin:      "missing-command\nexit\n",
		exit:       127,
		wantStderr: []string{"missing-command: not found"},
	}, {
		name: "rootListing",
		root: true,
		args: []string{"ls", "/"},
		want: []string{"bin", "usr"},
	}, {
		name:      "rootPathLookup",
		root:      true,
		script:    "ls /",
		want:      []string{"bin"},
		notStderr: []string{"can't execute", "sh: vfork", "sh: waitpid"},
	}, {
		name:   "rootChdirExec",
		root:   true,
		script: "cd /usr/bin && ls",
		want:   []string{"busybox", "vim"},
	}, {
		name:   "rootStat",
		root:   true,
		script: "set -- $(ls -ls /usr/bin/busybox); test $1 -gt 0 && echo has-blocks; echo $2",
		want:   []string{"has-blocks", "-rwx"},
	}, {
		name:  "rootSequentialForks",
		root:  true,
		stdin: "( true ); ( true ); echo subshells\ntrue | true | true; echo pipes\nn=$(echo abc | wc -c); echo nested-$n\nexit\n",
		want:  []string{"subshells", "pipes", "nested-4"},
	}, {
		// Each of these programs imports its own set of host functions.
		name:   "rootPrograms",
		root:   true,
		script: "vim --version | head -1; tput -T xterm cols; echo hi | xxd",
		want:   []string{"VIM - Vi IMproved", "80", "00000000: 6869 0a"},
	}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr := tc.run(t)
			tc.check(t, stdout, stderr)
		})
	}
}

// TestBoot checks Boot runs init as PID 1 and reports how the run ended: an
// exited init or a halt returns nil, and a restart returns ErrReboot.
func TestBoot(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr error
	}{
		{name: "exit", command: "echo booted"},
		{name: "poweroff", command: "echo booted; /bin/busybox poweroff -f"},
		{name: "reboot", command: "echo booted; /bin/busybox reboot -f", wantErr: ErrReboot},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Start a kernel whose init runs a shell command.
			bb := busyboxWasm(t)
			var stdout bytes.Buffer
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			k, err := New(ctx, testRuntime(t, ctx), &Config{
				Stdout:   &stdout,
				Stderr:   &stdout,
				Stdin:    strings.NewReader(""),
				InitPath: "/bin/busybox",
				InitArgs: []string{"sh", "-c", tc.command},
				Files:    BusyboxApplets(bb),
			})
			if err != nil {
				t.Fatal(err)
			}
			defer k.Close(ctx)

			// Boot it and check how the run ended and the command's output.
			if err := k.Boot(ctx); !errors.Is(err, tc.wantErr) {
				t.Fatalf("Boot = %v, want %v; output: %q", err, tc.wantErr, stdout.String())
			}
			if !strings.Contains(stdout.String(), "booted") {
				t.Errorf("output: %q, want booted", stdout.String())
			}
		})
	}
}

// TestVFS checks lookups through nested directories and symlinks.
func TestVFS(t *testing.T) {
	// Build a tree with nested directories and files.
	v := newVFS()
	v.addFile("/etc/hostname", []byte("test-host\n"), 0o644)
	v.mkDirAll("/usr/local/bin")
	tmp := v.mkDir(v.root, "tmp")
	v.mkFile(tmp, "real.txt", []byte("real"), 0o644)

	// Link to the file directly, through another link and in a loop.
	v.mkSymlink(tmp, "relative", "real.txt")
	v.mkSymlink(tmp, "absolute", "/tmp/real.txt")
	v.mkSymlink(tmp, "chain", "relative")
	v.mkSymlink(tmp, "loop1", "loop2")
	v.mkSymlink(tmp, "loop2", "loop1")

	// Check what each path resolves to.
	tests := []struct {
		path string
		want string
	}{
		{"/etc/hostname", "test-host\n"},
		{"/tmp/real.txt", "real"},
		{"/tmp/relative", "real"},
		{"/tmp/absolute", "real"},
		{"/tmp/chain", "real"},
	}
	for _, tc := range tests {
		n := v.lookup(tc.path)
		if n == nil {
			t.Errorf("%s: not found", tc.path)
		} else if string(n.data) != tc.want {
			t.Errorf("%s: %q, want %q", tc.path, n.data, tc.want)
		}
	}
	for _, path := range []string{"/missing", "/tmp/loop1"} {
		if v.lookup(path) != nil {
			t.Errorf("%s: found, want nil", path)
		}
	}
	if v.lookup("/usr/local/bin") == nil {
		t.Error("/usr/local/bin: not found")
	}

	// A symlink looked up without following is the link itself.
	link := v.lookupNoFollow("/tmp/relative")
	if link == nil || link.mode&sIFMT != sIFLNK || link.symlink != "real.txt" {
		t.Errorf("lookupNoFollow /tmp/relative: %+v", link)
	}
	if got := cleanPath("/foo/../bar/./baz"); got != "/bar/baz" {
		t.Errorf("cleanPath: %q, want /bar/baz", got)
	}
}

func TestONLCR(t *testing.T) {
	tests := []struct {
		name  string
		raw   bool
		oflag uint32
		in    string
		want  string
	}{
		{"translates", true, opost | onlcr, "a\nb\n", "a\r\nb\r\n"},
		{"keepsCRLF", true, opost | onlcr, "a\r\nb\r\n", "a\r\nb\r\n"},
		{"onlcrOff", true, opost, "a\nb\n", "a\nb\n"},
		{"cooked", false, opost | onlcr, "a\nb\n", "a\nb\n"},
	}
	for _, tc := range tests {
		k := &Kernel{}
		k.termRaw.Store(tc.raw)
		binary.LittleEndian.PutUint32(k.termIOS[4:8], tc.oflag)
		if got := string(k.onlcrProcess([]byte(tc.in))); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestFchdirAndFsync(t *testing.T) {
	// Open /tmp/f.txt and /tmp in a bare process.
	k := &Kernel{procs: newProcessTable(), vfs: newVFS()}
	tmp := k.vfs.mkDir(k.vfs.root, "tmp")
	f := k.vfs.mkFile(tmp, "f.txt", nil, 0o644)
	p := newProcess(k, 1, nil)
	fileFD := p.allocFD()
	p.fds[fileFD] = &OpenFile{node: f, path: "/tmp/f.txt"}
	dirFD := p.allocFD()
	p.fds[dirFD] = &OpenFile{node: tmp, path: "/tmp"}

	// Change to the directory and sync the file, then use a closed fd.
	if r := p.sysFchdir(dirFD); r != 0 || p.cwd != "/tmp" {
		t.Errorf("fchdir: %d, cwd %q", int32(r), p.cwd)
	}
	if r := p.sysFsync(fileFD); r != 0 {
		t.Errorf("fsync: %d", int32(r))
	}
	if r := p.sysFchdir(999); r != negErrno(EBADF) {
		t.Errorf("fchdir on a closed fd: %d", int32(r))
	}
	if r := p.sysFsync(999); r != negErrno(EBADF) {
		t.Errorf("fsync on a closed fd: %d", int32(r))
	}
}
