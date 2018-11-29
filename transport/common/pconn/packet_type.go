package pconn

import (
	"github.com/pkg/errors"
)

// Validate returns an error if the type is not valid.
func (t PacketType) Validate() error {
	if t > PacketType_PacketType_CLOSE_LINK || t < PacketType(0) {
		return errors.Errorf("packet type out of range: %s", t.String())
	}

	return nil
}
