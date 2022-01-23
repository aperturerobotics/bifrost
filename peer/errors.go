package peer

import "errors"

var (
	// ErrPeerIDEmpty is returned if the peer id cannot be empty.
	ErrPeerIDEmpty = errors.New("peer id cannot be empty")
	// ErrBodyEmpty is returned if the message body was empty.
	ErrBodyEmpty = errors.New("message body cannot be empty")
	// ErrSignatureInvalid is returned for an invalid signature.
	ErrSignatureInvalid = errors.New("message signature invalid")
	// ErrShortMessage is returned if a message is too short.
	ErrShortMessage = errors.New("message too short")
)
