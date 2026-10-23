package volume_rpc_client

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/s4wave/spacewave/db/coord"
	"github.com/s4wave/spacewave/db/coord/conformance"
	volume_rpc "github.com/s4wave/spacewave/db/volume/rpc"
)

func TestCoordinatorLeaseDetectedLossContract(t *testing.T) {
	lease := &lease{
		cancel: func() {},
		done:   make(chan struct{}),
	}
	capability := &coord.Capability{Supported: true, DetectsLoss: true}
	conformance.CheckDetectedLoss(t, capability, lease, func() {
		lease.watchStream(func() error { return io.EOF })
	})
	if !errors.Is(lease.Err(), io.EOF) {
		t.Fatalf("lease error = %v, want EOF", lease.Err())
	}
}

// originRecordingClient records the origin on lease calls and fails each call.
type originRecordingClient struct {
	volume_rpc.SRPCProxyVolumeClient
	origins []string
}

func (c *originRecordingClient) TryAcquireCoordinatorWriteLease(ctx context.Context, _ *volume_rpc.TryAcquireCoordinatorWriteLeaseRequest) (volume_rpc.SRPCProxyVolume_TryAcquireCoordinatorWriteLeaseClient, error) {
	c.origins = append(c.origins, directive.OriginFromContext(ctx).ControllerID)
	return nil, io.EOF
}

func (c *originRecordingClient) WaitAcquireCoordinatorWriteLease(ctx context.Context, _ *volume_rpc.WaitAcquireCoordinatorWriteLeaseRequest) (volume_rpc.SRPCProxyVolume_WaitAcquireCoordinatorWriteLeaseClient, error) {
	c.origins = append(c.origins, directive.OriginFromContext(ctx).ControllerID)
	return nil, io.EOF
}

func (c *originRecordingClient) ReleaseCoordinatorWriteLease(ctx context.Context, _ *volume_rpc.CoordinatorWriteLeaseRequest) (*volume_rpc.ReleaseCoordinatorWriteLeaseResponse, error) {
	c.origins = append(c.origins, directive.OriginFromContext(ctx).ControllerID)
	return &volume_rpc.ReleaseCoordinatorWriteLeaseResponse{}, nil
}

// TestCoordinatorLeaseCallsKeepOrigin checks that the lease streams and the
// release request, which outlive the caller's cancellation, keep its origin.
func TestCoordinatorLeaseCallsKeepOrigin(t *testing.T) {
	// Run each lease call under an origin controller.
	const originID = "test/lease-origin"
	ctx := directive.WithOriginController(t.Context(), originID)
	client := &originRecordingClient{}
	c := NewCoordinator(client)
	_, _, _ = c.TryAcquireWriteLease(ctx, coord.Scope{})
	_, _ = c.WaitAcquireWriteLease(ctx, coord.Scope{})
	l := &lease{client: client, cancel: func() {}, done: make(chan struct{})}
	if err := l.Release(ctx); err != nil {
		t.Fatal(err)
	}

	// Check each call kept the origin.
	if len(client.origins) != 3 {
		t.Fatalf("recorded %d lease calls, want 3", len(client.origins))
	}
	for i, got := range client.origins {
		if got != originID {
			t.Fatalf("lease call %d origin = %q, want %q", i, got, originID)
		}
	}
}
