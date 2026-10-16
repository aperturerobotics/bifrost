package bldr_manifest_world

import (
	"context"

	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// ExExtractManifestBundleOp extracts a manifest bundle with an operation.
func ExExtractManifestBundleOp(
	ctx context.Context,
	ws world.WorldState,
	sender peer.ID,
	objectKey string,
	linkObjKeys []string,
	manifestBundleRef *bucket.ObjectRef,
) error {
	op := NewExtractManifestBundleOp(
		objectKey,
		linkObjKeys,
		manifestBundleRef,
	)
	_, _, err := ws.ApplyWorldOp(ctx, op, sender)
	return err
}

// NewExtractManifestBundleOp constructs a new ExtractManifestBundleOp block.
func NewExtractManifestBundleOp(
	objectKey string,
	linkObjectKeys []string,
	manifestBundleRef *bucket.ObjectRef,
) *ExtractManifestBundleOp {
	return &ExtractManifestBundleOp{
		ObjectKey:      objectKey,
		LinkObjectKeys: linkObjectKeys,
		ManifestBundle: manifestBundleRef,
	}
}

// Validate performs cursory checks on the op.
func (o *ExtractManifestBundleOp) Validate() error {
	if len(o.GetObjectKey()) == 0 {
		return world.ErrEmptyObjectKey
	}
	if err := o.GetManifestBundle().Validate(); err != nil {
		return err
	}
	return nil
}

// ApplyWorldOp applies the operation as a world operation.
func (o *ExtractManifestBundleOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	// store the object for the manifest
	// Extract the manifest bundle into the World object.
	obj, _, _, err := ExtractManifestBundle(ctx, ws, sender, o.GetObjectKey(), o.GetManifestBundle())
	world.ReleaseObjectState(obj)
	if err != nil {
		return false, err
	}

	// Link the manifest object under each link object key.
	for _, objKey := range o.GetLinkObjectKeys() {
		quad := NewManifestQuad(objKey, o.GetObjectKey(), "")
		if err := ws.SetGraphQuad(ctx, quad); err != nil {
			return false, err
		}
	}

	return false, nil
}

// _ is a type assertion
var _ world.Operation = (*ExtractManifestBundleOp)(nil)
