package s4wave_vm

import (
	"context"
	"time"

	"github.com/aperturerobotics/cayley/quad"
	timestamppb "github.com/aperturerobotics/protobuf-go-lite/types/known/timestamppb"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	world_types "github.com/s4wave/spacewave/db/world/types"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// PredWasiVmImageRootfs is the graph predicate linking a WasiVmImage to the
// UnixFS object holding its root filesystem.
var PredWasiVmImageRootfs = quad.IRI("wasivmimage/rootfs")

// Validate performs cursory checks on the WasiVmImage.
func (v *WasiVmImage) Validate() error {
	if v.GetName() == "" {
		return errors.New("name is required")
	}
	return nil
}

// NewCreateWasiVmImageOp constructs a new CreateWasiVmImageOp.
func NewCreateWasiVmImageOp(objKey string, img *WasiVmImage, ts time.Time) *CreateWasiVmImageOp {
	return &CreateWasiVmImageOp{
		ObjectKey: objKey,
		Image:     img,
		Timestamp: timestamppb.New(ts),
	}
}

// Validate performs cursory checks on the op.
func (o *CreateWasiVmImageOp) Validate() error {
	if len(o.GetObjectKey()) == 0 {
		return world.ErrEmptyObjectKey
	}
	if err := o.GetTimestamp().Validate(false); err != nil {
		return err
	}
	if o.GetImage() == nil {
		return errors.New("image is required")
	}
	return o.GetImage().Validate()
}

// ApplyWorldOp applies the operation as a world operation.
func (o *CreateWasiVmImageOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	// Validate the image creation intent.
	if err := o.Validate(); err != nil {
		return false, err
	}

	// Create the image object with the operation's creation timestamp.
	objKey := o.GetObjectKey()
	img := o.GetImage().CloneVT()
	img.CreatedAt = o.GetTimestamp()
	createdObject, _, err := world.CreateWorldObject(ctx, ws, objKey, func(bcs *block.Cursor) error {
		bcs.SetBlock(img, true)
		return nil
	})
	world.ReleaseObjectState(createdObject)
	if err != nil {
		return false, err
	}

	// Mark the object with its block type.
	if err := world_types.SetObjectType(ctx, ws, objKey, WasiVmImageTypeID); err != nil {
		return false, err
	}
	return false, nil
}

// _ is a type assertion
var (
	_ block.Block     = (*WasiVmImage)(nil)
	_ world.Operation = (*CreateWasiVmImageOp)(nil)
)
