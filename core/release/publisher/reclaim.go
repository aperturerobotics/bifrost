package publisher

import (
	"context"

	"github.com/pkg/errors"
	cdn_publish "github.com/s4wave/spacewave/core/cdn/publish"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/packfile"
	"github.com/s4wave/spacewave/net/hash"
)

// reclaim rewrites the destination's packs that are at least half dead once
// the checkpoint whose World is head has been posted. blocks is the complete
// closure of head: every ref of a decoded block is in blocks. The pass stops
// when the destination head moves away from head. Returns the dead value
// bytes dropped.
func reclaim(ctx context.Context, opts cdn_publish.Options, head *bucket.ObjectRef, blocks []packedBlock) (uint64, error) {
	// Index the live blocks and require the closure to hold every stored ref
	// of a decoded block, since a missed subtree would read as dead.
	live := make([]*hash.Hash, len(blocks))
	stored := make(map[string]*block.StoredBlock, len(blocks))
	for i, entry := range blocks {
		live[i] = entry.ref.GetHash()
		stored[string(packfile.BlockKey(entry.ref.GetHash()))] = entry.stored
	}
	for _, entry := range blocks {
		if !entry.decoded {
			continue
		}
		for _, ref := range entry.stored.GetRefs() {
			if ref.GetEmpty() {
				continue
			}
			if _, ok := stored[string(packfile.BlockKey(ref.GetHash()))]; !ok {
				return 0, errors.Errorf("release closure lacks %s referenced by %s", ref.MarshalString(), entry.ref.MarshalString())
			}
		}
	}

	// The checkpoint holds head without its bucket binding.
	want := head.CloneVT()
	want.BucketId = ""

	// Replace the mostly dead packs while the destination head stays at head.
	read := func(_ context.Context, h *hash.Hash) (*block.StoredBlock, error) {
		return stored[string(packfile.BlockKey(h))], nil
	}
	check := func(ctx context.Context) error {
		current, err := cdn_publish.FetchSourceHeadRef(ctx, opts.Client, opts.DstSpaceID)
		if err != nil {
			return errors.Wrap(err, "read destination head")
		}
		if !current.EqualVT(want) {
			return errors.New("destination head moved during reclaim")
		}
		return nil
	}
	return cdn_publish.Reclaim(ctx, opts, live, read, check)
}
