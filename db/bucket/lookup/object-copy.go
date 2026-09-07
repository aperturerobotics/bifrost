package bucket_lookup

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	block_transform "github.com/s4wave/spacewave/db/block/transform"
	"github.com/s4wave/spacewave/db/bucket"
	trace "github.com/s4wave/spacewave/db/traceutil"
)

// CopyObjectToBucket copies an object from srcCursor to destCursor.
//
// rootCtor must construct the block located at srcCursor.
//
// The concurrency limit controls how many concurrent read/writes can be called.
// If maxConcurrency <= 0, has no limit on concurrent read/writes.
//
// Copies use the source transform and return the destination object ref with
// its bucket ID and transform configuration set explicitly.
//
// skipSubtreeExists skips a block ref tree if the block already existed in the
// target storage. This assumes that a block existing in storage implies that
// all blocks it references have also already been stored.
//
// cb is an optional callback to call with each block before copying.
// If cb is nil and a block is not found, returns block.ErrNotFound.
func CopyObjectToBucket(
	ctx context.Context,
	destCursor, srcCursor *Cursor,
	rootCtor block.Ctor,
	maxConcurrency int,
	skipSubtreeExists bool,
	cb WalkObjectBlocksCb,
) (*bucket.ObjectRef, error) {
	ref, _, err := CopyObjectToBucketWithStats(
		ctx,
		destCursor,
		srcCursor,
		rootCtor,
		maxConcurrency,
		skipSubtreeExists,
		cb,
	)
	return ref, err
}

// CopyObjectToBucketWithStats copies an object and returns its logical copy accounting.
func CopyObjectToBucketWithStats(
	ctx context.Context,
	destCursor, srcCursor *Cursor,
	rootCtor block.Ctor,
	maxConcurrency int,
	skipSubtreeExists bool,
	cb WalkObjectBlocksCb,
) (*bucket.ObjectRef, ObjectCopyStats, error) {
	return copyObjectToBucket(
		ctx,
		destCursor,
		srcCursor,
		rootCtor,
		maxConcurrency,
		skipSubtreeExists,
		cb,
		nil,
	)
}

// CopyObjectToBucketWithProgress copies an object and reports logical copy
// accounting after each processed source block.
func CopyObjectToBucketWithProgress(
	ctx context.Context,
	destCursor, srcCursor *Cursor,
	rootCtor block.Ctor,
	maxConcurrency int,
	skipSubtreeExists bool,
	cb WalkObjectBlocksCb,
	progress ObjectCopyProgress,
) (*bucket.ObjectRef, ObjectCopyStats, error) {
	return copyObjectToBucket(
		ctx,
		destCursor,
		srcCursor,
		rootCtor,
		maxConcurrency,
		skipSubtreeExists,
		cb,
		progress,
	)
}

// copyObjectToBucket copies each distinct source block and traverses its children once.
func copyObjectToBucket(
	ctx context.Context,
	destCursor, srcCursor *Cursor,
	rootCtor block.Ctor,
	maxConcurrency int,
	skipSubtreeExists bool,
	cb WalkObjectBlocksCb,
	progress ObjectCopyProgress,
) (*bucket.ObjectRef, ObjectCopyStats, error) {
	// Preserve the source encoding in the destination reference.
	srcRef := srcCursor.GetRef()
	destinationRef := srcRef.Clone()
	destinationRef.BucketId = destCursor.GetOpArgs().GetBucketId()
	destinationRef.TransformConf = srcCursor.GetTransformConf().Clone()
	destinationRef.TransformConfRef = nil

	// An empty source transform means the copied blocks are raw at rest. When
	// the destination parent is transformed, make that raw encoding explicit
	// in the returned ref so nested references reset the inherited decoder.
	if srcCursor.GetTransformConf().GetEmpty() && !destCursor.GetTransformConf().GetEmpty() {
		destBucket, ok := destCursor.GetBucket().(bucket.Bucket)
		if !ok {
			return nil, ObjectCopyStats{}, errors.New("destination bucket does not support transform config writes")
		}
		transformConfRef, _, err := WriteTransformConf(
			ctx,
			destBucket,
			destCursor.GetTransformer(),
			destCursor.defaultPutOpts(),
			&block_transform.Config{},
		)
		if err != nil {
			return nil, ObjectCopyStats{}, errors.Wrap(err, "write raw transform config")
		}
		destinationRef.TransformConf = nil
		destinationRef.TransformConfRef = transformConfRef
	}

	// Cursors in the same bucket and volume already share storage.
	if srcCursor.GetOpArgs().EqualVT(destCursor.GetOpArgs()) {
		return destinationRef, ObjectCopyStats{}, nil
	}

	writeCursor, err := destCursor.FollowRef(ctx, destinationRef)
	if err != nil {
		if err == context.Canceled {
			return nil, ObjectCopyStats{}, err
		}
		return nil, ObjectCopyStats{}, errors.Wrap(err, "construct write cursor")
	}
	defer writeCursor.Release()

	readBkt := srcCursor.GetBucket()
	readXfrm := srcCursor.GetTransformer()
	if readXfrm == nil {
		readXfrm = block_transform.NewTransformerWithSteps(nil)
	}
	writeBkt := writeCursor.GetBucket()

	// seenMtx guards the set of block references claimed by copy workers.
	var seenMtx sync.Mutex
	seenBlocks := make(map[string]struct{})
	var seenBlockCount atomic.Int64
	var copiedBlocks atomic.Int64
	var dedupedBlocks atomic.Int64
	var existingBlocks atomic.Int64
	var writtenBlocks atomic.Int64
	var skippedSubtrees atomic.Int64

	var logicalSourceBytes atomic.Int64
	var progressMtx sync.Mutex
	snapshot := func() ObjectCopyStats {
		return ObjectCopyStats{
			BlocksSeen:         seenBlockCount.Load(),
			BlocksCopied:       copiedBlocks.Load(),
			BlocksExisting:     existingBlocks.Load(),
			BlocksWritten:      writtenBlocks.Load(),
			BlocksDeduped:      dedupedBlocks.Load(),
			SubtreesSkipped:    skippedSubtrees.Load(),
			LogicalSourceBytes: logicalSourceBytes.Load(),
		}
	}
	reportProgress := func() error {
		if progress == nil {
			return nil
		}
		progressMtx.Lock()
		err := progress(snapshot())
		progressMtx.Unlock()
		return err
	}

	// GetBlockRefCtor supplies the decoder for each child in the block graph.
	// TODO: handle garbage collection (set parent in PutOpts).
	ctx = withCopyWorkerTrace(ctx)
	err = WalkObjectBlocks(
		ctx,
		NewWalkObjectBlocksWithRef(srcRef.GetRootRef(), rootCtor),
		func(ent *WalkObjectBlocksEntry) (bool, error) {
			var cntu bool
			var err error

			// The callback may recover a read error.
			if cb != nil {
				cntu, err = cb(ent)
			} else {
				err = ent.Err
				if err == nil && !ent.Found && !ent.IsSubBlock && !ent.Ref.GetEmpty() {
					err = errors.Wrap(block.ErrNotFound, ent.Ref.MarshalString())
				}
				cntu = err == nil
			}

			if err != nil || ent.IsSubBlock || !ent.Found || ent.Ref.GetEmpty() || len(ent.Data) == 0 {
				// Inline sub-blocks are traversed without a separate storage write.
				return cntu, err
			}

			// Only the first worker traverses a shared block's descendants.
			refStr := ent.Ref.MarshalString()
			seenMtx.Lock()
			_, seen := seenBlocks[refStr]
			seenBlocks[refStr] = struct{}{}
			seenMtx.Unlock()
			if seen {
				dedupedBlocks.Add(1)
				if err := reportProgress(); err != nil {
					return false, err
				}
				return false, nil
			}

			seenBlockCount.Add(1)
			logicalSourceBytes.Add(int64(len(ent.Data)))
			// Most implementations check existence inside PutBlock.
			var writeRef *block.BlockRef
			var writeExisted bool
			writeRef, writeExisted, err = writeBkt.PutBlock(ctx, ent.Data, &block.PutOpts{
				HashType:      ent.Ref.GetHash().GetHashType(),
				ForceBlockRef: ent.Ref,
			})
			if err == nil && !writeRef.EqualsRef(ent.Ref) {
				err = errors.Errorf("wrote to different ref %s", writeRef.MarshalString())
			}
			if err != nil && err != context.Canceled {
				err = errors.Wrapf(err, "write ref %s", ent.Ref.MarshalString())
			}
			if err == nil {
				copiedBlocks.Add(1)
				if writeExisted {
					existingBlocks.Add(1)
				} else {
					writtenBlocks.Add(1)
				}
			}

			if skipSubtreeExists && writeExisted && err == nil {
				skippedSubtrees.Add(1)
				if err := reportProgress(); err != nil {
					return false, err
				}
				return false, nil
			}

			if progressErr := reportProgress(); progressErr != nil {
				return false, progressErr
			}
			return err == nil, err
		},
		readBkt, readXfrm,
		maxConcurrency,
		false,
	)
	stats := snapshot()
	trace.Logf(ctx, "copy-block-seen-count", "%d", stats.BlocksSeen)
	trace.Logf(ctx, "copy-block-copied-count", "%d", stats.BlocksCopied)
	trace.Logf(ctx, "copy-block-dedupe-skip-count", "%d", stats.BlocksDeduped)
	trace.Logf(ctx, "copy-block-existing-count", "%d", stats.BlocksExisting)
	trace.Logf(ctx, "copy-block-written-count", "%d", stats.BlocksWritten)
	trace.Logf(ctx, "copy-block-skip-subtree-count", "%d", stats.SubtreesSkipped)
	trace.Logf(ctx, "copy-block-logical-source-byte-count", "%d", stats.LogicalSourceBytes)
	trace.Logf(ctx, "copy-block-destination-durable-byte-count", "%d", stats.DestinationDurableBytes)
	trace.Logf(ctx, "copy-block-demand-read-count", "%d", stats.DemandReadCount)
	trace.Logf(ctx, "copy-block-demand-read-byte-count", "%d", stats.DemandReadBytes)
	if err != nil {
		return nil, stats, err
	}

	return destinationRef, stats, nil
}
