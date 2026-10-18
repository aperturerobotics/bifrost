package kernel

import "slices"

// The mount flags the kernel reads.
const (
	msRdonly  = 1
	msRemount = 32
)

// mount is a filesystem mounted over a directory, which it hides until it
// is unmounted.
type mount struct {
	source string
	target string
	fstype string
	// flags holds the MS_ flags, of which the table shows MS_RDONLY.
	flags uint32
	// data holds the filesystem's options.
	data string
	// root is the root directory of the mounted filesystem.
	root *VNode
	// covered is the directory the mount hides, or nil for the root.
	covered *VNode
}

// options returns the mount's options as /proc/mounts lists them.
func (m *mount) options() string {
	opts := "rw"
	if m.flags&msRdonly != 0 {
		opts = "ro"
	}
	if m.data != "" {
		opts += "," + m.data
	}
	return opts
}

// newFS creates the root of a new filesystem of fstype, or returns nil for a
// type the kernel lacks. The in-memory types start empty.
func (k *Kernel) newFS(fstype string) *VNode {
	switch fstype {
	case "proc":
		return k.newProcFS()
	case "sysfs":
		return k.newSysFS()
	case "tmpfs", "ramfs", "devpts", "mqueue":
		return k.vfs.newDir(0o755)
	}
	return nil
}

// mountFS mounts m, whose root is set, over the directory dir.
func (k *Kernel) mountFS(dir *VNode, m *mount) {
	m.covered = dir
	link(dir.parent, nameIn(dir.parent, dir), m.root)
	k.mounts = append(k.mounts, m)
}

// sysMount mounts a new filesystem of fstype over the directory target, or
// with MS_REMOUNT changes the flags and options of the filesystem mounted
// there.
func (p *Process) sysMount(source, target, fstype string, flags uint32, data string) uint64 {
	// Find the directory.
	k := p.kernel
	resolved := resolvePath(p.cwd, p.fds, atFDCWD, target)
	dir := k.vfs.lookup(resolved)
	switch {
	case dir == nil:
		return negErrno(ENOENT)
	case dir.mode&sIFMT != sIFDIR:
		return negErrno(ENOTDIR)
	}

	// Change a mounted filesystem's flags.
	if flags&msRemount != 0 {
		m := k.mountOf(dir)
		if m == nil {
			return negErrno(EINVAL)
		}
		m.flags, m.data = flags&^msRemount, data
		return 0
	}

	// Mount a new filesystem.
	if dir == k.vfs.root {
		return negErrno(EBUSY)
	}
	root := k.newFS(fstype)
	if root == nil {
		return negErrno(ENODEV)
	}
	k.mountFS(dir, &mount{source: source, target: resolved, fstype: fstype, flags: flags, data: data, root: root})
	return 0
}

// sysUmount2 unmounts the filesystem mounted at target, uncovering the
// directory it hid.
func (p *Process) sysUmount2(target string) uint64 {
	// Find the mount.
	k := p.kernel
	dir := k.vfs.lookup(resolvePath(p.cwd, p.fds, atFDCWD, target))
	if dir == nil {
		return negErrno(ENOENT)
	}
	m := k.mountOf(dir)
	switch {
	case m == nil:
		return negErrno(EINVAL)
	case m.covered == nil:
		return negErrno(EBUSY)
	}

	// Put the covered directory back.
	link(dir.parent, nameIn(dir.parent, dir), m.covered)
	k.mounts = slices.DeleteFunc(k.mounts, func(o *mount) bool { return o == m })
	return 0
}

// mountOf returns the mount whose root is dir, or nil.
func (k *Kernel) mountOf(dir *VNode) *mount {
	for _, m := range slices.Backward(k.mounts) {
		if m.root == dir {
			return m
		}
	}
	return nil
}
