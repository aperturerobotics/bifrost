package kernel

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestWASIReadDirRoot checks that listing the root shows both the root
// filesystem's entries and the kernel's in-memory directories.
func TestWASIReadDirRoot(t *testing.T) {
	// Open the root.
	p, mem := newWASITestProcess(t)
	fd := wasiTestOpen(t, p, mem, "/", wasiOflagsDirectory)

	// List it.
	if errno := p.wasiFDReadDir(mem, fd, 256, 4096, 0, 196); errno != wasiErrnoSuccess {
		t.Fatalf("fd_readdir(/): errno %d", errno)
	}
	names := readWASIDirentNames(t, mem, 256, 196)
	for _, name := range []string{"usr", "dev", "proc"} {
		if !names[name] {
			t.Errorf("fd_readdir(/) lacks %q: %v", name, names)
		}
	}
}

// TestWASIFileStat checks the file type and size of root filesystem, device
// and in-memory paths, with and without following a final symlink.
func TestWASIFileStat(t *testing.T) {
	p, mem := newWASITestProcess(t)
	cases := []struct {
		path     string
		flags    uint32
		wantType uint8
		wantSize int
	}{
		{"/bin", 0, wasiFiletypeSymbolicLink, len("usr/bin")},
		{"/usr/bin/busybox", wasiLookupSymlinkFollow, wasiFiletypeRegular, len("busybox-bytes")},
		{"/bin/busybox", wasiLookupSymlinkFollow, wasiFiletypeRegular, len("busybox-bytes")},
		{"/dev/null", wasiLookupSymlinkFollow, wasiFiletypeCharacter, -1},
		{"/proc/self", wasiLookupSymlinkFollow, wasiFiletypeDirectory, -1},
	}
	for _, c := range cases {
		pathPtr, pathLen := writeWASIPath(t, mem, c.path)
		if errno := p.wasiPathFileStatGet(mem, wasiAtFDCWD, c.flags, pathPtr, pathLen, 512); errno != wasiErrnoSuccess {
			t.Errorf("path_filestat_get(%s): errno %d", c.path, errno)
			continue
		}
		filetype, _ := mem.ReadByteAt(512 + 16)
		size, _ := mem.ReadUint64Le(512 + 32)
		if filetype != c.wantType || (c.wantSize >= 0 && size != uint64(c.wantSize)) {
			t.Errorf("path_filestat_get(%s): type %d size %d, want type %d size %d", c.path, filetype, size, c.wantType, c.wantSize)
		}
	}
}

// TestWASIOpenReadClose checks that a descriptor path_open returns reads with
// fd_read and closes with the POSIX close.
func TestWASIOpenReadClose(t *testing.T) {
	// Open and read the file.
	p, mem := newWASITestProcess(t)
	fd := wasiTestOpen(t, p, mem, "/etc/message", 0)
	mem.WriteUint32Le(192, 256)
	mem.WriteUint32Le(196, 32)
	if errno := p.wasiFDRead(context.Background(), mem, fd, 192, 1, 224); errno != wasiErrnoSuccess {
		t.Fatalf("fd_read: errno %d", errno)
	}
	nread, _ := mem.ReadUint32Le(224)
	if got, _ := mem.Read(256, nread); string(got) != "hello from wasi\n" {
		t.Fatalf("fd_read read %q", got)
	}

	// Close it.
	if res := p.sysClose(fd); res != 0 {
		t.Fatalf("close: %d", int64(res))
	}
	if _, ok := p.fds[fd]; ok {
		t.Fatalf("fd %d open after close", fd)
	}
}

// TestWASIPathsBesideDescriptor3 checks that the kernel resolves paths
// without a root descriptor: with a file the program opened on descriptor 3,
// an absolute path opens from the root, a relative one from the working
// directory, and AT_FDCWD reports the working directory.
func TestWASIPathsBesideDescriptor3(t *testing.T) {
	// Put a file on descriptor 3, then open paths beside it.
	p, mem := newWASITestProcess(t)
	p.setFD(3, &OpenFile{node: p.kernel.vfs.lookup("/etc/message"), path: "/etc/message"})
	wasiTestOpen(t, p, mem, "/usr/bin/busybox", 0)
	p.cwd = "/etc"
	wasiTestOpen(t, p, mem, "message", 0)

	// openat reads the working directory's rights through AT_FDCWD.
	if errno := p.wasiFDFdstatGet(mem, wasiAtFDCWD, 256); errno != wasiErrnoSuccess {
		t.Fatalf("fd_fdstat_get(AT_FDCWD): errno %d", errno)
	}
	if filetype, _ := mem.ReadByteAt(256); filetype != wasiFiletypeDirectory {
		t.Errorf("AT_FDCWD filetype %d", filetype)
	}
}

// TestGetcwd checks that getcwd reports the directory chdir set, and ERANGE for
// a buffer too small to hold it with its NUL.
func TestGetcwd(t *testing.T) {
	// Change to /etc.
	p, mem := newWASITestProcess(t)
	if r := p.chdirResolved("/etc"); r != 0 {
		t.Fatalf("chdir: %d", int32(r))
	}

	// Read it back into a buffer that fits, then one that does not.
	if r := p.sysGetcwd(256, 5); r != 5 {
		t.Fatalf("getcwd: %d", int32(r))
	}
	if got, _ := mem.Read(256, 5); string(got) != "/etc\x00" {
		t.Errorf("getcwd wrote %q", got)
	}
	if r := p.sysGetcwd(256, 4); r != negErrno(ERANGE) {
		t.Errorf("getcwd into 4 bytes: %d", int32(r))
	}
}

// newWASITestProcess returns a process with test memory over a root
// filesystem holding /usr/bin/busybox, /etc/message and a /bin symlink to
// usr/bin.
func newWASITestProcess(t *testing.T) (*Process, *memMemory) {
	// Build the root filesystem.
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"usr/bin/busybox": "busybox-bytes",
		"etc/message":     "hello from wasi\n",
	}
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("usr/bin", filepath.Join(root, "bin")); err != nil {
		t.Fatal(err)
	}

	// Start a kernel over it with one process.
	k := &Kernel{procs: newProcessTable(), env: make(map[string]string)}
	k.initVFS(&Config{RootFS: hostFS(t, root)})
	p := newProcess(k, 1, nil)
	mem := newMemMemory(8192)
	p.inst = memInstance{mem: mem}
	k.procs.set(1, p)
	return p, mem
}

// wasiTestOpen opens path with path_open.
func wasiTestOpen(t *testing.T, p *Process, mem *memMemory, path string, oflags uint32) uint32 {
	// Open the path and read back the descriptor.
	t.Helper()
	pathPtr, pathLen := writeWASIPath(t, mem, path)
	if errno := p.wasiPathOpen(mem, wasiAtFDCWD, pathPtr, pathLen, oflags, 0, 128); errno != wasiErrnoSuccess {
		t.Fatalf("path_open(%s): errno %d", path, errno)
	}
	fd, _ := mem.ReadUint32Le(128)
	return fd
}

// writeWASIPath writes path to test memory at offset 64 and returns its
// pointer and length.
func writeWASIPath(t *testing.T, mem *memMemory, path string) (uint32, uint32) {
	t.Helper()
	if !mem.Write(64, []byte(path)) {
		t.Fatalf("write path %q", path)
	}
	return 64, uint32(len(path))
}

// readWASIDirentNames returns the names of the directory entries fd_readdir
// wrote at bufPtr, filling the size at usedPtr.
func readWASIDirentNames(t *testing.T, mem *memMemory, bufPtr, usedPtr uint32) map[string]bool {
	// Read the size.
	t.Helper()
	used, ok := mem.ReadUint32Le(usedPtr)
	if !ok {
		t.Fatal("fd_readdir wrote no size")
	}

	// Walk the entries.
	names := make(map[string]bool)
	for off := uint32(0); off < used; {
		nameLen, _ := mem.ReadUint32Le(bufPtr + off + 16)
		name, _ := mem.Read(bufPtr+off+wasiDirentHeaderSize, nameLen)
		names[string(name)] = true
		off += wasiDirentHeaderSize + nameLen
	}
	return names
}
