package provider

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/net/util/labels"
)

// ValidateProviderID validates a provider identifier.
func ValidateProviderID(id string) error {
	if id == "" {
		return ErrEmptyProviderID
	}
	if err := labels.ValidateDNSLabel(id); err != nil {
		return errors.Wrap(err, "provider id")
	}
	return nil
}

// UnmarshalProviderInfo unmarshals a ProviderInfo from a block cursor.
func UnmarshalProviderInfo(ctx context.Context, bcs *block.Cursor) (*ProviderInfo, error) {
	return block.UnmarshalBlock[ProviderInfo](ctx, bcs)
}

// Validate validates the ProviderInfo.
func (i *ProviderInfo) Validate() error {
	if err := ValidateProviderID(i.GetProviderId()); err != nil {
		return err
	}
	return nil
}

// _ is a type assertion
var _ block.Block = (*ProviderInfo)(nil)
