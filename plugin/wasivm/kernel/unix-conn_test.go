package kernel

import (
	"context"
	"errors"
	"io"
	"testing"
)

// TestDialUnix checks host code connects to a guest listener, both directions
// carry data, and each side reads end of file after the other closes.
func TestDialUnix(t *testing.T) {
	// A path without a listener refuses the host.
	p, mem := fsTestProcess()
	if _, err := p.kernel.DialUnix("/tmp/sock"); !errors.Is(err, ErrNoListener) {
		t.Fatalf("dial before listen = %v, want ErrNoListener", err)
	}

	// Connect the host and accept it in the guest.
	listener := listenTestSocket(t, p, mem, "/tmp/sock", guestSockStream)
	conn, err := p.kernel.DialUnix("/tmp/sock")
	if err != nil {
		t.Fatal(err)
	}
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
	conn, err = p.kernel.DialUnix("/tmp/sock")
	if err != nil {
		t.Fatal(err)
	}
	server = p.sysAccept(context.Background(), listener, 0)
	_ = conn.Close()
	if res, _ := recvTest(p, mem, uint32(server)); res != 0 {
		t.Fatalf("guest read after the host closed = %#x, want 0", res)
	}
}
