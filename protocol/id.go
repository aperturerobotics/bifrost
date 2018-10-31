package protocol

import (
	"errors"
)

// ID is a protocol identifier.
type ID string

func (i ID) Validate() error {
	if i == "" {
		return errors.New("protocol id cannot be empty")
	}

	// TODO: validate protocol id
	return nil
}
