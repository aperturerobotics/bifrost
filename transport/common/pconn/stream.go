package pconn

import (
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/xtaci/smux"
)

// smuxStream implements a stream between two machines with smux.
type smuxStream struct {
	*smux.Stream
}

// NewSmuxStream builds a new stream from an smux stream
func NewSmuxStream(sstream *smux.Stream) stream.Stream {
	return &smuxStream{Stream: sstream}
}
