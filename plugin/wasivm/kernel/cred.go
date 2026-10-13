package kernel

// The kernel runs every process as root, so every credential change is
// permitted. The credential calls track the real, effective, saved and
// filesystem ids and the supplementary groups so that a get returns what a
// set stored. File permissions are not checked against them.

// ids are the real, effective, saved and filesystem ids of one kind, user or
// group.
type ids struct {
	real, eff, saved, fs uint32
}

// idKeep is the id argument -1, which leaves an id unchanged.
const idKeep = ^uint32(0)

// set sets all four ids, as setuid does for root.
func (s *ids) set(id uint32) {
	*s = ids{real: id, eff: id, saved: id, fs: id}
}

// setRes sets the real, effective and saved ids, as setresuid does, keeping
// each one passed as idKeep. The filesystem id follows the effective id.
func (s *ids) setRes(r, e, saved uint32) {
	for _, id := range []struct {
		field *uint32
		to    uint32
	}{{&s.real, r}, {&s.eff, e}, {&s.saved, saved}} {
		if id.to != idKeep {
			*id.field = id.to
		}
	}
	s.fs = s.eff
}

// setRe sets the real and effective ids, as setreuid does: the saved id
// takes the new effective id when the real id is set or the effective id
// changes away from the real one.
func (s *ids) setRe(r, e uint32) {
	saved := idKeep
	if r != idKeep || (e != idKeep && e != s.real) {
		saved = e
		if e == idKeep {
			saved = s.eff
		}
	}
	s.setRes(r, e, saved)
}

// writeRes writes the real, effective and saved ids to their pointers, as
// getresuid does.
func (p *Process) writeRes(s ids, rPtr, ePtr, sPtr uint32) uint64 {
	mem := p.inst.Memory()
	if !mem.WriteUint32Le(rPtr, s.real) || !mem.WriteUint32Le(ePtr, s.eff) || !mem.WriteUint32Le(sPtr, s.saved) {
		return negErrno(EFAULT)
	}
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

// sysSetgroups sets the supplementary groups to the size ids at listPtr.
func (p *Process) sysSetgroups(size, listPtr uint32) uint64 {
	// Linux allows up to NGROUPS_MAX groups.
	if size > ngroupsMax {
		return negErrno(EINVAL)
	}

	// Read the groups.
	mem := p.inst.Memory()
	groups := make([]uint32, size)
	for i := range groups {
		g, ok := mem.ReadUint32Le(listPtr + uint32(i)*4)
		if !ok {
			return negErrno(EFAULT)
		}
		groups[i] = g
	}
	p.groups = groups
	return 0
}

// ngroupsMax is Linux's limit on supplementary groups.
const ngroupsMax = 65536
