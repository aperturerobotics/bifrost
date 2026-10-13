package kernel

import (
	"context"
	"crypto/rand"
	"os"
	"time"
)

// sysRead handles the read syscall. A blocking read returns EINTR when a
// signal is caught.
func (p *Process) sysRead(ctx context.Context, fd, bufPtr, count uint32) uint64 {
	// A read of nothing returns at once.
	if count == 0 {
		return 0
	}

	// Let a caught signal interrupt a blocking read.
	ctx, release := p.interruptible(ctx)
	defer release()

	// Read from the descriptor's pipe or file.
	of, ok := p.fds[fd]
	switch {
	case !ok:
		return negErrno(EBADF)
	case of.pipe != nil && of.pipe.reader:
		return p.readPipe(ctx, of, bufPtr, count)
	case of.node != nil:
		return p.readVFS(ctx, of, bufPtr, count)
	default:
		return negErrno(EBADF)
	}
}

// readStdin reads the kernel's terminal input into guest memory.
func (p *Process) readStdin(ctx context.Context, bufPtr, count uint32) uint64 {
	buf := make([]byte, count)
	n, _ := p.kernel.stdin.read(ctx, buf)
	return p.readResult(ctx, bufPtr, buf[:n])
}

// readPipe reads a pipe into guest memory.
func (p *Process) readPipe(ctx context.Context, of *OpenFile, bufPtr, count uint32) uint64 {
	buf := make([]byte, count)
	n, _ := of.pipe.read(ctx, buf)
	return p.readResult(ctx, bufPtr, buf[:n])
}

// readResult copies data read from a blocking source to guest memory. An
// empty read that a signal interrupted returns EINTR.
func (p *Process) readResult(ctx context.Context, bufPtr uint32, data []byte) uint64 {
	if len(data) == 0 && interrupted(ctx) {
		return negErrno(EINTR)
	}
	p.inst.Memory().Write(bufPtr, data)
	return uint64(len(data))
}

// readVFS reads from a device or regular file at the descriptor's offset.
func (p *Process) readVFS(ctx context.Context, of *OpenFile, bufPtr, count uint32) uint64 {
	// Read a character device.
	mem := p.inst.Memory()
	if of.node.mode&sIFMT == sIFCHR {
		switch of.node.rdev {
		case devNull:
			return 0
		case devZero:
			mem.Write(bufPtr, make([]byte, count))
			return uint64(count)
		case devURandom:
			buf := make([]byte, count)
			_, _ = rand.Read(buf)
			mem.Write(bufPtr, buf)
			return uint64(count)
		case devTTY, devConsole:
			return p.readStdin(ctx, bufPtr, count)
		}
		return negErrno(EIO)
	}

	// Read a regular file and advance the offset.
	data, errno := p.readNodeBytes(of, of.offset, int(count))
	if errno != 0 {
		return errno
	}
	mem.Write(bufPtr, data)
	of.offset += int64(len(data))
	return uint64(len(data))
}

// sysWrite handles the write syscall.
func (p *Process) sysWrite(ctx context.Context, fd, bufPtr, count uint32) uint64 {
	buf, ok := p.inst.Memory().Read(bufPtr, count)
	if !ok {
		return negErrno(EFAULT)
	}
	return p.writeOutBytes(ctx, fd, buf)
}

// writeOutBytes writes buf to a descriptor's pipe or file. write, writev and
// sendfile share it.
func (p *Process) writeOutBytes(ctx context.Context, fd uint32, buf []byte) uint64 {
	of, ok := p.fds[fd]
	switch {
	case !ok:
		return negErrno(EBADF)
	case of.pipe != nil && !of.pipe.reader:
		return p.writePipe(ctx, of, buf)
	case of.node != nil:
		return p.writeVFS(of, buf)
	default:
		return negErrno(EBADF)
	}
}

// writePipe writes buf to a pipe, waiting while it is full. A caught signal
// interrupts the wait. A write to a pipe without a reader raises SIGPIPE and
// fails with EPIPE.
func (p *Process) writePipe(ctx context.Context, of *OpenFile, buf []byte) uint64 {
	// Write, waiting while the pipe is full.
	ctx, release := p.interruptible(ctx)
	defer release()
	n, ok := of.pipe.write(ctx, buf)

	// Report a closed reader or an interrupted wait.
	switch {
	case !ok:
		p.deliverSignal(SIGPIPE)
		return negErrno(EPIPE)
	case n == 0 && len(buf) != 0 && interrupted(ctx):
		return negErrno(EINTR)
	}
	return uint64(n)
}

// writeVFS writes to a device or regular file at the descriptor's offset.
func (p *Process) writeVFS(of *OpenFile, buf []byte) uint64 {
	// Write a character device.
	if of.node.mode&sIFMT == sIFCHR {
		switch of.node.rdev {
		case devNull:
			return uint64(len(buf))
		case devTTY, devConsole:
			return p.writeConsole(of, buf)
		}
		return negErrno(EIO)
	}

	// Write a regular file and advance the offset.
	n, errno := p.writeNodeBytes(of, of.offset, buf)
	if errno != 0 {
		return errno
	}
	of.offset += int64(n)
	return uint64(n)
}

// writeConsole writes to the kernel's stdout, or its stderr for the standard
// error descriptor. The onlcrWriter on both applies output processing.
func (p *Process) writeConsole(of *OpenFile, buf []byte) uint64 {
	w := p.kernel.stdout
	if of.stderr {
		w = p.kernel.stderr
	}
	n, _ := w.Write(buf)
	return uint64(n)
}

// truncateNode resizes a file to size, zero-extending it or dropping its
// tail. truncate and ftruncate share it.
func (p *Process) truncateNode(node *VNode, size int64) uint64 {
	// Resize a host file.
	if node.hostPath != "" {
		if err := p.kernel.vfs.truncateHost(node.hostPath, size); err != nil {
			return negErrno(EIO)
		}
		node.size = size
		return 0
	}

	// Resize an in-memory file.
	node.data = resize(node.data, size)
	node.size = size
	return 0
}

// resize returns data cut or zero-extended to size bytes.
func resize(data []byte, size int64) []byte {
	if size <= int64(len(data)) {
		return data[:size]
	}
	return append(data, make([]byte, size-int64(len(data)))...)
}

// readNodeBytes reads up to count bytes of an open file at off. It returns
// fewer bytes at the end of the file, or a negative errno. It does not move
// the descriptor offset.
func (p *Process) readNodeBytes(of *OpenFile, off int64, count int) ([]byte, uint64) {
	// Read a host file.
	node := of.node
	if node == nil {
		return nil, negErrno(EBADF)
	}
	if node.hostPath != "" {
		f, err := p.kernel.vfs.host.Open(node.hostPath)
		if err != nil {
			return nil, negErrno(EIO)
		}
		defer f.Close()
		buf := make([]byte, count)
		n, _ := f.ReadAt(buf, off)
		return buf[:n], 0
	}

	// Copy from an in-memory file.
	if off >= node.size {
		return nil, 0
	}
	n := min(int64(count), node.size-off)
	return append([]byte(nil), node.data[off:off+n]...), 0
}

// writeNodeBytes writes data to an open file at off, growing the file as
// needed. It returns the bytes written or a negative errno. It does not move
// the descriptor offset.
func (p *Process) writeNodeBytes(of *OpenFile, off int64, data []byte) (int, uint64) {
	// Write a host file and take its new size.
	node := of.node
	if node == nil {
		return 0, negErrno(EBADF)
	}
	if node.hostPath != "" {
		f, err := p.kernel.vfs.host.OpenFile(node.hostPath, os.O_WRONLY, 0)
		if err != nil {
			return 0, negErrno(EIO)
		}
		defer f.Close()
		n, err := f.WriteAt(data, off)
		if err != nil {
			return 0, negErrno(EIO)
		}
		if info, err := f.Stat(); err == nil {
			node.size = info.Size()
		}
		node.mtime = time.Now()
		return n, 0
	}

	// Write into an in-memory file.
	end := off + int64(len(data))
	if end > int64(len(node.data)) {
		node.data = resize(node.data, end)
	}
	copy(node.data[off:], data)
	node.size = max(node.size, end)
	node.mtime = time.Now()
	return len(data), 0
}
