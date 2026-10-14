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
