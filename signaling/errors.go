package signaling

import "errors"

var (
	// ErrEmptySignalingID is returned if the signaling id cannot be empty.
	ErrEmptySignalingID = errors.New("signaling id cannot be empty")
)
