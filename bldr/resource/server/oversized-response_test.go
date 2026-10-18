package resource_server_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aperturerobotics/starpc/echo"
	"github.com/aperturerobotics/starpc/srpc"
)

// TestResourceRPCOversizedResponseReportsError checks that a response over the
// transport's message limit fails the call with the size error, rather than
// closing the route before the handler's error completion reaches the caller.
func TestResourceRPCOversizedResponseReportsError(t *testing.T) {
	// Answer with a message over starpc's 10 MB packet limit.
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	rootMux := srpc.InvokerFunc(func(serviceID, methodID string, strm srpc.Stream) (bool, error) {
		if err := strm.MsgRecv(&echo.EchoMsg{}); err != nil {
			return true, err
		}
		return true, strm.MsgSend(&echo.EchoMsg{Body: strings.Repeat("x", 10_000_001)})
	})

	// Call the handler over a real packet transport.
	client, _, rootRPC := newPendingResourceTestClient(t, ctx, rootMux, nil)
	defer client.Release()
	_, err := echo.NewSRPCEchoerClient(rootRPC).Echo(ctx, &echo.EchoMsg{})

	// The caller sees the size error the handler returned.
	if err == nil || !strings.Contains(err.Error(), srpc.ErrMessageTooLarge.Error()) {
		t.Fatalf("Echo error = %v, want %v", err, srpc.ErrMessageTooLarge)
	}
}
