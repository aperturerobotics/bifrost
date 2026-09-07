package cdn_sharedobject

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/core/space"
	"github.com/sirupsen/logrus"
)

// ResolveMountSharedObjectBody builds the resolver for a CDN-backed Space body.
func ResolveMountSharedObjectBody(
	le *logrus.Entry,
	b bus.Bus,
	dir sobject.MountSharedObjectBody,
) ([]directive.Resolver, error) {
	return directive.R(directive.NewAccessResolver(func(ctx context.Context, released func()) (space.MountSharedObjectBodyValue, func(), error) {
		// Admit only the CDN object implementation for this body type.
		cdnSO, ok := dir.MountSharedObjectBodySource().(*CdnSharedObject)
		if !ok {
			return nil, nil, errors.Errorf("cdn body type on non-cdn shared object: %T", dir.MountSharedObjectBodySource())
		}

		// Tie the read-only engine to the mounted body reference.
		we, err := NewWorldEngine(ctx, le, b, cdnSO)
		if err != nil {
			return nil, nil, errors.Wrap(err, "build cdn world engine")
		}
		var body space.SpaceSharedObjectBody = NewCdnSpaceBody(cdnSO, we)
		ret := sobject.NewMountSharedObjectBodyValue[space.SpaceSharedObjectBody](
			dir.MountSharedObjectBodyRef(),
			CdnBodyType,
			cdnSO,
			body,
		)

		// Release the engine when the mount is withdrawn.
		return ret, we.Release, nil
	}), nil)
}
