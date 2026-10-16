package s4wave_wizard

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
)

// UnmarshalWizardState unmarshals a wizard state from a cursor.
func UnmarshalWizardState(ctx context.Context, bcs *block.Cursor) (*WizardState, error) {
	return block.UnmarshalBlock[WizardState](ctx, bcs)
}

// Validate performs cursory checks on the WizardState block.
func (s *WizardState) Validate() error {
	return nil
}

var _ block.Block = (*WizardState)(nil)
