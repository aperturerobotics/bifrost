package stream_grpcaccept

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
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

	// Read pump (from stream -> grpc)
	errCh := make(chan error, 3)
	go func() {
		defer s.Close()
		buf := make([]byte, 1500)
		for {
			n, err := s.Read(buf)
			if err != nil {
				errCh <- err
				return
			}

			err = rpc.Send(&Response{
				Data: buf[:n],
			})
			if err != nil {
				errCh <- err
				return
			}
		}
	}()

	// Write pump (grpc -> stream)
	go func() {
		defer s.Close()
		for {
			msg, err := rpc.Recv()
			if err != nil {
				errCh <- err
				return
			}

			_, err = s.Write(msg.GetData())
			if err != nil {
				errCh <- err
				return
			}
		}
	}()

	// Error catcher
	go func() {
		err := <-errCh
		m.rpc.doneCb(err)
	}()

	return nil
}

var _ link.MountedStreamHandler = ((*MountedStreamHandler)(nil))
