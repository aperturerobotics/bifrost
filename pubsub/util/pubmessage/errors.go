package pubmessage

import "errors"

var (
	// ErrInvalidChannelID is returned for an invalid inner channel ID.
	ErrInvalidChannelID = errors.New("invalid or empty channel id on pubmessage")
)
