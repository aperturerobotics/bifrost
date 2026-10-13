package s4wave_vm

import (
	"github.com/aperturerobotics/cayley/quad"
)

// WasiVmTypeID is the type identifier for WasiVm objects.
const WasiVmTypeID = "vm/wasivm"

// PredWasiVmRootfs is the graph predicate linking a WasiVm to the UnixFS
// object holding its root filesystem.
var PredWasiVmRootfs = quad.IRI("wasivm/rootfs")

// PredWasiVmImage is the graph predicate linking a WasiVm to the WasiVmImage
// supplying its root filesystem when it has no wasivm/rootfs edge.
var PredWasiVmImage = quad.IRI("wasivm/image")

// DefaultWasiVmInitPath is the init program run when the config names none.
const DefaultWasiVmInitPath = "/bin/sh"

// GetBlockTypeId returns the block type identifier.
func (v *WasiVm) GetBlockTypeId() string {
	return WasiVmTypeID
}

// MarshalBlock marshals the block to binary.
func (v *WasiVm) MarshalBlock() ([]byte, error) {
	return v.MarshalVT()
}

// UnmarshalBlock unmarshals the block from binary.
func (v *WasiVm) UnmarshalBlock(data []byte) error {
	return v.UnmarshalVT(data)
}

// Validate performs cursory checks on the WasiVm.
func (v *WasiVm) Validate() error {
	return nil
}
