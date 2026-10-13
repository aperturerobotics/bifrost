package kernel

// utsFieldLen is the size of each struct utsname field, with its NUL.
const utsFieldLen = 65

// sysUname writes the system's names to a struct utsname at ptr: sysname,
// nodename, release, version, machine and domainname, each in a field of
// utsFieldLen bytes.
func (p *Process) sysUname(ptr uint32) uint64 {
	// Lay out the fields, each NUL-padded.
	k := p.kernel
	names := []string{"Linux", k.hostname, utsRelease, utsVersion, "wasm32", k.domainname}
	buf := make([]byte, len(names)*utsFieldLen)
	for i, name := range names {
		copy(buf[i*utsFieldLen:(i+1)*utsFieldLen-1], name)
	}

	// Write them.
	if !p.inst.Memory().Write(ptr, buf) {
		return negErrno(EFAULT)
	}
	return 0
}

// utsRelease and utsVersion are the kernel release and version uname
// reports.
const (
	utsRelease = "6.12.0-wasivm"
	utsVersion = "#1 SMP"
)

// sysSetUTSName sets *name to the size bytes at ptr, as sethostname and
// setdomainname do.
func (p *Process) sysSetUTSName(name *string, ptr, size uint32) uint64 {
	// The name must fit its utsname field with the NUL.
	if size >= utsFieldLen {
		return negErrno(EINVAL)
	}
	data, ok := p.inst.Memory().Read(ptr, size)
	if !ok {
		return negErrno(EFAULT)
	}
	*name = string(data)
	return 0
}
