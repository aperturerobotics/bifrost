package world_mock

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// NewMockWorldOp constructs a new MockWorldOp block.
func NewMockWorldOp(objectKey, msg string) *MockWorldOp {
	return &MockWorldOp{
		ObjectKey: objectKey,
		NextMsg:   msg,
	}
}

// Validate performs cursory checks on the op.
func (m *MockWorldOp) Validate() error {
	if len(m.GetNextMsg()) == 0 {
		return ErrEmptyNextMsg
	}
	return nil
}

// ApplyWorldOp applies the operation as a world operation.
func (m *MockWorldOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	worldHandle world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	// Identify the World object and message targeted by the mock operation.
	nextMsg := m.GetNextMsg()
	objectKey := m.GetObjectKey()

	// re-use the logic for the object op
	objState, err := world.MustGetObject(ctx, worldHandle, objectKey)
	defer world.ReleaseObjectState(objState)
	if err != nil {
		return false, err
	}

	// Apply the mock message update through the object operation implementation.
	op := NewMockObjectOp(nextMsg)
	return op.ApplyWorldObjectOp(ctx, le, objState, sender)
}

// _ is a type assertion
var (
	_ block.Block     = (*MockWorldOp)(nil)
	_ world.Operation = (*MockWorldOp)(nil)
)
