package publish

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/packfile/writer"
)

// Repack rewrites every pack of the destination as packs of the blocks of
// live, in its order, and supersedes the old packs once the new ones are
// durable. live must yield the complete closure of every root the destination
// serves: a block it omits is dropped. Pass a writer.WalkAccessOrder walk to
// lay the blocks out in the order readers fetch them.
//
// The destination must have one publisher at a time, and at most
// reclaimMaxReplaced packs, so one push supersedes them all and no reader
// sees a catalog that lacks a live block.
func Repack(ctx context.Context, opts Options, live writer.BlockIterator) error {
	// List the packs to supersede.
	entries, err := FetchPackEntries(ctx, opts.Client, opts.DstSpaceID)
	if err != nil {
		return err
	}
	replaced := make([]string, len(entries))
	for i, entry := range entries {
		replaced[i] = entry.GetId()
	}
	if len(replaced) > reclaimMaxReplaced {
		return errors.Errorf("%d packs exceed the %d one repack supersedes; reclaim first", len(replaced), reclaimMaxReplaced)
	}

	// Write the blocks and supersede the old packs.
	return errors.Wrap(pushSuperseding(ctx, opts, live, replaced), "repack")
}
