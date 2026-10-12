//go:build !js

package bldr_project_controller

import (
	"context"

	"github.com/aperturerobotics/starpc/srpc"
	frontend "github.com/s4wave/spacewave/bldr/frontend"
)

// frontendWatchStream captures responses and announces Watch entry.
type frontendWatchStream struct {
	// Stream supplies the unused transport methods.
	srpc.Stream
	// ctx is the lifetime of this attachment.
	ctx context.Context
	// entered reports that Watch acquired the stream context.
	entered chan struct{}
	// events receives responses sent by Watch.
	events chan *frontend.Event
}

// Context announces Watch entry before returning its lifetime.
func (s *frontendWatchStream) Context() context.Context {
	select {
	case s.entered <- struct{}{}:
	default:
	}
	return s.ctx
}

// Send records an attachment response until the consumer cancels.
func (s *frontendWatchStream) Send(event *frontend.Event) error {
	select {
	case s.events <- event:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// SendAndClose records the final attachment response.
func (s *frontendWatchStream) SendAndClose(event *frontend.Event) error {
	return s.Send(event)
}

// _ is a type assertion.
var _ frontend.SRPCFrontend_WatchStream = (*frontendWatchStream)(nil)
