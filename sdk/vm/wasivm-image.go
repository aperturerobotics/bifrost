package s4wave_vm

import (
	"context"
	"time"

	"github.com/aperturerobotics/cayley/quad"
	timestamppb "github.com/aperturerobotics/protobuf-go-lite/types/known/timestamppb"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/blocktype"
	"github.com/s4wave/spacewave/db/world"
	world_types "github.com/s4wave/spacewave/db/world/types"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// WasiVmImageTypeID is the type identifier for WasiVmImage objects.
const WasiVmImageTypeID = "vm/image/wasivm"

// WasiVmImageBlockType constructs WasiVmImage blocks for typed cursor reads.
var WasiVmImageBlockType = blocktype.NewBlockType(
	WasiVmImageTypeID,
	func() *WasiVmImage { return &WasiVmImage{} },
)

// PredWasiVmImageRootfs is the graph predicate linking a WasiVmImage to the
// UnixFS object holding its root filesystem.
var PredWasiVmImageRootfs = quad.IRI("wasivmimage/rootfs")

// GetBlockTypeId returns the block type identifier.
func (v *WasiVmImage) GetBlockTypeId() string {
	return WasiVmImageTypeID
}

// MarshalBlock marshals the block to binary.
func (v *WasiVmImage) MarshalBlock() ([]byte, error) {
	return v.MarshalVT()
}

// UnmarshalBlock unmarshals the block from binary.
func (v *WasiVmImage) UnmarshalBlock(data []byte) error {
	return v.UnmarshalVT(data)
}

// Validate performs cursory checks on the WasiVmImage.
func (v *WasiVmImage) Validate() error {
	if v.GetName() == "" {
		return errors.New("name is required")
	}
	return nil
}

// CreateWasiVmImageOpId is the operation id for CreateWasiVmImageOp.
var CreateWasiVmImageOpId = "vm/image/wasivm/create"

// NewCreateWasiVmImageOp constructs a new CreateWasiVmImageOp.
func NewCreateWasiVmImageOp(objKey string, img *WasiVmImage, ts time.Time) *CreateWasiVmImageOp {
	return &CreateWasiVmImageOp{
		ObjectKey: objKey,
		Image:     img,
		Timestamp: timestamppb.New(ts),
	}
}

// NewCreateWasiVmImageOpBlock constructs a new CreateWasiVmImageOp block.
func NewCreateWasiVmImageOpBlock() block.Block {
	return &CreateWasiVmImageOp{}
}

// GetOperationTypeId returns the operation type identifier.
func (o *CreateWasiVmImageOp) GetOperationTypeId() string {
	return CreateWasiVmImageOpId
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

// ApplyWorldObjectOp applies the operation to a world object handle.
func (o *CreateWasiVmImageOp) ApplyWorldObjectOp(
	ctx context.Context,
	le *logrus.Entry,
	os world.ObjectState,
	sender peer.ID,
) (sysErr bool, err error) {
	return false, world.ErrUnhandledOp
}

// MarshalBlock marshals the block to binary.
func (o *CreateWasiVmImageOp) MarshalBlock() ([]byte, error) {
	return o.MarshalVT()
}

// UnmarshalBlock unmarshals the block to the object.
func (o *CreateWasiVmImageOp) UnmarshalBlock(data []byte) error {
	return o.UnmarshalVT(data)
}

// LookupCreateWasiVmImageOp looks up a CreateWasiVmImageOp operation type.
func LookupCreateWasiVmImageOp(ctx context.Context, operationTypeID string) (world.Operation, error) {
	if operationTypeID == CreateWasiVmImageOpId {
		return &CreateWasiVmImageOp{}, nil
	}
	return nil, nil
}

// _ is a type assertion
var (
	_ block.Block     = (*WasiVmImage)(nil)
	_ world.Operation = (*CreateWasiVmImageOp)(nil)
)
