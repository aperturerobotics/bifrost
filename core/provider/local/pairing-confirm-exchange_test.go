package provider_local

import (
	"context"
	"net"
	"testing"
	"time"

	stream_packet "github.com/s4wave/spacewave/net/stream/packet"
)

// TestPairingApprovalCancellation closes a blocked duplex exchange when the
// user leaves pairing, including while the other client has not decided.
func TestPairingApprovalCancellation(t *testing.T) {
	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	sess := stream_packet.NewSession(left, 1024)
	status, err := exchangePairingApproval(ctx, sess, make(chan bool), "operation", func(PairingStatus) {})
	if err == nil || status != PairingStatusConfirmationTimeout {
		t.Fatalf("canceled approval returned status %v, error %v", status, err)
	}
	_ = right.SetWriteDeadline(time.Now().Add(time.Second))
	if _, err := right.Write([]byte{0}); err == nil {
		t.Fatal("canceled approval left its peer stream open")
	}
}
