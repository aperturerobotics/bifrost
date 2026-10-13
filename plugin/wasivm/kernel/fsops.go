package kernel

import (
	"context"
	"os"
)

// sysSendfile copies up to count bytes from descriptor inFd to outFd. When
// offsetPtr is set, the copy reads at the offset stored there and stores the
// new offset back, leaving inFd's offset alone. Otherwise it advances inFd.
func (p *Process) sysSendfile(ctx context.Context, outFd, inFd, offsetPtr, count uint32) uint64 {
	// Find the source file and the offset to read at.
	in, ok := p.fds[inFd]
	if !ok || in.node == nil {
		return negErrno(EBADF)
	}
	if in.node.mode&sIFMT == sIFDIR {
		return negErrno(EINVAL)
	}
	mem := p.inst.Memory()
	off := in.offset
	if offsetPtr != 0 {
		v, ok := mem.ReadUint32Le(offsetPtr)
		if !ok {
			return negErrno(EFAULT)
		}
		off = int64(v)
	}

	// Copy the bytes.
	buf, errno := p.readNodeBytes(in, off, int(count))
	if errno != 0 || len(buf) == 0 {
		return errno
	}
	res := p.writeOutBytes(ctx, outFd, buf)
	if errnoFailed(res) {
		return res
	}

	// Advance the offset the copy read at.
	if offsetPtr != 0 {
		mem.WriteUint32Le(offsetPtr, uint32(off+int64(res)))
	} else {
		in.offset += int64(res)
	}
	return res
}

// sysMknodat creates a regular file, FIFO or device node. A mode without a
// file type creates a regular file.
func (p *Process) sysMknodat(dirfd, pathPtr, mode, dev uint32) uint64 {
	// Find the directory that will hold the node.
	path := readCString(p.inst.Memory(), pathPtr)
	if path == "" {
		return negErrno(ENOENT)
	}
	parent, name, errno := p.newEntry(resolvePath(p.cwd, p.fds, dirfd, path))
	if errno != 0 {
		return errno
	}

	// Create the node.
	switch typ := mode & sIFMT; typ {
	case 0, sIFREG:
		p.kernel.vfs.mkFile(parent, name, nil, mode&0o7777)
	case sIFIFO, sIFCHR, sIFBLK:
		p.kernel.vfs.mkNode(parent, name, typ|mode&0o7777, uint64(dev))
	default:
		return negErrno(EINVAL)
	}
	return 0
}

// sysFlock applies or releases an advisory lock. The kernel does not enforce
// advisory locks, so every valid request succeeds.
func (p *Process) sysFlock(fd, operation uint32) uint64 {
	if _, ok := p.fds[fd]; !ok {
		return negErrno(EBADF)
	}
	switch operation &^ lockNB {
	case lockSH, lockEX, lockUN:
		return 0
	}
	return negErrno(EINVAL)
}

// sysMkdirat creates a directory.
func (p *Process) sysMkdirat(dirfd, pathPtr, mode uint32) uint64 {
	path := readCString(p.inst.Memory(), pathPtr)
	return p.mkdirResolved(resolvePath(p.cwd, p.fds, dirfd, path), mode)
}

// newEntry returns the directory that will hold a new entry at a resolved
// path, with the entry's name. It fails with ENOENT when the directory is
// missing and EEXIST when the entry exists.
func (p *Process) newEntry(resolved string) (*VNode, string, uint64) {
	// Find the directory.
	vfs := p.kernel.vfs
	dir, name := splitPath(resolved)
	parent := vfs.lookup(dir)
	if parent == nil || parent.mode&sIFMT != sIFDIR {
		return nil, "", negErrno(ENOENT)
	}

	// Check the name is free, including on the root filesystem.
	vfs.refreshDirChildren(parent)
	if parent.children[name] != nil {
		return nil, "", negErrno(EEXIST)
	}
	return parent, name, 0
}

// mkdirResolved creates a directory at a resolved path. A directory in a
// host-backed directory is created on the root filesystem.
func (p *Process) mkdirResolved(resolved string, mode uint32) uint64 {
	// Find the parent, and create an in-memory directory under an in-memory
	// parent.
	vfs := p.kernel.vfs
	parent, name, errno := p.newEntry(resolved)
	if errno != 0 {
		return errno
	}
	if parent.hostPath == "" {
		vfs.mkDir(parent, name)
		return 0
	}

	// Create the directory on the root filesystem and load its node.
	hostPath := hostChildPath(parent.hostPath, name)
	if err := vfs.host.MkdirAll(hostPath, os.FileMode(mode&0o7777)); err != nil {
		return negErrno(EIO)
	}
	vfs.child(parent, name)
	return 0
}

// unlinkResolved removes the entry at a resolved path. removeDir removes an
// empty directory, as rmdir does; otherwise the entry must not be a
// directory, as unlink requires.
func (p *Process) unlinkResolved(resolved string, removeDir bool) uint64 {
	// Find the entry.
	vfs := p.kernel.vfs
	dir, name := splitPath(resolved)
	parent := vfs.lookup(dir)
	if parent == nil {
		return negErrno(ENOENT)
	}
	vfs.refreshDirChildren(parent)
	child := parent.children[name]
	if child == nil {
		return negErrno(ENOENT)
	}

	// Check its type against the call.
	isDir := child.mode&sIFMT == sIFDIR
	switch {
	case removeDir && !isDir:
		return negErrno(ENOTDIR)
	case !removeDir && isDir:
		return negErrno(EISDIR)
	case removeDir:
		vfs.refreshDirChildren(child)
		if len(child.children) != 0 {
			return negErrno(ENOTEMPTY)
		}
	}

	// Remove it from the root filesystem and the tree.
	if child.hostPath != "" {
		if err := vfs.host.Remove(child.hostPath); err != nil {
			return negErrno(EIO)
		}
	}
	delete(parent.children, name)
	return 0
}

// sysChdir changes the current working directory.
func (p *Process) sysChdir(pathPtr uint32) uint64 {
	path := readCString(p.inst.Memory(), pathPtr)
	return p.chdirResolved(resolvePath(p.cwd, p.fds, atFDCWD, path))
}

// chdirResolved changes the current working directory to a resolved path.
func (p *Process) chdirResolved(resolved string) uint64 {
	node := p.kernel.vfs.lookup(resolved)
	switch {
	case node == nil:
		return negErrno(ENOENT)
	case node.mode&sIFMT != sIFDIR:
		return negErrno(ENOTDIR)
	}
	p.cwd = resolved
	return 0
}

// renameResolved moves the entry at a resolved path to another, renaming it
// on the root filesystem when both are host-backed.
func (p *Process) renameResolved(oldResolved, newResolved string) uint64 {
	// Find the entry.
	vfs := p.kernel.vfs
	oldDir, oldName := splitPath(oldResolved)
	oldParent := vfs.lookup(oldDir)
	if oldParent == nil {
		return negErrno(ENOENT)
	}
	vfs.refreshDirChildren(oldParent)
	child := oldParent.children[oldName]
	if child == nil {
		return negErrno(ENOENT)
	}

	// Find the destination directory.
	newDir, newName := splitPath(newResolved)
	newParent := vfs.lookup(newDir)
	switch {
	case newParent == nil:
		return negErrno(ENOENT)
	case newParent.mode&sIFMT != sIFDIR:
		return negErrno(ENOTDIR)
	}

	// Move the file on the root filesystem.
	if child.hostPath != "" && newParent.hostPath != "" {
		newHostPath := hostChildPath(newParent.hostPath, newName)
		if err := vfs.host.Rename(child.hostPath, newHostPath); err != nil {
			return negErrno(EIO)
		}
		child.hostPath = newHostPath
	}

	// Move the node in the tree.
	delete(oldParent.children, oldName)
	newParent.children[newName] = child
	child.parent = newParent
	return 0
}

// symlinkResolved creates an in-memory symbolic link to target at a resolved
// path.
func (p *Process) symlinkResolved(target, resolved string) uint64 {
	parent, name, errno := p.newEntry(resolved)
	if errno != 0 {
		return errno
	}
	p.kernel.vfs.mkSymlink(parent, name, target)
	return 0
}

// linkResolved creates a hard link at newResolved to the node at oldResolved.
// follow links the target of a symlink at oldResolved rather than the
// symlink.
func (p *Process) linkResolved(oldResolved, newResolved string, follow bool) uint64 {
	// Find the node to link.
	vfs := p.kernel.vfs
	node := vfs.lookupNoFollow(oldResolved)
	if follow {
		node = vfs.lookup(oldResolved)
	}
	switch {
	case node == nil:
		return negErrno(ENOENT)
	case node.mode&sIFMT == sIFDIR:
		return negErrno(EPERM)
	}

	// Add it under the new name.
	parent, name, errno := p.newEntry(newResolved)
	if errno != 0 {
		return errno
	}
	parent.children[name] = node
	node.nlink++
	return 0
}
