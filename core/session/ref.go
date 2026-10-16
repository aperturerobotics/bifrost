package session

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
	"github.com/sirupsen/logrus"
)

// UnmarshalSessionRef unmarshals a SessionRef from a block cursor.
func UnmarshalSessionRef(ctx context.Context, bcs *block.Cursor) (*SessionRef, error) {
	return block.UnmarshalBlock[SessionRef](ctx, bcs)
}

// Validate validates the shared object ref.
func (i *SessionRef) Validate() error {
	if err := i.GetProviderResourceRef().Validate(); err != nil {
		return err
	}
	return nil
}

// GetLogger adds debug values to the logger.
func (i *SessionRef) GetLogger(le *logrus.Entry) *logrus.Entry {
	return i.
		GetProviderResourceRef().
		GetLogger(le)
}

// _ is a type assertion
var _ block.Block = (*SessionRef)(nil)
