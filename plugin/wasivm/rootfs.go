package wasivm_plugin

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/cdn"
	cdn_sharedobject "github.com/s4wave/spacewave/core/cdn/sharedobject"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	"github.com/sirupsen/logrus"
)

// rootfsSource is the UnixFS object holding the root filesystem of a VM and
// the World it lives in.
type rootfsSource struct {
	// ws is the World holding the object.
	ws world.WorldState
	// key is the object key.
	key string
	// release closes the CDN Space mounted to read the object, if any.
	release func()
}

// resolveRootfs finds the root filesystem of the VM vmKey in ws: the
// wasivm/rootfs edge when set, else the wasivmimage/rootfs edge of the image
// linked by wasivm/image, else that edge on the published image named by the
// image's cdn_source. A published image's CDN Space stays mounted until the
// source is released or ctx ends, so the run reads the filesystem lazily.
func resolveRootfs(
	ctx context.Context,
	le *logrus.Entry,
	b bus.Bus,
	ws world.WorldState,
	vmKey string,
) (*rootfsSource, error) {
	// Prefer the filesystem linked to the VM itself.
	key, ok, err := world.LookupGraphObjectKey(ctx, ws, vmKey, string(s4wave_vm.PredWasiVmRootfs))
	if err != nil {
		return nil, err
	}
	if ok {
		return &rootfsSource{ws: ws, key: key, release: func() {}}, nil
	}

	// Use the filesystem of the linked image.
	imageKey, ok, err := world.LookupGraphObjectKey(ctx, ws, vmKey, string(s4wave_vm.PredWasiVmImage))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("the VM has no root filesystem or image")
	}
	key, ok, err = world.LookupGraphObjectKey(ctx, ws, imageKey, string(s4wave_vm.PredWasiVmImageRootfs))
	if err != nil {
		return nil, err
	}
	if ok {
		return &rootfsSource{ws: ws, key: key, release: func() {}}, nil
	}

	// Fall back to the published image the linked image reads from.
	src, err := readImageCdnSource(ctx, ws, imageKey)
	if err != nil {
		return nil, err
	}
	if src.GetSpaceId() == "" || src.GetObjectKey() == "" {
		return nil, errors.Errorf("image %s has no root filesystem", imageKey)
	}

	// Mount the CDN Space and find the published image's filesystem in it.
	cdnWS, release, err := cdn_sharedobject.MountWorld(ctx, le, b, cdn.BaseURL(), src.GetSpaceId())
	if err != nil {
		return nil, errors.Wrapf(err, "mount the CDN Space %s", src.GetSpaceId())
	}
	key, ok, err = world.LookupGraphObjectKey(ctx, cdnWS, src.GetObjectKey(), string(s4wave_vm.PredWasiVmImageRootfs))
	if err == nil && !ok {
		err = errors.Errorf("published image %s/%s has no root filesystem", src.GetSpaceId(), src.GetObjectKey())
	}
	if err != nil {
		release()
		return nil, err
	}
	return &rootfsSource{ws: cdnWS, key: key, release: release}, nil
}

// readImageCdnSource reads the cdn_source of the WasiVmImage at imageKey.
func readImageCdnSource(ctx context.Context, ws world.WorldState, imageKey string) (*s4wave_vm.VmImageCdnSource, error) {
	// Open the image object.
	objState, found, err := ws.GetObject(ctx, imageKey)
	defer world.ReleaseObjectState(objState)
	if err != nil {
		return nil, errors.Wrapf(err, "get image %s", imageKey)
	}
	if !found {
		return nil, errors.Errorf("image %s not found", imageKey)
	}

	// Copy the source out of the image block.
	var src *s4wave_vm.VmImageCdnSource
	_, _, err = world.AccessObjectState(ctx, objState, false, func(bcs *block.Cursor) error {
		img, err := block.UnmarshalBlock[*s4wave_vm.WasiVmImage](ctx, bcs, func() block.Block {
			return &s4wave_vm.WasiVmImage{}
		})
		if err != nil {
			return err
		}
		src = img.GetCdnSource().CloneVT()
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "read image %s", imageKey)
	}
	return src, nil
}
