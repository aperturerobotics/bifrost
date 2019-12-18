package stream_echo

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// EchoResolver resolves HandleMountedStream by echoing data.
type EchoResolver struct {
	// le is the logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
}

// NewEchoResolver constructs a new dial resolver.
func NewEchoResolver(le *logrus.Entry, bus bus.Bus) (*EchoResolver, error) {
	return &EchoResolver{le: le, bus: bus}, nil
}

// Resolve resolves the values, emitting them to the handler.
func (r *EchoResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	h, err := NewMountedStreamHandler(r.le, r.bus)
	if err != nil {
		return err
	}

	handler.AddValue(link.MountedStreamHandler(h))
	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*EchoResolver)(nil))
