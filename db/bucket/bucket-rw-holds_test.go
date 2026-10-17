package bucket

import (
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_store_test "github.com/s4wave/spacewave/db/block/store/test"
)

// TestBucketRWStageHoldsCompleteRoots checks holds reach the write stage, not the read store.
func TestBucketRWStageHoldsCompleteRoots(t *testing.T) {
	block_store_test.RequireStageHold(t, func(inner block.StoreOps) block.StoreOps {
		return NewBucketRW(&bucketRW{conf: &Config{Id: "stage"}}, inner)
	})
}
