package stream_forwarding

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/directive"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

// DialResolver resolves HandleMountedStream by dialing a multiaddr.
type DialResolver struct {
	// le is the logger
	le *logrus.Entry
	// ma is the multiaddress to dial
	ma ma.Multiaddr
}

// NewDialResolver constructs a new dial resolver.
func NewDialResolver(le *logrus.Entry, ma ma.Multiaddr) (*DialResolver, error) {
	return &DialResolver{le: le, ma: ma}, nil
}

// Resolve resolves the values, emitting them to the handler.
func (r *DialResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	h, err := NewMountedStreamHandler(r.le, r.ma)
	if err != nil {
		return err
	}

	handler.AddValue(link.MountedStreamHandler(h))
	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*DialResolver)(nil))
