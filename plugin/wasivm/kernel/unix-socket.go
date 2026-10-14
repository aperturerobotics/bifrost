package kernel

import (
	"bytes"
	"context"
	"encoding/binary"
)

// The wasm32 struct sockaddr_un: a 16-bit family, then the path.
const (
	sockaddrUnSize     = 110
	sockaddrPathOffset = 2
)

// The fields of the wasm32 struct msghdr the kernel reads and writes.
const (
	msghdrNamelen    = 4
	msghdrIov        = 8
	msghdrIovlen     = 12
	msghdrControllen = 20
	msghdrFlags      = 24
)

// shutdown directions, which the guest's SHUT_RD and SHUT_WR share with the
// WASI sdflags.
const (
	shutRD = 1
	shutWR = 2
)

// unixSocket is a Unix stream socket. A connected socket's open file holds
// the connection's two pipes, so the socket keeps only its addresses and a
// listener's queue. The kernel lock guards its fields.
type unixSocket struct {
	// ino is the socket's inode number.
	ino uint64
	// path is the address bind gave the socket, or an accepted socket's
	// listener, or empty.
	path string
	// peerPath is the connected peer's address, or empty.
	peerPath string
	// node is the filesystem node bind created, or nil.
	node *VNode
	// listener is a listening socket's queue, or nil.
	listener *unixListener
}

// vnode returns a node that describes the socket.
func (s *unixSocket) vnode() *VNode {
	return &VNode{ino: s.ino, mode: sIFSOCK | 0o777, nlink: 1}
}

// close stops a listening socket: later connects are refused, and the
// connections it has not accepted close.
func (s *unixSocket) close() {
	if s.listener == nil {
		return
	}
	if s.node != nil && s.node.socket == s {
		s.node.socket = nil
	}
	for _, conn := range s.listener.queue {
		conn.release()
	}
	s.listener.queue = nil
}

// unixListener holds a listening socket's connections that wait for accept.
// The kernel lock guards its fields.
type unixListener struct {
	// queue holds the server ends of the pending connections, oldest first.
	queue []*OpenFile
	// waitCh closes when a connection arrives.
	waitCh chan struct{}
}

// push queues a connection and wakes the waiters.
func (l *unixListener) push(conn *OpenFile) {
	l.queue = append(l.queue, conn)
	close(l.waitCh)
	l.waitCh = make(chan struct{})
}

// state reports whether accept would proceed without waiting, and a channel
// that closes when a connection arrives.
func (l *unixListener) state() (ready bool, change <-chan struct{}) {
	return len(l.queue) != 0, l.waitCh
}

// newSocketFile returns an open file for a new unconnected socket with the
// kernel's status flags.
func (k *Kernel) newSocketFile(status uint32) *OpenFile {
	return &OpenFile{flags: oRDWR | status, sock: &unixSocket{ino: k.vfs.allocIno()}}
}

// connectPair connects two socket files with a pipe in each direction.
func (k *Kernel) connectPair(a, b *OpenFile) {
	toA := newBufferedPipe(k.vfs.allocIno())
	toB := newBufferedPipe(k.vfs.allocIno())
	a.rx, a.tx = toA, toB
	b.rx, b.tx = toB, toA
}

// connectUnix connects the unconnected socket file of to the listener bound
// at the absolute path resolved. The connection waits in the listener's
// queue, so it returns at once. The caller holds the kernel lock.
func (k *Kernel) connectUnix(of *OpenFile, resolved string) uint64 {
	// Find the listener.
	node := k.vfs.lookup(resolved)
	switch {
	case node == nil:
		return negErrno(ENOENT)
	case node.socket == nil:
		return negErrno(ECONNREFUSED)
	}
	server := node.socket

	// Queue the server's end of a new connection.
	conn := &OpenFile{flags: oRDWR, sock: &unixSocket{
		ino:      k.vfs.allocIno(),
		path:     server.path,
		peerPath: of.sock.path,
	}}
	k.connectPair(of, conn)
	of.sock.peerPath = server.path
	server.listener.push(conn)
	return 0
}

// socketFile returns the open file of socket descriptor fd.
func (p *Process) socketFile(fd uint32) (*OpenFile, uint64) {
	of, ok := p.fds[fd]
	switch {
	case !ok:
		return nil, negErrno(EBADF)
	case of.sock == nil:
		return nil, negErrno(ENOTSOCK)
	}
	return of, 0
}

// socketStatus returns the kernel status flags of a socket type, which must
// be a stream socket of the default protocol in the Unix family.
func socketStatus(domain, typ, protocol uint32) (uint32, uint64) {
	switch {
	case domain != guestAFUNIX:
		return 0, negErrno(EAFNOSUPPORT)
	case typ&^(guestSockNonblock|guestSockCloexec) != guestSockStream || protocol != 0:
		return 0, negErrno(EPROTONOSUPPORT)
	case typ&guestSockNonblock != 0:
		return oNONBLOCK, 0
	}
	return 0, 0
}

// sysSocket creates an unconnected Unix stream socket.
func (p *Process) sysSocket(domain, typ, protocol uint32) uint64 {
	status, errno := socketStatus(domain, typ, protocol)
	if errno != 0 {
		return errno
	}
	return uint64(p.installFD(p.kernel.newSocketFile(status)))
}

// sysSocketpair creates two connected Unix stream sockets and writes their
// descriptors to svPtr.
func (p *Process) sysSocketpair(domain, typ, protocol, svPtr uint32) uint64 {
	// Connect two sockets.
	status, errno := socketStatus(domain, typ, protocol)
	if errno != 0 {
		return errno
	}
	a, b := p.kernel.newSocketFile(status), p.kernel.newSocketFile(status)
	p.kernel.connectPair(a, b)

	// Install them and return their numbers.
	afd, bfd := p.installFD(a), p.installFD(b)
	mem := p.inst.Memory()
	if !mem.WriteUint32Le(svPtr, afd) || !mem.WriteUint32Le(svPtr+4, bfd) {
		_ = p.sysClose(afd)
		_ = p.sysClose(bfd)
		return negErrno(EFAULT)
	}
	return 0
}

// readSockaddrPath reads the path of the sockaddr_un of addrLen bytes at
// addrPtr. It returns the path as given and resolved against the working
// directory. The kernel has no abstract namespace: an abstract address, a name
// after a leading null byte, returns abstract set and no path.
func (p *Process) readSockaddrPath(addrPtr, addrLen uint32) (path, resolved string, abstract bool, errno uint64) {
	// Read the address.
	if addrLen <= sockaddrPathOffset || addrLen > sockaddrUnSize {
		return "", "", false, negErrno(EINVAL)
	}
	buf, ok := p.inst.Memory().Read(addrPtr, addrLen)
	switch {
	case !ok:
		return "", "", false, negErrno(EFAULT)
	case binary.LittleEndian.Uint16(buf) != guestAFUNIX:
		return "", "", false, negErrno(EINVAL)
	}

	// Take the path up to its terminator.
	name := buf[sockaddrPathOffset:]
	if len(name) > 1 && name[0] == 0 {
		return "", "", true, 0
	}
	if i := bytes.IndexByte(name, 0); i >= 0 {
		name = name[:i]
	}
	if len(name) == 0 {
		return "", "", false, negErrno(EINVAL)
	}
	path = string(name)
	return path, resolvePath(p.cwd, p.fds, atFDCWD, path), false, 0
}

// writeSockaddr writes a sockaddr_un holding path to addrPtr, truncated to
// the capacity at lenPtr, and its full length to lenPtr. A null addrPtr
// writes nothing.
func (p *Process) writeSockaddr(path string, addrPtr, lenPtr uint32) uint64 {
	// Build the address.
	if addrPtr == 0 {
		return 0
	}
	addr := binary.LittleEndian.AppendUint16(nil, guestAFUNIX)
	if path != "" {
		addr = append(append(addr, path...), 0)
	}

	// Write what fits, and the full length.
	mem := p.inst.Memory()
	capacity, ok := mem.ReadUint32Le(lenPtr)
	if !ok || !mem.Write(addrPtr, addr[:min(int(capacity), len(addr))]) || !mem.WriteUint32Le(lenPtr, uint32(len(addr))) {
		return negErrno(EFAULT)
	}
	return 0
}

// sysBind binds a socket to a path, creating a socket node there.
func (p *Process) sysBind(fd, addrPtr, addrLen uint32) uint64 {
	// Read the address of an unbound socket.
	of, errno := p.socketFile(fd)
	if errno != 0 {
		return errno
	}
	if of.sock.path != "" {
		return negErrno(EINVAL)
	}
	path, resolved, abstract, errno := p.readSockaddrPath(addrPtr, addrLen)
	switch {
	case errno != 0:
		return errno
	case abstract:
		return negErrno(EINVAL)
	}

	// Create the node, which must not exist.
	parent, name, errno := p.newEntry(resolved)
	switch {
	case errno == negErrno(EEXIST):
		return negErrno(EADDRINUSE)
	case errno != 0:
		return errno
	}
	of.sock.node = p.kernel.vfs.mkNode(parent, name, sIFSOCK|0o777&^p.umask, 0)
	of.sock.path = path
	return 0
}

// sysListen makes a bound socket accept connections at its node. The kernel
// queues every connection, so it ignores the backlog.
func (p *Process) sysListen(fd uint32) uint64 {
	of, errno := p.socketFile(fd)
	switch {
	case errno != 0:
		return errno
	case of.sock.node == nil || of.rx != nil:
		return negErrno(EINVAL)
	case of.sock.listener == nil:
		of.sock.listener = &unixListener{waitCh: make(chan struct{})}
		of.sock.node.socket = of.sock
		p.kernel.listens.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
			broadcast()
		})
	}
	return 0
}

// sysConnect connects a socket to the listener bound at a path. An abstract
// address refuses the connection, as Linux does for a name no socket binds, so
// a client such as libxcb falls back to the path.
func (p *Process) sysConnect(fd, addrPtr, addrLen uint32) uint64 {
	// Read the address of an unconnected socket.
	of, errno := p.socketFile(fd)
	switch {
	case errno != 0:
		return errno
	case of.rx != nil:
		return negErrno(EISCONN)
	case of.sock.listener != nil:
		return negErrno(EINVAL)
	}
	_, resolved, abstract, errno := p.readSockaddrPath(addrPtr, addrLen)
	switch {
	case errno != 0:
		return errno
	case abstract:
		return negErrno(ECONNREFUSED)
	}
	return p.kernel.connectUnix(of, resolved)
}

// sysAccept takes the oldest connection waiting on a listening socket and
// returns its descriptor, with the kernel status flags in status. It waits
// for a connection unless the socket is nonblocking.
func (p *Process) sysAccept(ctx context.Context, fd, status uint32) uint64 {
	// Find the listener.
	of, errno := p.socketFile(fd)
	if errno != 0 {
		return errno
	}
	l := of.sock.listener
	if l == nil {
		return negErrno(EINVAL)
	}

	// Wait for a connection.
	ctx, release := p.interruptible(ctx)
	defer release()
	for len(l.queue) == 0 {
		if wait(of.ioContext(ctx), l.waitCh) {
			continue
		}
		if interrupted(ctx) {
			return negErrno(EINTR)
		}
		return negErrno(EAGAIN)
	}

	// Install it.
	conn := l.queue[0]
	l.queue = l.queue[1:]
	conn.flags |= status
	return uint64(p.installFD(conn))
}

// sysSockname writes a socket's address, or its peer's when peer is set, to
// addrPtr and its length to lenPtr.
func (p *Process) sysSockname(fd, addrPtr, lenPtr uint32, peer bool) uint64 {
	// Find the address.
	of, errno := p.socketFile(fd)
	if errno != 0 {
		return errno
	}
	path := of.sock.path
	if peer {
		if of.rx == nil {
			return negErrno(ENOTCONN)
		}
		path = of.sock.peerPath
	}
	return p.writeSockaddr(path, addrPtr, lenPtr)
}

// sysShutdown closes the reading or writing direction of a connected
// socket, or both. A peer reading a closed direction sees end of file.
func (p *Process) sysShutdown(fd, how uint32) uint64 {
	// Check the socket and the directions.
	of, errno := p.socketFile(fd)
	switch {
	case errno != 0:
		return errno
	case how == 0 || how&^(shutRD|shutWR) != 0:
		return negErrno(EINVAL)
	case of.rx == nil:
		return negErrno(ENOTCONN)
	}

	// Close them.
	if how&shutRD != 0 {
		of.rx.close(true)
	}
	if how&shutWR != 0 {
		of.tx.close(false)
	}
	return 0
}

// sysSetsockopt accepts every option on a socket. The kernel's sockets have
// no tunable buffers, timeouts or addresses to reuse.
func (p *Process) sysSetsockopt(fd uint32) uint64 {
	_, errno := p.socketFile(fd)
	return errno
}

// msgContext returns the context of a send or receive with the guest's
// message flags. MSG_DONTWAIT makes it nonblocking; MSG_WAITALL is accepted
// and a short read returns what arrived. Peeking is unsupported.
func msgContext(ctx context.Context, flags uint32) (context.Context, uint64) {
	switch {
	case flags&^(guestMsgDontwait|guestMsgWaitall) != 0:
		return nil, negErrno(EOPNOTSUPP)
	case flags&guestMsgDontwait != 0:
		return nonblocking(ctx), 0
	}
	return ctx, 0
}

// sysSendto writes count bytes at bufPtr to a connected socket. A connected
// stream socket takes no destination address.
func (p *Process) sysSendto(ctx context.Context, fd, bufPtr, count, flags, addrLen uint32) uint64 {
	// Check the socket and the flags.
	_, errno := p.socketFile(fd)
	switch {
	case errno != 0:
		return errno
	case addrLen != 0:
		return negErrno(EISCONN)
	}
	ctx, errno = msgContext(ctx, flags)
	if errno != 0 {
		return errno
	}

	// Write the data.
	return p.sysWrite(ctx, fd, bufPtr, count)
}

// sysRecvfrom reads up to count bytes from a connected socket to bufPtr. A
// stream socket reports no source address.
func (p *Process) sysRecvfrom(ctx context.Context, fd, bufPtr, count, flags, lenPtr uint32) uint64 {
	// Read the data.
	_, errno := p.socketFile(fd)
	if errno != 0 {
		return errno
	}
	ctx, errno = msgContext(ctx, flags)
	if errno != 0 {
		return errno
	}
	res := p.sysRead(ctx, fd, bufPtr, count)

	// Report an empty address.
	if !errnoFailed(res) && lenPtr != 0 && !p.inst.Memory().WriteUint32Le(lenPtr, 0) {
		return negErrno(EFAULT)
	}
	return res
}

// sysMsg sends, when send is set, or receives the data of the msghdr at
// msgPtr through a connected socket. Descriptor passing is unsupported, so
// a send fails with control data and a receive returns none.
func (p *Process) sysMsg(ctx context.Context, fd, msgPtr, flags uint32, send bool) uint64 {
	// Read the header.
	if _, errno := p.socketFile(fd); errno != 0 {
		return errno
	}
	mem := p.inst.Memory()
	iovsPtr, ok1 := mem.ReadUint32Le(msgPtr + msghdrIov)
	iovsLen, ok2 := mem.ReadUint32Le(msgPtr + msghdrIovlen)
	controlLen, ok3 := mem.ReadUint32Le(msgPtr + msghdrControllen)
	switch {
	case !ok1 || !ok2 || !ok3:
		return negErrno(EFAULT)
	case send && controlLen != 0:
		return negErrno(EOPNOTSUPP)
	}
	ctx, errno := msgContext(ctx, flags)
	if errno != 0 {
		return errno
	}

	// Move the data, then report no address, control data or flags.
	res := p.moveIovecs(ctx, fd, iovsPtr, iovsLen, send)
	if send || errnoFailed(res) {
		return res
	}
	for _, off := range []uint32{msghdrNamelen, msghdrControllen, msghdrFlags} {
		if !mem.WriteUint32Le(msgPtr+off, 0) {
			return negErrno(EFAULT)
		}
	}
	return res
}

// moveIovecs writes the iovecs at iovsPtr to fd when send is set, or reads
// fd into them, as writev and readv do.
func (p *Process) moveIovecs(ctx context.Context, fd, iovsPtr, iovsLen uint32, send bool) uint64 {
	total, errno := wasiMoveIovecs(p.inst.Memory(), iovsPtr, iovsLen, func(base, length uint32) (uint32, uint32) {
		if send {
			return wasiResult(p.sysWrite(ctx, fd, base, length))
		}
		return wasiResult(p.sysRead(ctx, fd, base, length))
	})
	if errno != wasiErrnoSuccess {
		return negErrno(linuxErrno(errno))
	}
	return uint64(total)
}
