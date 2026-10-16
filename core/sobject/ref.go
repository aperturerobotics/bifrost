package sobject

import (
	"context"

	"github.com/s4wave/spacewave/core/bstore"
	provider "github.com/s4wave/spacewave/core/provider"
	"github.com/s4wave/spacewave/db/block"
	"github.com/sirupsen/logrus"
)

// NewSharedObjectRef builds a new SharedObjectRef.
func NewSharedObjectRef(providerID, providerAccountID, sobjectID, blockStoreID string) *SharedObjectRef {
	return &SharedObjectRef{
		ProviderResourceRef: &provider.ProviderResourceRef{
			Id:                sobjectID,
			ProviderId:        providerID,
			ProviderAccountId: providerAccountID,
		},
		BlockStoreId: blockStoreID,
	}
}

// UnmarshalSharedObjectRef unmarshals a SharedObjectRef from a block cursor.
func UnmarshalSharedObjectRef(ctx context.Context, bcs *block.Cursor) (*SharedObjectRef, error) {
	return block.UnmarshalBlock[SharedObjectRef](ctx, bcs)
}

// Validate validates the shared object ref.
func (i *SharedObjectRef) Validate() error {
	if err := i.GetProviderResourceRef().Validate(); err != nil {
		return err
	}
	if i.GetBlockStoreId() == "" {
		return bstore.ErrEmptyBlockStoreID
	}
	return nil
}

// GetLogger adds debug values to the logger.
func (i *SharedObjectRef) GetLogger(le *logrus.Entry) *logrus.Entry {
	return i.
		GetProviderResourceRef().
		GetLogger(le)
}

// _ is a type assertion
var _ block.Block = (*SharedObjectRef)(nil)
