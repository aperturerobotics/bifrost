// Package s4wave_apt implements apt repository block types and world operations.
package s4wave_apt

import (
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
)

// Validate performs cursory checks on the AptRepository.
func (r *AptRepository) Validate() error {
	// Validate the repository state, distribution, and index reference.
	if err := r.GetState().Validate(); err != nil {
		return err
	}
	if r.GetDistribution() == "" {
		return errors.New("distribution is required")
	}
	if len(r.GetComponents()) == 0 {
		return errors.New("components are required")
	}
	if len(r.GetArchitectures()) == 0 {
		return errors.New("architectures are required")
	}
	if r.GetState() == AptRepositoryState_AptRepositoryState_READY && r.GetIndexRef().GetEmpty() {
		return errors.New("index_ref is required when repository is ready")
	}
	if !r.GetIndexRef().GetEmpty() {
		return r.GetIndexRef().Validate()
	}
	return nil
}

var _ block.Block = (*AptRepository)(nil)
