package kernel

// sysSetpriority sets the calling process's nice value. The kernel stores it
// for getpriority and does not schedule by it.
func (p *Process) sysSetpriority(which, who, prio uint32) uint64 {
	// Accept only the calling process as the target.
	if which > prioUser {
		return negErrno(EINVAL)
	}
	if who != 0 && who != uint32(p.pid) {
		return negErrno(ESRCH)
	}

	// Store the nice value.
	nice := int32(prio)
	if nice < -20 || nice > 19 {
		return negErrno(EINVAL)
	}
	p.nice = nice
	return 0
}

// sysGetpriority returns the calling process's nice value.
func (p *Process) sysGetpriority(which, who uint32) uint64 {
	if which > prioUser {
		return negErrno(EINVAL)
	}
	if who != 0 && who != uint32(p.pid) {
		return negErrno(ESRCH)
	}
	return uint64(uint32(p.nice))
}
