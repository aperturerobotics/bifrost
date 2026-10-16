package s4wave_device

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
)

// ComputersDashboardTypeID is the object type identifier for Computers dashboards.
const ComputersDashboardTypeID = "spacewave/computers"

// UnmarshalComputersDashboard unmarshals a ComputersDashboard from a cursor.
func UnmarshalComputersDashboard(ctx context.Context, bcs *block.Cursor) (*ComputersDashboard, error) {
	return block.UnmarshalBlock[ComputersDashboard](ctx, bcs)
}

// GetBlockTypeId returns the block type identifier.
func (d *ComputersDashboard) GetBlockTypeId() string {
	return ComputersDashboardTypeID
}

// Validate performs cursory checks on the ComputersDashboard.
func (d *ComputersDashboard) Validate() error {
	return nil
}
