package stream_grpc

import (
	"context"

	"github.com/aperturerobotics/bifrost/stream"
)

// RPC matches the GRPC request/response interface.
type RPC interface {
	// Context returns the context.
	Context() context.Context
	// Send sends a packet.
	Send(*Response) error
	// Recv receives a packet.
	Recv() (*Request, error)
}

// AttachRPCToStream attaches a RPC to a stream.
func AttachRPCToStream(rpc RPC, s stream.Stream) error {
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

	// Return when any errors.
	return <-errCh
}
