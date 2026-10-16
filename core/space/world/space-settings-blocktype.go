package space_world

import "github.com/s4wave/spacewave/db/blocktype"

// SpaceSettingsBlockType is the BlockType for SpaceSettings.
var SpaceSettingsBlockType = blocktype.New[SpaceSettings](
	"space/settings",
)
