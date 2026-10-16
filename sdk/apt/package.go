package s4wave_apt

import (
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
)

// Validate performs cursory checks on the AptPackage.
func (p *AptPackage) Validate() error {
	// Validate the package identity, archive reference, and checksums.
	if err := p.GetState().Validate(); err != nil {
		return err
	}
	if p.GetName() == "" {
		return errors.New("name is required")
	}
	if p.GetVersion() == "" {
		return errors.New("version is required")
	}
	if p.GetArchitecture() == "" {
		return errors.New("architecture is required")
	}
	if p.GetState() != AptPackageState_AptPackageState_IMPORTING && p.GetDebRef().GetEmpty() {
		return errors.New("deb_ref is required")
	}
	if !p.GetDebRef().GetEmpty() {
		if err := p.GetDebRef().Validate(false); err != nil {
			return err
		}
	}
	for _, checksum := range p.GetChecksums() {
		if checksum.GetAlgorithm() == "" {
			return errors.New("checksum algorithm is required")
		}
		if checksum.GetHex() == "" {
			return errors.New("checksum hex is required")
		}
	}
	return nil
}

var _ block.Block = (*AptPackage)(nil)
