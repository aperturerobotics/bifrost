package stream_grpc_accept

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/sirupsen/logrus"
)

// MountedStreamHandler implements the mounted stream handler.
type MountedStreamHandler struct {
	// le is the logger entry
	le *logrus.Entry
	// rpc is the rpc to use
	rpc *queuedRPC
}

// NewMountedStreamHandler constructs the mounted stream handler.
func NewMountedStreamHandler(le *logrus.Entry, rpc *queuedRPC) (*MountedStreamHandler, error) {
	return &MountedStreamHandler{le: le, rpc: rpc}, nil
}

// HandleMountedStream handles an incoming mounted stream.
// Any returned error indicates the stream should be closed.
// This function should return as soon as possible, and start
// additional goroutines to manage the lifecycle of the stream.
func (m *MountedStreamHandler) HandleMountedStream(
	ctx context.Context,
	strm link.MountedStream,
) error {
	rpc := m.rpc.rpc
	s := strm.GetStream()
	return stream_grpc.AttachRPCToStream(rpc, s)
}

var _ link.MountedStreamHandler = ((*MountedStreamHandler)(nil))
