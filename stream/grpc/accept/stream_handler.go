package stream_grpc_accept

import (
	"context"
	"io"

	"github.com/aperturerobotics/bifrost/link"
	stream_grpc "github.com/aperturerobotics/bifrost/stream/grpc/rpc"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/sirupsen/logrus"
)

// MountedStreamHandler implements the mounted stream handler.
type MountedStreamHandler struct {
	// le is the logger entry
	le *logrus.Entry
	// rpc is the rpc to use
	rpc *queuedRPC
	// b is the bus
	b bus.Bus
}

// NewMountedStreamHandler constructs the mounted stream handler.
func NewMountedStreamHandler(
	le *logrus.Entry,
	bus bus.Bus,
	rpc *queuedRPC,
) (*MountedStreamHandler, error) {
	return &MountedStreamHandler{le: le, rpc: rpc, b: bus}, nil
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
	_, estLinkInst, err := m.b.AddDirective(
		link.NewEstablishLinkWithPeer(strm.GetLink().GetLocalPeer(), strm.GetPeerID()),
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		defer estLinkInst.Release()

		if err := rpc.Send(&stream_grpc.Data{
			State: stream_grpc.StreamState_StreamState_ESTABLISHED,
		}); err != nil {
			s.Close()
			return
		}

		if err := stream_grpc.AttachRPCToStream(rpc, s, nil); err != nil &&
			err != io.EOF &&
			err != context.Canceled {
			m.le.WithError(err).Warn("rpc stream returned an error")
		}
	}()
	return nil
}

var _ link.MountedStreamHandler = ((*MountedStreamHandler)(nil))
