package kernel

// The kernel runs every process as root, so every credential change is
// permitted. The credential syscalls track the real, effective, saved and
// filesystem ids and the supplementary groups so that a get returns what a
// set stored. File permissions are not checked against them.

// sysSetuid sets the real, effective, saved and filesystem uid.
func (p *Process) sysSetuid(uid uint32) uint64 {
	p.uid, p.euid, p.suid, p.fsuid = uid, uid, uid, uid
	return 0
}

// sysSetgid sets the real, effective, saved and filesystem gid.
func (p *Process) sysSetgid(gid uint32) uint64 {
	p.gid, p.egid, p.sgid, p.fsgid = gid, gid, gid, gid
	return 0
}

// sysGetgroups writes up to size supplementary group ids to listPtr and
// returns their count. A size of zero returns the count alone.
func (p *Process) sysGetgroups(size, listPtr uint32) uint64 {
	// Check the list has room.
	n := uint32(len(p.groups))
	if size == 0 {
		return uint64(n)
	}
	if size < n {
		return negErrno(EINVAL)
	}

	// Write the groups.
	mem := p.inst.Memory()
	for i, g := range p.groups {
		if !mem.WriteUint32Le(listPtr+uint32(i)*4, g) {
			return negErrno(EFAULT)
		}
	}
	return uint64(n)
}
