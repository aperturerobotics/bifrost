package world_block

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/world"
	world_types "github.com/s4wave/spacewave/db/world/types"
)

// WalkBlocks visits the persisted World metadata and each current object's full
// block graph. resolve supplies the object's root decoder; unknown types fail
// instead of treating an opaque root as a complete object. The callback is
// sequential and can copy each block to durable storage. External bucket and
// transform references are resolved by the World storage.
// Like other WorldState operations, the caller serializes access to the state.
func (t *WorldState) WalkBlocks(ctx context.Context, resolve func(context.Context, string) (block.Ctor, error), visit func(*block.BlockRef, []byte) error) error {
	seen := make(map[string]struct{})
	walk := func(ref *block.BlockRef, ctor block.Ctor, store block.StoreOps, xfrm block.Transformer) error {
		if ref.GetEmpty() {
			return nil
		}
		return bucket_lookup.WalkObjectBlocks(ctx, bucket_lookup.NewWalkObjectBlocksWithRef(ref, ctor), func(entry *bucket_lookup.WalkObjectBlocksEntry) (bool, error) {
			if entry.Err != nil {
				return false, entry.Err
			}
			if entry.IsSubBlock || entry.Ref.GetEmpty() {
				return true, nil
			}
			if !entry.Found {
				return false, errors.Wrap(block.ErrNotFound, entry.Ref.MarshalString())
			}
			// An opaque metadata reference can later be visited with its object
			// decoder. Deduplicate writes without skipping that traversal.
			key := entry.Ref.MarshalString()
			if _, found := seen[key]; !found {
				if err := visit(entry.Ref, entry.Data); err != nil {
					return false, err
				}
				seen[key] = struct{}{}
			}
			return true, nil
		}, store, xfrm, 1, false)
	}
	if err := walk(t.GetRootRef(), NewWorldBlock, t.store, t.xfrm); err != nil {
		return err
	}

	objects := t.IterateObjects(ctx, "", false)
	defer objects.Close()
	for objects.Next() {
		object, found, err := t.GetObject(ctx, objects.Key())
		if err != nil {
			return err
		}
		if !found {
			return world.ErrObjectNotFound
		}
		err = func() error {
			defer world.ReleaseObjectState(object)
			ref, _, err := object.GetRootRef(ctx)
			if err != nil || ref.GetRootRef().GetEmpty() {
				return err
			}
			typeID, err := world_types.GetObjectType(ctx, t, object.GetKey())
			if err != nil {
				return err
			}
			ctor, err := resolve(ctx, typeID)
			if err != nil {
				return err
			}
			if ctor == nil {
				return errors.Errorf("block decoder unavailable for object %s (%s)", object.GetKey(), typeID)
			}
			return object.AccessWorldState(ctx, nil, func(cursor *bucket_lookup.Cursor) error {
				return walk(ref.GetRootRef(), ctor, cursor.GetBucket(), cursor.GetTransformer())
			})
		}()
		if err != nil {
			return err
		}
	}
	return objects.Err()
}
