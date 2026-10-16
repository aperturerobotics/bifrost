package space_world

import (
	"context"

	"github.com/s4wave/spacewave/db/blocktype"
	unixfs_block "github.com/s4wave/spacewave/db/unixfs/block"
	unixfs_world "github.com/s4wave/spacewave/db/unixfs/world"
	s4wave_layout_world "github.com/s4wave/spacewave/sdk/layout/world"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// LookupBlockType looks up a core-owned block type by ID.
// Returns nil if not found. VM images resolve here because the CDN catalog
// Space that publishes them is mounted without plugins. Plugin-owned block
// types (such as the SQL cursor blocks owned by the sql plugin) resolve
// through their plugin's LookupBlockType directive handler, not here.
func LookupBlockType(
	ctx context.Context,
	typeID string,
) (blocktype.BlockType, error) {
	switch typeID {
	case SpaceSettingsBlockType.GetBlockTypeID(), "github.com/s4wave/spacewave/core/space/world.SpaceSettings":
		return SpaceSettingsBlockType, nil
	case s4wave_layout_world.ObjectLayoutTypeID:
		return s4wave_layout_world.ObjectLayoutBlockType, nil
	case s4wave_vm.V86ImageTypeID:
		return s4wave_vm.V86ImageBlockType, nil
	case s4wave_vm.WasiVmImageTypeID:
		return s4wave_vm.WasiVmImageBlockType, nil
	case unixfs_world.FSNodeTypeID:
		return blocktype.New[unixfs_block.FSNode](typeID), nil
	case unixfs_world.FSObjectTypeID:
		return blocktype.New[unixfs_block.FSObject](typeID), nil
	case unixfs_world.FSHostVolumeTypeID:
		return blocktype.New[unixfs_block.FSHostVolume](typeID), nil
	default:
		return lookupApplicationBlockType(ctx, typeID)
	}
}
