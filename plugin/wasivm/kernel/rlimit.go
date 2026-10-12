package kernel

// The kernel keeps each process's soft and hard resource limits so that
// getrlimit, setrlimit and prlimit64 return what the guest stored. It does
// not enforce them. NOFILE and STACK have finite defaults because programs
// size their descriptor tables and stacks from them.

// rlimit is one resource limit. rlimInfinity means unbounded.
type rlimit struct {
	// cur is the soft limit.
	cur uint64
	// max is the hard limit, the ceiling for the soft limit.
	max uint64
}

// defaultRlimits returns a new process's resource limits.
func defaultRlimits() [rlimitNlimits]rlimit {
	// Leave every resource unbounded except NOFILE and STACK.
	var r [rlimitNlimits]rlimit
	for i := range r {
		r[i] = rlimit{cur: rlimInfinity, max: rlimInfinity}
	}
	r[rlimitNofile] = rlimit{cur: 1024, max: 4096}
	r[rlimitStack] = rlimit{cur: 8 * 1024 * 1024, max: rlimInfinity}
	return r
}

// sysGetrlimit writes a resource's soft and hard limits to rlimPtr as two
// little-endian 64-bit values.
func (p *Process) sysGetrlimit(resource, rlimPtr uint32) uint64 {
	// Write the limits.
	if resource >= rlimitNlimits {
		return negErrno(EINVAL)
	}
	rl := p.rlimits[resource]
	mem := p.inst.Memory()
	if !mem.WriteUint64Le(rlimPtr, rl.cur) || !mem.WriteUint64Le(rlimPtr+8, rl.max) {
		return negErrno(EFAULT)
	}
	return 0
}

// sysSetrlimit sets a resource's soft and hard limits from rlimPtr. A soft
// limit above the hard limit is invalid.
func (p *Process) sysSetrlimit(resource, rlimPtr uint32) uint64 {
	// Read the limits.
	if resource >= rlimitNlimits {
		return negErrno(EINVAL)
	}
	mem := p.inst.Memory()
	soft, softOK := mem.ReadUint64Le(rlimPtr)
	hard, hardOK := mem.ReadUint64Le(rlimPtr + 8)
	if !softOK || !hardOK {
		return negErrno(EFAULT)
	}

	// Store them.
	if soft > hard {
		return negErrno(EINVAL)
	}
	p.rlimits[resource] = rlimit{cur: soft, max: hard}
	return 0
}
