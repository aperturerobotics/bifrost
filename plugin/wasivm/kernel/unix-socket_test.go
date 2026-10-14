package kernel

import (
	"context"
	"encoding/binary"
	"testing"
	"time"
)

// Guest memory addresses the socket tests use.
const (
	sockTestAddr = 1024
	sockTestLen  = 1200
	sockTestBuf  = 2048
	sockTestPair = 3072
)

// writeTestSockaddr writes a sockaddr_un holding path to sockTestAddr and
// returns its length.
func writeTestSockaddr(mem *memMemory, path string) uint32 {
	addr := binary.LittleEndian.AppendUint16(nil, guestAFUNIX)
	addr = append(append(addr, path...), 0)
	mem.Write(sockTestAddr, addr)
	return uint32(len(addr))
}

// listenTestSocket binds a socket to path, listens on it and returns its
// descriptor.
func listenTestSocket(t *testing.T, p *Process, mem *memMemory, path string, typ uint32) uint32 {
	// Create, bind and listen.
	t.Helper()
	fd := uint32(p.sysSocket(guestAFUNIX, typ, 0))
	if res := p.sysBind(fd, sockTestAddr, writeTestSockaddr(mem, path)); res != 0 {
		t.Fatalf("bind(%s) = %#x", path, res)
	}
	if res := p.sysListen(fd); res != 0 {
		t.Fatalf("listen = %#x", res)
	}
	return fd
}

// sendTest writes data to fd and checks the whole write took.
func sendTest(t *testing.T, p *Process, mem *memMemory, fd uint32, data string) {
	t.Helper()
	mem.Write(sockTestBuf, []byte(data))
	if res := p.sysWrite(context.Background(), fd, sockTestBuf, uint32(len(data))); res != uint64(len(data)) {
		t.Fatalf("write(%d, %q) = %#x", fd, data, res)
	}
}

// recvTest reads up to 64 bytes from fd and returns the result and the
// bytes read.
func recvTest(p *Process, mem *memMemory, fd uint32) (uint64, string) {
	res := p.sysRead(context.Background(), fd, sockTestBuf, 64)
	if errnoFailed(res) {
		return res, ""
	}
	buf, _ := mem.Read(sockTestBuf, uint32(res))
	return res, string(buf)
}

// TestUnixSocketConnectAccept checks a client connects to a bound listener,
// both directions carry data, the names report the bound path, and closing
// one end gives the other end of file.
func TestUnixSocketConnectAccept(t *testing.T) {
	// Connect a client to a listener and accept it.
	p, mem := fsTestProcess()
	listener := listenTestSocket(t, p, mem, "/tmp/sock", guestSockStream)
	client := uint32(p.sysSocket(guestAFUNIX, guestSockStream, 0))
	if res := p.sysConnect(client, sockTestAddr, writeTestSockaddr(mem, "/tmp/sock")); res != 0 {
		t.Fatalf("connect = %#x", res)
	}
	server := p.sysAccept(context.Background(), listener, 0)
	if errnoFailed(server) {
		t.Fatalf("accept = %#x", server)
	}

	// Send each way.
	sendTest(t, p, mem, client, "ping")
	if _, got := recvTest(p, mem, uint32(server)); got != "ping" {
		t.Fatalf("server read %q, want ping", got)
	}
	sendTest(t, p, mem, uint32(server), "pong")
	if _, got := recvTest(p, mem, client); got != "pong" {
		t.Fatalf("client read %q, want pong", got)
	}

	// The client's peer is the listener's path.
	mem.WriteUint32Le(sockTestLen, sockaddrUnSize)
	if res := p.sysSockname(client, sockTestAddr, sockTestLen, true); res != 0 {
		t.Fatalf("getpeername = %#x", res)
	}
	n, _ := mem.ReadUint32Le(sockTestLen)
	addr, _ := mem.Read(sockTestAddr, n)
	if string(addr[sockaddrPathOffset:]) != "/tmp/sock\x00" {
		t.Fatalf("peer name %q, want /tmp/sock", addr[sockaddrPathOffset:])
	}

	// Closing the server's end ends the client's reads.
	_ = p.sysClose(uint32(server))
	if res, _ := recvTest(p, mem, client); res != 0 {
		t.Fatalf("read after the peer closed = %#x, want 0", res)
	}
}

// TestUnixSocketAddressErrors checks a bound path refuses a second bind, and
// a path whose listener closed refuses connections.
func TestUnixSocketAddressErrors(t *testing.T) {
	// Bind the path twice.
	p, mem := fsTestProcess()
	listener := listenTestSocket(t, p, mem, "/tmp/sock", guestSockStream)
	other := uint32(p.sysSocket(guestAFUNIX, guestSockStream, 0))
	if res := p.sysBind(other, sockTestAddr, writeTestSockaddr(mem, "/tmp/sock")); res != negErrno(EADDRINUSE) {
		t.Fatalf("second bind = %#x, want EADDRINUSE", res)
	}

	// Connect after the listener closed.
	_ = p.sysClose(listener)
	if res := p.sysConnect(other, sockTestAddr, writeTestSockaddr(mem, "/tmp/sock")); res != negErrno(ECONNREFUSED) {
		t.Fatalf("connect to a closed listener = %#x, want ECONNREFUSED", res)
	}
	if res := p.sysSocket(guestAFUNIX+1, guestSockStream, 0); res != negErrno(EAFNOSUPPORT) {
		t.Fatalf("socket in another family = %#x, want EAFNOSUPPORT", res)
	}
}

// TestUnixSocketNonblocking checks a nonblocking accept and read return
// EAGAIN instead of waiting, and that F_GETFL reports the guest's flags.
func TestUnixSocketNonblocking(t *testing.T) {
	// Accept with nothing queued.
	p, mem := fsTestProcess()
	listener := listenTestSocket(t, p, mem, "/tmp/sock", guestSockStream|guestSockNonblock)
	if res := p.sysAccept(context.Background(), listener, 0); res != negErrno(EAGAIN) {
		t.Fatalf("nonblocking accept = %#x, want EAGAIN", res)
	}

	// Read an empty socket made nonblocking by F_SETFL.
	if res := p.sysSocketpair(guestAFUNIX, guestSockStream, 0, sockTestPair); res != 0 {
		t.Fatalf("socketpair = %#x", res)
	}
	a, _ := mem.ReadUint32Le(sockTestPair)
	if res := p.sysFcntl(a, fSETFL, guestONONBLOCK); res != 0 {
		t.Fatalf("F_SETFL = %#x", res)
	}
	if res, _ := recvTest(p, mem, a); res != negErrno(EAGAIN) {
		t.Fatalf("nonblocking read = %#x, want EAGAIN", res)
	}
	if got, want := p.sysFcntl(a, fGETFL, 0), uint64(guestORDONLY|guestOWRONLY|guestONONBLOCK); got != want {
		t.Fatalf("F_GETFL = %#x, want %#x", got, want)
	}
}

// writeTestSubscription writes a poll_oneoff descriptor subscription at index
// i of the array at inPtr.
func writeTestSubscription(mem *memMemory, inPtr, i uint32, userdata uint64, eventType uint8, fd uint32) {
	// Clear the subscription, then fill its header.
	sub := inPtr + i*wasiSubscriptionSize
	mem.Write(sub, make([]byte, wasiSubscriptionSize))
	mem.WriteUint64Le(sub, userdata)
	mem.WriteByteAt(sub+8, eventType)
	mem.WriteUint32Le(sub+16, fd)
}

// TestPollOneoffWaitsForReadiness checks poll_oneoff reports a writable
// socket at once, waits for data on an empty one, and reports a hangup.
func TestPollOneoffWaitsForReadiness(t *testing.T) {
	// Connect a pair and subscribe to both directions of one end.
	const inPtr, outPtr, neventsPtr = 1024, 1536, 2000
	p, mem := fsTestProcess()
	if res := p.sysSocketpair(guestAFUNIX, guestSockStream, 0, sockTestPair); res != 0 {
		t.Fatalf("socketpair = %#x", res)
	}
	a, _ := mem.ReadUint32Le(sockTestPair)
	b, _ := mem.ReadUint32Le(sockTestPair + 4)
	writeTestSubscription(mem, inPtr, 0, 1, wasiEventtypeFDRead, a)
	writeTestSubscription(mem, inPtr, 1, 2, wasiEventtypeFDWrite, a)
	ctx := context.Background()

	// Only the write is ready.
	if errno := p.wasiPollOneoff(ctx, mem, inPtr, outPtr, 2, neventsPtr); errno != wasiErrnoSuccess {
		t.Fatalf("poll_oneoff = %d", errno)
	}
	if n, _ := mem.ReadUint32Le(neventsPtr); n != 1 {
		t.Fatalf("poll_oneoff reported %d events, want 1", n)
	}
	if userdata, _ := mem.ReadUint64Le(outPtr); userdata != 2 {
		t.Fatalf("poll_oneoff reported userdata %d, want 2", userdata)
	}

	// A read waits until the peer writes.
	peer := p.fds[b].tx
	go func() {
		time.Sleep(10 * time.Millisecond)
		_, _ = peer.write(ctx, []byte("x"))
	}()
	if errno := p.wasiPollOneoff(ctx, mem, inPtr, outPtr, 1, neventsPtr); errno != wasiErrnoSuccess {
		t.Fatalf("poll_oneoff for data = %d", errno)
	}
	if userdata, _ := mem.ReadUint64Le(outPtr); userdata != 1 {
		t.Fatalf("poll_oneoff for data reported userdata %d, want 1", userdata)
	}

	// After the peer closes, the read reports a hangup.
	_ = p.sysClose(b)
	if errno := p.wasiPollOneoff(ctx, mem, inPtr, outPtr, 1, neventsPtr); errno != wasiErrnoSuccess {
		t.Fatalf("poll_oneoff after close = %d", errno)
	}
	if flags, _ := mem.Read(outPtr+24, 2); binary.LittleEndian.Uint16(flags) != wasiEventrwflagsHangup {
		t.Fatalf("poll_oneoff after close reported flags %v, want hangup", flags)
	}
}
