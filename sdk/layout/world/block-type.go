package s4wave_layout_world

import "github.com/s4wave/spacewave/db/blocktype"

// ObjectLayoutBlockType decodes layout bodies for typed cursor operations.
var ObjectLayoutBlockType = blocktype.NewBlockType(ObjectLayoutTypeID, NewObjectLayout)
