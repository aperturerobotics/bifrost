package kernel

import (
	"strings"
	"testing"
)

// newProcTestKernel returns a kernel over an in-memory root with init, PID
// 1, and its child, PID 2, which is the calling process.
func newProcTestKernel() (*Kernel, *Process) {
	// Start the kernel, then add the processes.
	k := &Kernel{procs: newProcessTable(), env: map[string]string{"HOME": "/"}}
	k.console = newTTY(k, false, false)
	k.initVFS(&Config{})
	for pid := range 2 {
		p := newProcess(k, pid+1, []string{"/bin/sh", "-c", "true"})
		p.exe = "/bin/sh"
		if pid != 0 {
			p.ppid, p.pgid, p.sid = initPID, initPID, initPID
		}
		k.procs.set(p.pid, p)
	}
	k.caller = k.procs.get(2)
	return k, k.caller
}

// readProcFile opens path as p and returns the file's content.
func readProcFile(t *testing.T, p *Process, path string) string {
	// Open the file, then read it whole.
	t.Helper()
	res := p.openResolved(path, oRDONLY, 0)
	if int64(res) < 0 {
		t.Fatalf("open %s: %#x", path, res)
	}
	fd := uint32(res)
	defer p.sysClose(fd)
	data, errno := p.kernel.readNodeBytes(p.fds[fd], 0, 1<<16)
	if errno != 0 {
		t.Fatalf("read %s: %#x", path, errno)
	}
	return string(data)
}

// TestProcFS checks the files /proc shows for the system and a process.
func TestProcFS(t *testing.T) {
	// Read each file as process 2.
	k, p := newProcTestKernel()
	tests := []struct {
		path, want string
	}{
		{path: "/proc/self/stat", want: "2 (sh) R 1 1 1 "},
		{path: "/proc/1/stat", want: "1 (sh) S 0 1 1 "},
		{path: "/proc/1/task/1/stat", want: "1 (sh) S 0 1 1 "},
		{path: "/proc/self/cmdline", want: "/bin/sh\x00-c\x00true\x00"},
		{path: "/proc/2/environ", want: "HOME=/\x00"},
		{path: "/proc/self/status", want: "Name:\tsh\nUmask:\t0022\nState:\tR\n"},
		{path: "/proc/mounts", want: "rootfs / rootfs rw 0 0\nproc /proc proc rw 0 0\n"},
	}
	for _, tc := range tests {
		if got := readProcFile(t, p, tc.path); !strings.HasPrefix(got, tc.want) {
			t.Fatalf("%s = %q, want prefix %q", tc.path, got, tc.want)
		}
	}

	// The links follow the process.
	p.cwd = "/tmp"
	if node := k.vfs.lookup("/proc/self/cwd"); node == nil || node != k.vfs.lookup("/tmp") {
		t.Fatalf("/proc/self/cwd does not lead to /tmp")
	}
	if node := k.vfs.lookupNoFollow("/proc/self/fd/0"); node == nil || node.symlink != "/dev/console" {
		t.Fatalf("/proc/self/fd/0 does not name the console")
	}

	// A reaped process leaves /proc.
	k.procs.remove(1)
	if k.vfs.lookup("/proc/1") != nil {
		t.Fatalf("/proc/1 outlives its process")
	}
}

// TestMount checks that a mount hides the directory under it until it is
// unmounted, and that the mount table follows.
func TestMount(t *testing.T) {
	// Leave a file in /tmp, then mount a tmpfs over it.
	k, p := newProcTestKernel()
	k.vfs.mkFile(k.vfs.lookup("/tmp"), "kept", []byte("kept"), 0o644)
	if res := p.sysMount("none", "/tmp", "tmpfs", 0, "size=1m"); res != 0 {
		t.Fatalf("mount = %#x", res)
	}
	if k.vfs.lookup("/tmp/kept") != nil {
		t.Fatalf("the tmpfs shows the covered file")
	}
	if got, want := readProcFile(t, p, "/proc/mounts"), "none /tmp tmpfs rw,size=1m 0 0\n"; !strings.HasSuffix(got, want) {
		t.Fatalf("/proc/mounts = %q, want suffix %q", got, want)
	}

	// Remount it read-only.
	if res := p.sysMount("", "/tmp", "", msRemount|msRdonly, ""); res != 0 {
		t.Fatalf("remount = %#x", res)
	}
	if got := readProcFile(t, p, "/proc/mounts"); !strings.HasSuffix(got, "none /tmp tmpfs ro 0 0\n") {
		t.Fatalf("/proc/mounts after remount = %q", got)
	}

	// Reject what Linux rejects.
	tests := []struct {
		name string
		res  uint64
		want uint64
	}{
		{name: "unknown type", res: p.sysMount("/dev/sda", "/tmp", "ext4", 0, ""), want: negErrno(ENODEV)},
		{name: "missing target", res: p.sysMount("none", "/missing", "tmpfs", 0, ""), want: negErrno(ENOENT)},
		{name: "remount a plain dir", res: p.sysMount("", "/dev", "", msRemount, ""), want: negErrno(EINVAL)},
		{name: "unmount a plain dir", res: p.sysUmount2("/dev"), want: negErrno(EINVAL)},
		{name: "unmount the root", res: p.sysUmount2("/"), want: negErrno(EBUSY)},
	}
	for _, tc := range tests {
		if tc.res != tc.want {
			t.Fatalf("%s = %#x, want %#x", tc.name, tc.res, tc.want)
		}
	}

	// Unmount it and find the file again.
	if res := p.sysUmount2("/tmp"); res != 0 {
		t.Fatalf("umount = %#x", res)
	}
	if k.vfs.lookup("/tmp/kept") == nil {
		t.Fatalf("the covered file is gone after umount")
	}
	if got := readProcFile(t, p, "/proc/mounts"); strings.Contains(got, "/tmp") {
		t.Fatalf("/proc/mounts after umount = %q", got)
	}
}
