package wasivm_plugin

import (
	"context"
	"testing"
	"time"

	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// TestDownloadsLocalRootfs checks that a run whose root filesystem lives in
// the Space reports one empty snapshot and nothing after it.
func TestDownloadsLocalRootfs(t *testing.T) {
	// Watch until a short deadline, keeping every snapshot sent.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	var sent []*s4wave_vm.WasiVmDownloads
	err := newDownloads(nil).watch(ctx, func(d *s4wave_vm.WasiVmDownloads) error {
		sent = append(sent, d)
		return nil
	})

	// The watch sends the empty snapshot once and then only waits.
	if err != context.DeadlineExceeded {
		t.Fatalf("watch returned %v, want the context's deadline", err)
	}
	if len(sent) != 1 || sent[0].GetRemote() || sent[0].GetBytes() != 0 {
		t.Fatalf("watch sent %v, want one empty snapshot", sent)
	}
}
