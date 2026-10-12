package kernel

import "time"

// VNode is a file, directory, symlink or device in the VFS. A host-backed
// node mirrors the file at hostPath in the root filesystem, which holds its
// content; a host-backed directory loads its entries on demand.
type VNode struct {
	// ino is the inode number.
	ino uint64
	// mode holds the file type and permission bits.
	mode  uint32
	nlink uint32
	// rdev is the device number of a device node.
	rdev uint64
	size int64
	// data is an in-memory regular file's content.
	data []byte
	// children holds a directory's entries by name.
	children map[string]*VNode
	// xattrs holds the extended attributes by full name.
	xattrs map[string][]byte
	parent *VNode
	// symlink is a symlink's target.
	symlink string
	// hostPath is the root filesystem path of a host-backed node, or empty.
	hostPath string
	atime    time.Time
	mtime    time.Time
	ctime    time.Time
}

// isTerminal reports whether node is the terminal: /dev/tty or /dev/console.
func (node *VNode) isTerminal() bool {
	return node.mode&sIFMT == sIFCHR && (node.rdev == devTTY || node.rdev == devConsole)
}
