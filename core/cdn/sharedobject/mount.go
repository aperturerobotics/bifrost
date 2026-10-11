package cdn_sharedobject

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/pkg/errors"
	cdn_bstore "github.com/s4wave/spacewave/core/cdn/bstore"
	"github.com/s4wave/spacewave/db/world"
	"github.com/sirupsen/logrus"
)

// MountWorld opens a read-only World over the published head of the CDN Space
// spaceID served from baseURL. The World owns its CDN transport and reads
// blocks lazily as they are first used. It keeps the head it mounted, so
// objects stay readable while the CDN publishes a newer one. The release
// function closes the World and its transport.
func MountWorld(
	ctx context.Context,
	le *logrus.Entry,
	b bus.Bus,
	baseURL, spaceID string,
) (world.WorldState, func(), error) {
	// Own the CDN transport until the mounted World is released.
	store, err := cdn_bstore.NewCdnBlockStore(cdn_bstore.Options{
		CdnBaseURL: baseURL,
		SpaceID:    spaceID,
		PointerTTL: -1,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "build cdn block store")
	}

	// Bind the published head to the transport's block store.
	so, err := NewCdnSharedObject(CdnSharedObjectOptions{
		SpaceID:    spaceID,
		Bus:        b,
		BlockStore: store,
	})
	if err != nil {
		store.Close()
		return nil, nil, errors.Wrap(err, "build cdn shared object")
	}

	// Keep the read engine and its transport under one release function.
	we, err := NewWorldEngine(ctx, le, b, so)
	if err != nil {
		store.Close()
		return nil, nil, errors.Wrap(err, "mount cdn world")
	}
	ws := world.NewEngineWorldState(we.Engine, false)
	return ws, func() {
		we.Release()
		store.Close()
	}, nil
}
