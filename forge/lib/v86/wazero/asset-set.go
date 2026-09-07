package v86_wazero

// AssetSet is the materialized v86 boot image used by the wazero harness.
type AssetSet struct {
	// Dir is the base directory containing the boot files.
	Dir string
	// ImageKey identifies the source V86Image object.
	ImageKey string
	// Wasm is the emulator module path.
	Wasm string
	// SeaBIOS is the machine BIOS path.
	SeaBIOS string
	// VGABIOS is the display BIOS path.
	VGABIOS string
	// Kernel is the guest kernel path.
	Kernel string
	// RootfsTar is the local guest filesystem archive path.
	RootfsTar string
	// RootfsJSON is the v86 filesystem index path.
	RootfsJSON string
	// RootfsFlatDir contains files addressed by the v86 filesystem index.
	RootfsFlatDir string
	// RootfsObjectKey identifies a CDN filesystem when no local archive is used.
	RootfsObjectKey string
}
