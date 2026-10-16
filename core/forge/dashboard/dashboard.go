// Package forge_dashboard implements the ForgeDashboard block type and world operations.
package forge_dashboard

import (
	"github.com/aperturerobotics/cayley/quad"
)

// ForgeDashboardTypeID is the type identifier for ForgeDashboard objects.
const ForgeDashboardTypeID = "spacewave/forge/dashboard"

// PredDashboardForgeRef is the graph predicate for linking a dashboard to a Forge entity.
var PredDashboardForgeRef = quad.IRI("dashboard/forge-ref")

// GetBlockTypeId returns the block type identifier.
func (d *ForgeDashboard) GetBlockTypeId() string {
	return ForgeDashboardTypeID
}

// Validate performs cursory checks on the ForgeDashboard.
func (d *ForgeDashboard) Validate() error {
	return nil
}
