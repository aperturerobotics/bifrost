package pconn

import (
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/hashicorp/yamux"
)

// yamuxWrapper wraps yamux.
type yamuxWrapper struct {
	*yamux.Session
}

// OpenStream opens a stream.
func (m *yamuxWrapper) OpenStream() (stream.Stream, error) {
	return m.Session.OpenStream()
}

// AcceptStream accepts a stream.
func (m *yamuxWrapper) AcceptStream() (stream.Stream, error) {
	return m.Session.AcceptStream()
}

// _ is a type assertion
var _ streamMuxer = ((*yamuxWrapper)(nil))
