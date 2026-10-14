package kernel

import "encoding/binary"

// The ioctl requests the kernel serves, in the Linux values the C library
// passes through.
const (
	tiocsctty  = 0x540e
	tiocgpgrp  = 0x540f
	tiocspgrp  = 0x5410
	tiocgwinsz = 0x5413
	tiocswinsz = 0x5414
	fionread   = 0x541b
	fionbio    = 0x5421
	tiocnotty  = 0x5422
	tiocgsid   = 0x5429
	tiocgptn   = 0x80045430
	tiocsptlck = 0x40045431
)

// sysIoctl serves the ioctl requests on descriptor fd. arg is the request's
// guest pointer, or its integer for TIOCSCTTY. A request the kernel does not
// serve fails with ENOTTY.
func (p *Process) sysIoctl(fd, req, arg uint32) uint64 {
	// Serve the requests of any file.
	of, ok := p.fds[fd]
	if !ok {
		return negErrno(EBADF)
	}
	mem := p.inst.Memory()
	switch req {
	case fionread:
		return p.putInt(arg, uint32(of.pending()))
	case fionbio:
		on, ok := mem.ReadUint32Le(arg)
		if !ok {
			return negErrno(EFAULT)
		}
		of.flags &^= oNONBLOCK
		if on != 0 {
			of.flags |= oNONBLOCK
		}
		return 0
	case tiocgptn, tiocsptlck:
		return p.ioctlMaster(of, req, arg)
	}

	// Serve the terminal requests.
	t := of.terminal()
	if t == nil {
		return negErrno(ENOTTY)
	}
	switch req {
	case tiocgwinsz:
		cols, rows := t.size()
		var ws [8]byte
		binary.LittleEndian.PutUint16(ws[0:], uint16(rows))
		binary.LittleEndian.PutUint16(ws[2:], uint16(cols))
		if !mem.Write(arg, ws[:]) {
			return negErrno(EFAULT)
		}
		return 0
	case tiocswinsz:
		ws, ok := mem.Read(arg, 4)
		if !ok {
			return negErrno(EFAULT)
		}
		t.resize(uint32(binary.LittleEndian.Uint16(ws[2:])), uint32(binary.LittleEndian.Uint16(ws[0:])))
		return 0
	case tiocsctty:
		return p.setControlling(t)
	case tiocnotty:
		if p.controllingFD(fd) == nil {
			return negErrno(ENOTTY)
		}
		if p.ctty != nil && p.sid == p.pid {
			p.ctty.sid, p.ctty.fgPgid = 0, 0
		}
		p.ctty = nil
		return 0
	case tiocgpgrp:
		res := p.sysTcgetpgrp(fd)
		if errnoFailed(res) {
			return res
		}
		return p.putInt(arg, uint32(res))
	case tiocspgrp:
		pgid, ok := mem.ReadUint32Le(arg)
		if !ok {
			return negErrno(EFAULT)
		}
		return p.sysTcsetpgrp(fd, int32(pgid))
	case tiocgsid:
		if p.controllingFD(fd) == nil {
			return negErrno(ENOTTY)
		}
		return p.putInt(arg, uint32(p.sid))
	}
	return negErrno(ENOTTY)
}

// ioctlMaster serves the requests of a pty master: TIOCGPTN reads the pty's
// index, and TIOCSPTLCK locks or unlocks its slave.
func (p *Process) ioctlMaster(of *OpenFile, req, arg uint32) uint64 {
	// Read the index.
	pt := of.ptm
	if pt == nil {
		return negErrno(ENOTTY)
	}
	if req == tiocgptn {
		return p.putInt(arg, uint32(pt.index))
	}

	// Set the lock.
	lock, ok := p.inst.Memory().ReadUint32Le(arg)
	if !ok {
		return negErrno(EFAULT)
	}
	pt.locked = lock != 0
	return 0
}

// setControlling makes t the controlling terminal of p, a session leader,
// for TIOCSCTTY. A pty another session controls stays with it.
func (p *Process) setControlling(t *tty) uint64 {
	switch {
	case p.sid != p.pid:
		return negErrno(EPERM)
	case t.pty == nil:
		p.ctty = nil
		return 0
	case t.sid != 0 && t.sid != p.sid:
		return negErrno(EPERM)
	}
	p.control(t)
	return 0
}

// control makes t the controlling terminal of p, a session leader, with p's
// group in the foreground.
func (p *Process) control(t *tty) {
	p.ctty = t
	t.sid, t.fgPgid = p.sid, p.pgid
}

// putInt writes v to the guest int at ptr.
func (p *Process) putInt(ptr, v uint32) uint64 {
	if !p.inst.Memory().WriteUint32Le(ptr, v) {
		return negErrno(EFAULT)
	}
	return 0
}

// pending returns the number of bytes a read of the file could take now.
func (of *OpenFile) pending() int {
	switch {
	case of.rx != nil:
		return of.rx.buffered()
	case of.tty != nil:
		return of.tty.in.pending()
	case of.node != nil && of.node.mode&sIFMT == sIFREG:
		return int(max(of.node.size-of.offset, 0))
	}
	return 0
}

// fdMaster returns the pty whose master is descriptor fd. It fails with
// EBADF for a closed descriptor and EINVAL for any other file.
func (p *Process) fdMaster(fd uint32) (*pty, uint64) {
	of, ok := p.fds[fd]
	switch {
	case !ok:
		return nil, negErrno(EBADF)
	case of.ptm == nil:
		return nil, negErrno(EINVAL)
	}
	return of.ptm, 0
}

// ttyname writes the path of the terminal on fd, null-terminated, to the
// size bytes at buf, as ttyname_r does. It returns 0 or a WASI errno.
func (p *Process) ttyname(fd, buf, size uint32) uint32 {
	of, ok := p.fds[fd]
	switch {
	case !ok:
		return wasiErrno(EBADF)
	case of.terminal() == nil:
		return wasiErrno(ENOTTY)
	case uint32(len(of.path)) >= size:
		return wasiErrno(ERANGE)
	}
	if !p.inst.Memory().Write(buf, append([]byte(of.path), 0)) {
		return wasiErrno(EFAULT)
	}
	return 0
}
