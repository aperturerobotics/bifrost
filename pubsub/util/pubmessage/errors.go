package pubmessage

import "errors"

var (
	// ErrInvalidSignature is returned for an invalid signature.
	ErrInvalidSignature = errors.New("invalid signature on pubmessage")
	// ErrInvalidChannelID is returned for an invalid inner channel ID.
	ErrInvalidChannelID = errors.New("invalid or empty channel id on pubmessage")
)
