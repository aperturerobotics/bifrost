package stream_grpc

import (
	"context"
	"io"
)

// RPC matches the GRPC request/response interface.
type RPC interface {
	// Context returns the context.
	Context() context.Context
	// Send sends a packet.
	Send(*Data) error
	// Recv receives a packet.
	Recv() (*Data, error)
}

// AttachRPCToStream attaches a RPC to a stream.
func AttachRPCToStream(rpc RPC, s io.ReadWriteCloser) error {
	// Read pump (from stream -> grpc)
	errCh := make(chan error, 3)
	go func() {
		defer s.Close()

		buf := make([]byte, 1500)
		d := &Data{}
		for {
			n, err := s.Read(buf)
			if err != nil {
				errCh <- err
				return
			}

			d.Data = buf[:n]
			err = rpc.Send(d)
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
