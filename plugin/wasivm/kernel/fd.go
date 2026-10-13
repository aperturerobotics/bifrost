package kernel

import "os"

// openResolved opens an already-resolved VFS path and installs it in the
// process fd table. POSIX openat and WASI path_open share this owner path so
// descriptors are readable, seekable, and closeable through either ABI.
func (p *Process) openResolved(resolved string, flags, mode uint32) uint64 {
	// Find the file, creating it when asked.
	node := p.kernel.vfs.lookup(resolved)
	switch {
	case node == nil && flags&oCREAT == 0:
		return negErrno(ENOENT)
	case node == nil:
		var errno uint64
		if node, errno = p.createFile(resolved, flags, mode&0o7777); errno != 0 {
			return errno
		}
	case flags&oCREAT != 0 && flags&oEXCL != 0:
		return negErrno(EEXIST)
	}

	// Check the type and apply truncation.
	if flags&oDIRECTORY != 0 && node.mode&sIFMT != sIFDIR {
		return negErrno(ENOTDIR)
	}
	if flags&oTRUNC != 0 && node.mode&sIFMT == sIFREG {
		_ = p.truncateNode(node, 0)
	}

	// Install the descriptor, starting an appending file at its end. A
	// synthesized file reads the content it has at the open.
	of := &OpenFile{node: node, path: resolved, flags: flags}
	if node.gen != nil {
		of.data = node.gen()
	}
	if flags&oAPPEND != 0 {
		p.kernel.vfs.refreshHost(node)
		of.offset = node.size
	}
	fd := p.allocFD()
	p.fds[fd] = of
	return uint64(fd)
}

// createFile creates a regular file at a resolved path whose parent
// directory exists. A file in a host-backed directory is created on the host.
func (p *Process) createFile(resolved string, flags, perm uint32) (*VNode, uint64) {
	// Find the parent directory.
	vfs := p.kernel.vfs
	dir, name := splitPath(resolved)
	parent := vfs.lookup(dir)
	if parent == nil || parent.mode&sIFMT != sIFDIR {
		return nil, negErrno(ENOENT)
	}
	if parent.hostPath == "" {
		return vfs.mkFile(parent, name, nil, perm), 0
	}

	// Create the host file.
	hostPath := hostChildPath(parent.hostPath, name)
	osFlags := os.O_CREATE | os.O_RDWR
	if flags&oEXCL != 0 {
		osFlags |= os.O_EXCL
	}
	f, err := vfs.host.OpenFile(hostPath, osFlags, os.FileMode(perm))
	if err != nil {
		return nil, negErrno(EIO)
	}
	_ = f.Close()

	// Add its node.
	node := vfs.mkFile(parent, name, nil, perm)
	node.hostPath = hostPath
	return node, 0
}

// sysClose closes a file descriptor.
func (p *Process) sysClose(fd uint32) uint64 {
	// Drop the descriptor's reference to its open file.
	of, ok := p.fds[fd]
	if !ok {
		return negErrno(EBADF)
	}
	of.unref()
	delete(p.fds, fd)
	return 0
}

// setFD points descriptor fd at of, closing the file fd held before.
func (p *Process) setFD(fd uint32, of *OpenFile) {
	of.ref()
	if old, ok := p.fds[fd]; ok {
		old.unref()
	}
	p.fds[fd] = of
}

// sysFcntl handles file control operations. The kernel has no close-on-exec
// flag, so the descriptor flags read as zero.
func (p *Process) sysFcntl(fd, cmd, arg uint32) uint64 {
	of, ok := p.fds[fd]
	if !ok {
		return negErrno(EBADF)
	}
	switch cmd {
	case fDUPFD, fDUPFD_CLOEXEC:
		newFD := arg
		for p.fds[newFD] != nil {
			newFD++
		}
		p.setFD(newFD, of)
		return uint64(newFD)
	case fGETFD, fSETFD:
		return 0
	case fGETFL:
		return uint64(of.flags)
	case fSETFL:
		of.flags = arg
		return 0
	}
	return negErrno(EINVAL)
}

// sysDup duplicates a file descriptor to the next free number.
func (p *Process) sysDup(fd uint32) uint64 {
	// Share the open file under a new number.
	of, ok := p.fds[fd]
	if !ok {
		return negErrno(EBADF)
	}
	newFD := p.allocFD()
	p.setFD(newFD, of)
	return uint64(newFD)
}

// sysDup2 duplicates a file descriptor to a specific number.
func (p *Process) sysDup2(oldfd, newfd uint32) uint64 {
	of, ok := p.fds[oldfd]
	if !ok {
		return negErrno(EBADF)
	}
	p.setFD(newfd, of)
	return uint64(newfd)
}

// isConsoleFD reports whether fd reads the kernel's console input.
func (p *Process) isConsoleFD(fd uint32) bool {
	of, ok := p.fds[fd]
	return ok && of.node != nil && of.node.isTerminal()
}

// sysPipe2 creates a pipe and writes its read and write descriptors to
// pipefdPtr.
func (p *Process) sysPipe2(pipefdPtr uint32) uint64 {
	// Install both ends.
	pr, pw := newPipeEnds(p.kernel.vfs.allocIno())
	rfd := p.allocFD()
	p.fds[rfd] = &OpenFile{pipe: pr}
	wfd := p.allocFD()
	p.fds[wfd] = &OpenFile{pipe: pw}

	// Return their numbers.
	mem := p.inst.Memory()
	mem.WriteUint32Le(pipefdPtr, rfd)
	mem.WriteUint32Le(pipefdPtr+4, wfd)
	return 0
}

// sysFchdir changes the current working directory to an open directory.
func (p *Process) sysFchdir(fd uint32) uint64 {
	of, ok := p.fds[fd]
	if !ok || of.path == "" {
		return negErrno(EBADF)
	}
	return p.chdirResolved(of.path)
}

// sysFsync is a no-op: the in-memory VFS has nothing to flush, and host
// writes go straight to the host file.
func (p *Process) sysFsync(fd uint32) uint64 {
	if _, ok := p.fds[fd]; !ok {
		return negErrno(EBADF)
	}
	return 0
}
