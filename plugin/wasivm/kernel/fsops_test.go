package kernel

import "testing"

// fsTestProcess returns a process with four pages of memory and an in-memory
// VFS holding /tmp.
func fsTestProcess() (*Process, *memMemory) {
	p, mem := newMemProcess(4 * wasmPageSize)
	p.kernel.vfs = newVFS()
	p.kernel.vfs.mkDir(p.kernel.vfs.root, "tmp")
	return p, mem
}

// openTestFile creates /tmp/name holding content and opens it, returning the
// descriptor and the node.
func openTestFile(p *Process, name string, content []byte) (uint32, *VNode) {
	node := p.kernel.vfs.addFile("/tmp/"+name, content, 0o644)
	fd := p.allocFD()
	p.fds[fd] = &OpenFile{node: node, path: "/tmp/" + name}
	return fd, node
}

// TestSendfileAdvancesOffset checks sendfile without an offset pointer copies
// from the source descriptor's offset and advances it.
func TestSendfileAdvancesOffset(t *testing.T) {
	// Copy four bytes.
	p, _ := fsTestProcess()
	inFD, _ := openTestFile(p, "in", []byte("abcdefgh"))
	outFD, out := openTestFile(p, "out", nil)
	if res := p.sysSendfile(outFD, inFD, 0, 4); res != 4 {
		t.Fatalf("sendfile = %#x, want 4", res)
	}

	// Check the copy and the offset.
	if string(out.data) != "abcd" || p.fds[inFD].offset != 4 {
		t.Fatalf("copied %q to offset %d, want %q to offset 4", out.data, p.fds[inFD].offset, "abcd")
	}
}

// TestSendfileOffsetPointer checks sendfile with an offset pointer copies
// from that offset and advances it, leaving the descriptor's offset alone.
func TestSendfileOffsetPointer(t *testing.T) {
	// Copy three bytes from offset 2.
	p, mem := fsTestProcess()
	inFD, _ := openTestFile(p, "in", []byte("abcdefgh"))
	outFD, out := openTestFile(p, "out", nil)
	const offPtr = 2048
	mem.WriteUint32Le(offPtr, 2)
	if res := p.sysSendfile(outFD, inFD, offPtr, 3); res != 3 {
		t.Fatalf("sendfile = %#x, want 3", res)
	}

	// Check the copy and both offsets.
	off, _ := mem.ReadUint32Le(offPtr)
	if string(out.data) != "cde" || off != 5 || p.fds[inFD].offset != 0 {
		t.Fatalf("copied %q, pointer offset %d, descriptor offset %d; want %q, 5, 0", out.data, off, p.fds[inFD].offset, "cde")
	}
}

// TestMknodat checks mknodat creates a FIFO and a regular file, and fails with
// EEXIST on an existing path.
func TestMknodat(t *testing.T) {
	p, mem := fsTestProcess()
	tests := []struct {
		path     string
		mode     uint32
		wantType uint32
		want     uint64
	}{
		{path: "/tmp/fifo", mode: sIFIFO | 0o644, wantType: sIFIFO},
		{path: "/tmp/fifo", mode: sIFIFO | 0o644, wantType: sIFIFO, want: negErrno(EEXIST)},
		{path: "/tmp/reg", mode: 0o600, wantType: sIFREG},
	}
	for _, tc := range tests {
		// Create the node.
		const pathPtr = 2048
		mem.writeCString(pathPtr, tc.path)
		if res := p.sysMknodat(atFDCWD, pathPtr, tc.mode, 0); res != tc.want {
			t.Fatalf("mknodat(%s, %#o) = %#x, want %#x", tc.path, tc.mode, res, tc.want)
		}

		// Check its type.
		if node := p.kernel.vfs.lookup(tc.path); node == nil || node.mode&sIFMT != tc.wantType {
			t.Fatalf("%s is %+v, want type %#o", tc.path, node, tc.wantType)
		}
	}
}

// TestFlockValidatesOperation checks flock accepts valid operations and fails
// for a bad descriptor or operation.
func TestFlockValidatesOperation(t *testing.T) {
	p, _ := fsTestProcess()
	fd, _ := openTestFile(p, "f", nil)
	tests := []struct {
		fd, op uint32
		want   uint64
	}{
		{fd: fd, op: lockEX},
		{fd: fd, op: lockSH | lockNB},
		{fd: fd, op: 0, want: negErrno(EINVAL)},
		{fd: 999, op: lockEX, want: negErrno(EBADF)},
	}
	for _, tc := range tests {
		if res := p.sysFlock(tc.fd, tc.op); res != tc.want {
			t.Fatalf("flock(%d, %#x) = %#x, want %#x", tc.fd, tc.op, res, tc.want)
		}
	}
}
