//go:build tinygo

package space_world

import (
	"context"

	"github.com/s4wave/spacewave/db/blocktype"
)

// lookupApplicationBlockType leaves plugin-owned decoders to their controllers.
func lookupApplicationBlockType(_ context.Context, _ string) (blocktype.BlockType, error) {
	return nil, nil
}
