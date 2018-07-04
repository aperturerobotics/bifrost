package stream

import (
	"time"
)

// OpenOpts are optional arguments when opening a stream.
type OpenOpts struct {
	// Encrypted indicates the stream MUST be encrypted.
	// An error is returned if the stream cannot be encrypted.
	Encrypted bool
	// Reliable indicates the stream MUST be reliable / ordered.
	// An error is returned if the stream cannot be reliable.
	Reliable bool
}

// Stream is a stream-based data channel between two peers over a link.
type Stream interface {
	// Read data from the stream.
	Read(b []byte) (n int, err error)
	// Write data to the stream.
	Write(b []byte) (n int, err error)
	// SetReadDeadline sets the read deadline as defined by
	// net.Conn.SetReadDeadline.
	// A zero time value disables the deadline.
	SetReadDeadline(t time.Time) error
	// SetWriteDeadline sets the write deadline as defined by
	// net.Conn.SetWriteDeadline.
	// A zero time value disables the deadline.
	SetWriteDeadline(t time.Time) error
	// SetDeadline sets both read and write deadlines as defined by
	// net.Conn.SetDeadline.
	// A zero time value disables the deadlines.
	SetDeadline(t time.Time) error
	// Close implements net.Conn
	Close() error
}
