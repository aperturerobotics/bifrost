//go:build !js

package wasivm_plugin

import (
	"testing"
	"time"

	"github.com/s4wave/spacewave/db/world"
	db_world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// TestResolveRootfsThroughImage checks that a VM with no rootfs edge of its
// own boots the root filesystem of its linked image.
func TestResolveRootfsThroughImage(t *testing.T) {
	// Start a World testbed.
	ctx := t.Context()
	wtb, err := db_world_testbed.Default(ctx, db_world_testbed.WithWorldVerbose(false))
	if err != nil {
		t.Fatal(err)
	}
	defer wtb.Release()
	ws := world.NewEngineWorldState(wtb.Engine, true)

	// Create an image on a rootfs and a VM on the image.
	const rootfsKey, imageKey, vmKey = "fs/rootfs", "vm-image/test", "vm/wasivm/test"
	writeRootfsFile(t, ws, rootfsKey, "init", []byte("init"))
	createImage := s4wave_vm.NewCreateWasiVmImageOp(imageKey, &s4wave_vm.WasiVmImage{Name: "test"}, time.Unix(1, 0))
	if _, _, err := ws.ApplyWorldOp(ctx, createImage, ""); err != nil {
		t.Fatal(err)
	}
	edge := world.NewGraphQuadWithKeys(imageKey, string(s4wave_vm.PredWasiVmImageRootfs), rootfsKey, "")
	if err := ws.SetGraphQuad(ctx, edge); err != nil {
		t.Fatal(err)
	}
	createVm := s4wave_vm.NewCreateWasiVmOp(vmKey, "test", "", imageKey, time.Unix(1, 0))
	if _, _, err := ws.ApplyWorldOp(ctx, createVm, ""); err != nil {
		t.Fatal(err)
	}

	// The VM resolves the image's rootfs in the same World.
	src, err := resolveRootfs(ctx, wtb.Logger, nil, ws, vmKey)
	if err != nil {
		t.Fatal(err)
	}
	defer src.release()
	if src.key != rootfsKey || src.ws != ws {
		t.Fatalf("rootfs = %q, want %q in the VM's World", src.key, rootfsKey)
	}
}
