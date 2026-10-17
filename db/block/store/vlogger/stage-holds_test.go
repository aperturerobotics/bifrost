package block_store_vlogger

import (
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_store "github.com/s4wave/spacewave/db/block/store"
	block_store_test "github.com/s4wave/spacewave/db/block/store/test"
)

// TestVLoggerStageHoldsCompleteRoots checks holds through the logging stage wrapper.
func TestVLoggerStageHoldsCompleteRoots(t *testing.T) {
	block_store_test.RequireStageHold(t, func(inner block.StoreOps) block.StoreOps {
		return NewVLoggerStore(nil, block_store.NewStore("stage", inner))
	})
}
