//go:build !js && !tinygo

package sobject_world_engine

import (
	"context"
	"runtime/trace"

	"github.com/s4wave/spacewave/core/bstore"
	"github.com/s4wave/spacewave/db/block"
)

// replayTimingBlockStore traces reads through the real testbed block store.
type replayTimingBlockStore struct {
	// BlockStore holds the encrypted World blocks.
	bstore.BlockStore
}

// GetBlock traces the storage read, excluding the cursor's transform decode.
func (s *replayTimingBlockStore) GetBlock(ctx context.Context, ref *block.BlockRef) ([]byte, bool, error) {
	ctx, task := trace.NewTask(ctx, "test/replay/block-read")
	defer task.End()
	return s.BlockStore.GetBlock(ctx, ref)
}

var _ bstore.BlockStore = (*replayTimingBlockStore)(nil)
