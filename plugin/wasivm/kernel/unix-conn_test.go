package kernel

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
)

// TestDialUnix checks a host dial waits for a guest listener, both directions
// carry data, and each side reads end of file after the other closes.
func TestDialUnix(t *testing.T) {
	// A dial that ends before a listener binds the path fails.
	p, mem := fsTestProcess()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.kernel.DialUnix(ctx, "/tmp/sock"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled dial = %v, want context.Canceled", err)
	}

	// Dial before the guest listens.
	type dialResult struct {
		conn net.Conn
		err  error
	}
	dialed := make(chan dialResult, 1)
	go func() {
		conn, err := p.kernel.DialUnix(context.Background(), "/tmp/sock")
		dialed <- dialResult{conn, err}
	}()

	// Listen under the kernel lock, as a host call does, and accept the
	// connection the dial queued.
	p.kernel.mu.Lock()
	listener := listenTestSocket(t, p, mem, "/tmp/sock", guestSockStream)
	p.kernel.mu.Unlock()
	res := <-dialed
	if res.err != nil {
		t.Fatal(res.err)
	}
	conn := res.conn
	server := p.sysAccept(context.Background(), listener, 0)
	if errnoFailed(server) {
		t.Fatalf("accept = %#x", server)
	}

	// Send each way.
	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	if _, got := recvTest(p, mem, uint32(server)); got != "ping" {
		t.Fatalf("guest read %q, want ping", got)
	}
	sendTest(t, p, mem, uint32(server), "pong")
	buf := make([]byte, 64)
	if n, err := conn.Read(buf); err != nil || string(buf[:n]) != "pong" {
		t.Fatalf("host read %q, %v, want pong", buf[:n], err)
	}

	// Closing the guest's end ends the host's reads, and closing the host's
	// ends the guest's.
	_ = p.sysClose(uint32(server))
	if _, err := conn.Read(buf); err != io.EOF {
		t.Fatalf("host read after the guest closed = %v, want EOF", err)
	}
	conn, err := p.kernel.DialUnix(context.Background(), "/tmp/sock")
	if err != nil {
		t.Fatal(err)
	}
	server = p.sysAccept(context.Background(), listener, 0)
	_ = conn.Close()
	if res, _ := recvTest(p, mem, uint32(server)); res != 0 {
		t.Fatalf("guest read after the host closed = %#x, want 0", res)
	}
}

// TestListenUnix checks a host listener accepts a guest's connection, takes
// the files the guest sends and sends them back, and that its path can be
// bound again after it closes.
func TestListenUnix(t *testing.T) {
	// Listen in a directory the guest lacks, and connect from the guest.
	p, mem := fsTestProcess()
	l, err := p.kernel.ListenUnix("/run/wayland-0")
	if err != nil {
		t.Fatal(err)
	}
	client := uint32(p.sysSocket(guestAFUNIX, guestSockStream, 0))
	if res := p.sysConnect(client, sockTestAddr, writeTestSockaddr(mem, "/run/wayland-0")); res != 0 {
		t.Fatalf("connect = %#x", res)
	}
	conn, err := l.AcceptUnix()
	if err != nil {
		t.Fatal(err)
	}

	// The host receives a pipe's write end and closes the guest's.
	r, w := pipeTest(t, p, mem)
	sendRightsTest(t, p, mem, client, "x", w)
	_ = p.sysClose(w)
	buf := make([]byte, 64)
	n, files, err := conn.ReadMsg(buf)
	if err != nil || string(buf[:n]) != "x" || len(files) != 1 {
		t.Fatalf("host ReadMsg = %q with %d files, %v, want x with 1", buf[:n], len(files), err)
	}

	// The host sends it back and closes its own, and the guest's copy
	// writes the pipe.
	if _, err := conn.WriteMsg([]byte("y"), files); err != nil {
		t.Fatal(err)
	}
	_ = files[0].Close()
	data, fds, _ := recvRightsTest(t, p, mem, client, 64)
	if data != "y" || len(fds) != 1 {
		t.Fatalf("guest recvmsg = %q with %v, want y with 1", data, fds)
	}
	sendTest(t, p, mem, fds[0], "hi")
	if _, got := recvTest(p, mem, r); got != "hi" {
		t.Fatalf("pipe read %q, want hi", got)
	}

	// A closed listener refuses connects and its path binds again.
	_ = l.Close()
	if _, err := l.AcceptUnix(); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("accept after close = %v, want net.ErrClosed", err)
	}
	other := uint32(p.sysSocket(guestAFUNIX, guestSockStream, 0))
	if res := p.sysConnect(other, sockTestAddr, writeTestSockaddr(mem, "/run/wayland-0")); res != negErrno(ECONNREFUSED) {
		t.Fatalf("connect to a closed listener = %#x, want ECONNREFUSED", res)
	}
	l, err = p.kernel.ListenUnix("/run/wayland-0")
	if err != nil {
		t.Fatalf("listen again = %v", err)
	}
	_ = l.Close()
}
