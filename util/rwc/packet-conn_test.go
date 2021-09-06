package rwc

import (
	"context"
	"net"
	"testing"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/pkg/errors"
)

// TestPacketConn tests the packet conn.
func TestPacketConn(t *testing.T) {
	ctx := context.Background()

	peer1, err := peer.NewPeer(nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	peer2, err := peer.NewPeer(nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	a1, a2 := peer.NewNetAddr(peer1.GetPeerID()), peer.NewNetAddr(peer2.GetPeerID())
	c1, c2 := net.Pipe()

	maxPacketSize := uint32(1500)
	pc1 := NewPacketConn(ctx, c1, a1, a2, maxPacketSize, 10)
	pc2 := NewPacketConn(ctx, c2, a2, a1, maxPacketSize, 10)

	data := []byte("testing 1234")
	n, err := pc1.WriteTo(data, pc2.LocalAddr())
	if err == nil && n != len(data) {
		err = errors.Errorf("expected to write %d but wrote %d", len(data), n)
	}
	if err != nil {
		t.Fatal(err.Error())
	}

	outData := make([]byte, len(data)*2)
	on, oa, err := pc2.ReadFrom(outData)
	if err != nil {
		t.Fatal(err.Error())
	}
	outData = outData[:on]
	if on != len(data) {
		t.Fatalf(
			"length incorrect received %v != %v data: %v",
			on,
			len(data),
			string(outData),
		)
	}

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
