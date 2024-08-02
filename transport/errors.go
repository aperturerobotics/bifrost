package transport

import "errors"

var (
	// ErrNotTransportDialer is returned if a Transport does not implement TransportDialer.
	ErrNotTransportDialer = errors.New("transport does not implement a dialer")
)
