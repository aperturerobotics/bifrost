package cdn

import (
	"context"

	"github.com/s4wave/spacewave/core/cdn/v86copy"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/world"
)

// V86ImageCopyProgressFunc receives cumulative block-copy progress.
type V86ImageCopyProgressFunc func(bucket_lookup.ObjectCopyStats) error

// CopyV86ImageFromCdn copies image metadata and referenced assets into a writable World.
func CopyV86ImageFromCdn(ctx context.Context, src, dst world.WorldState, srcObjectKey, dstObjectKey string) error {
	return CopyV86ImageFromCdnWithProgress(ctx, src, dst, srcObjectKey, dstObjectKey, nil)
}

// CopyV86ImageFromCdnWithProgress reports cumulative progress while copying an image.
// GoScript loads the image implementation when the copy is requested.
func CopyV86ImageFromCdnWithProgress(ctx context.Context, src, dst world.WorldState, srcObjectKey, dstObjectKey string, progress V86ImageCopyProgressFunc) error {
	return v86copy.CopyV86ImageFromCdnWithProgress(ctx, src, dst, srcObjectKey, dstObjectKey, progress)
}
