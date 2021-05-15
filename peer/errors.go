package peer

import "errors"

var (
	// ErrPeerIDEmpty is returned if the peer id cannot be empty.
	ErrPeerIDEmpty = errors.New("peer id cannot be empty")
)
