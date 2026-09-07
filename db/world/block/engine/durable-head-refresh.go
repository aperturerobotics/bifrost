package world_block_engine

import (
	"context"

	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/object"
)

// refreshDurableHeadRef reads the persisted World head for engine transactions
// and the engine-owned coordinator watch. The caller retains stateStore until
// Engine.Close has joined all refreshes.
func (c *Controller) refreshDurableHeadRef(stateStore object.ObjectStore) func(context.Context) (*bucket.ObjectRef, error) {
	return func(ctx context.Context) (*bucket.ObjectRef, error) {
		headState, found, err := c.loadHeadState(ctx, stateStore)
		if err != nil || !found {
			return nil, err
		}
		return headState.GetHeadRef(), nil
	}
}
