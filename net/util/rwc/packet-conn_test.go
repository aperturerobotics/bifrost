package rwc

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/net/peer"
)

// TestPacketConn tests the packet conn.
func TestPacketConn(t *testing.T) {
	// Prepare the connection test context.
	ctx := context.Background()

	// Create the first peer for the local connection address.
	peer1, err := peer.NewPeer(nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create the second peer for the remote connection address.
	peer2, err := peer.NewPeer(nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Connect the peer addresses with an in-memory stream pair.
	a1, a2 := peer.NewNetAddr(peer1.GetPeerID()), peer.NewNetAddr(peer2.GetPeerID())
	c1, c2 := net.Pipe()

	// Wrap both stream ends with bounded packet connections.
	maxPacketSize := uint32(1500)
	pc1 := NewPacketConn(ctx, c1, a1, a2, maxPacketSize, 10)
	pc2 := NewPacketConn(ctx, c2, a2, a1, maxPacketSize, 10)

	// Send the test payload and require a complete write.
	data := []byte("testing 1234")
	n, err := pc1.WriteTo(data, pc2.LocalAddr())
	if err == nil && n != len(data) {
		err = errors.Errorf("expected to write %d but wrote %d", len(data), n)
	}
	if err != nil {
		t.Fatal(err.Error())
	}

	// Receive the payload through the remote connection.
	outData := make([]byte, len(data)*2)
	on, oa, err := pc2.ReadFrom(outData)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Verify that the received payload has the expected length.
	outData = outData[:on]
	if on != len(data) {
		t.Fatalf(
			"length incorrect received %v != %v data: %v",
			on,
			len(data),
			string(outData),
		)
	}

	// Verify that the packet retains the sending peer address.
	outAddr := oa.String()
	expectedAddr := a1.String()
	if outAddr != expectedAddr {
		t.Fatalf(
			"expected remote addr %s but got %s",
			expectedAddr,
			outAddr,
		)
	}
}

// tempReadErr is a temporary net.Error like the dtls deadline error pion sctp
// leaves on a closed data channel.
type tempReadErr struct{}

func (tempReadErr) Error() string   { return "deadline exceeded" }
func (tempReadErr) Timeout() bool   { return true }
func (tempReadErr) Temporary() bool { return true }

// tempErrRwc fails every read with tempReadErr.
type tempErrRwc struct{ io.ReadWriteCloser }

func (tempErrRwc) Read([]byte) (int, error) { return 0, tempReadErr{} }

// TestRwcPacketConnReadErrNotTemporary checks that a read error is not a
// net.Error, so quic-go closes the transport instead of retrying forever.
func TestRwcPacketConnReadErrNotTemporary(t *testing.T) {
	// Read once from a conn whose reader fails with a temporary error.
	pc := NewRwcPacketConn(tempErrRwc{}, nil, nil)
	_, _, err := pc.ReadFrom(make([]byte, 16))

	// Require a wrapped error that is no longer a net.Error.
	if err == nil {
		t.Fatal("expected read error")
	}
	if _, ok := err.(net.Error); ok {
		t.Fatalf("read error %v is a net.Error", err)
	}
	if !errors.Is(err, tempReadErr{}) {
		t.Fatalf("read error %v does not wrap the cause", err)
	}
}
