package stream_drpc

import (
	"context"

	"github.com/aperturerobotics/bifrost/stream"
	"storj.io/drpc/drpcconn"
)

// NewDrpcConn constructs a new dprc conn from a stream.
func NewDrpcConn(ctx context.Context, strm stream.Stream, opts *DrpcOpts) *drpcconn.Conn {
	return drpcconn.NewWithOptions(strm, opts.BuildOpts())
}
