package web_pkg_fs_controller

import (
	"context"
	"io/fs"
	"path"
	"strings"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/pkg/errors"
	bldr_plugin "github.com/s4wave/spacewave/bldr/plugin"
	web_pkg "github.com/s4wave/spacewave/bldr/web/pkg"
	web_pkg_controller "github.com/s4wave/spacewave/bldr/web/pkg/controller"
	web_pkg_fs "github.com/s4wave/spacewave/bldr/web/pkg/fs"
	web_pkg_static "github.com/s4wave/spacewave/bldr/web/pkg/static"
	"github.com/s4wave/spacewave/db/unixfs"
	unixfs_access "github.com/s4wave/spacewave/db/unixfs/access"
	unixfs_errors "github.com/s4wave/spacewave/db/unixfs/errors"
	unixfs_iofs "github.com/s4wave/spacewave/db/unixfs/iofs"
	"github.com/sirupsen/logrus"
)

// ControllerID is the controller identifier.
const ControllerID = "bldr/web/pkg/fs/controller"

// Version is the controller version.
var Version = controller.MustParseVersion("0.0.1")

// Controller uses AccessUnixFS to resolve LookupWebPkg directives.
type Controller = web_pkg_controller.Controller

// NewController constructs a new web pkg fs controller.
func NewController(
	le *logrus.Entry,
	b bus.Bus,
	cc *Config,
) (*Controller, error) {
	return web_pkg_controller.NewController(
		le,
		controller.NewInfo(ControllerID, Version, "web pkg fs controller"),
		NewWebPkgGetter(b, cc.GetUnixfsId(), cc.GetUnixfsPrefix(), cc.GetNotFoundIfIdle()),
		cc.GetWebPkgIdList(),
	), nil
}

// NewWebPkgGetter constructs a new web pkg getter function.
//
// A package on a manifest-bound plugin assets filesystem resolves at once: its
// immutable URL is all the import map and the HTTP redirect need, and the
// plugin declared the package when it built, so the getter opens the
// filesystem only when a caller asks for the package's files.
func NewWebPkgGetter(b bus.Bus, unixFsID, unixFsPrefix string, returnIfIdle bool) web_pkg_controller.WebPkgGetter {
	return func(ctx context.Context, webPkgID string, released func()) (web_pkg.LookupWebPkgValue, func(), error) {
		// Pin module URLs to the manifest when the filesystem serves an immutable one.
		assetBasePath, err := pinnedAssetBasePath(unixFsID, unixFsPrefix, webPkgID)
		if err != nil {
			return nil, nil, err
		}
		if assetBasePath != "" {
			pkg, err := web_pkg_static.NewStaticWebPkg(
				&web_pkg.WebPkgInfo{Id: webPkgID, AssetBasePath: assetBasePath},
				func(ctx context.Context) (*unixfs.FSHandle, error) {
					return openWebPkgFs(ctx, b, unixFsID, path.Join(unixFsPrefix, webPkgID))
				},
			)
			return pkg, nil, err
		}

		// Resolve the UnixFS directive that supplies the package filesystem.
		val, valRef, err := unixfs_access.ExAccessUnixFS(ctx, b, unixFsID, returnIfIdle, released)
		if err != nil {
			return nil, nil, err
		}
		if valRef == nil {
			return nil, nil, errors.Wrap(unixfs_errors.ErrFsNotFound, unixFsID)
		}

		// Acquire a filesystem handle while retaining the UnixFS directive.
		fsHandle, fsHandleRel, err := val(ctx, released)
		if err != nil {
			valRef.Release()
			return nil, nil, err
		}

		// Resolve the configured package prefix within the filesystem.
		var childHandle *unixfs.FSHandle
		if unixFsPrefix != "" {
			childHandle, _, err = fsHandle.LookupPath(ctx, unixFsPrefix)
			if err != nil {
				fsHandleRel()
				valRef.Release()
				return nil, nil, err
			}
		}

		// Adapt the selected filesystem root for Web package lookup.
		var ifs fs.FS
		if childHandle != nil {
			ifs = unixfs_iofs.NewFS(ctx, childHandle)
		} else {
			ifs = unixfs_iofs.NewFS(ctx, fsHandle)
		}

		// Load the Web package and release filesystem references if it is absent.
		pkg, pkgRel, err := web_pkg_fs.GetWebPkg(ctx, ifs, webPkgID)
		if err != nil || pkg == nil {
			if childHandle != nil {
				childHandle.Release()
			}
			fsHandleRel()
			valRef.Release()
			return nil, nil, err
		}

		return pkg, func() {
			pkgRel()
			if childHandle != nil {
				childHandle.Release()
			}
			fsHandleRel()
			valRef.Release()
		}, nil
	}
}

// openWebPkgFs opens the package directory at pkgPath on the UnixFS unixFsID.
// Releasing the returned handle releases the filesystem it came from.
func openWebPkgFs(ctx context.Context, b bus.Bus, unixFsID, pkgPath string) (*unixfs.FSHandle, error) {
	// Resolve the UnixFS directive and a root handle on it.
	val, valRef, err := unixfs_access.ExAccessUnixFS(ctx, b, unixFsID, false, nil)
	if err != nil {
		return nil, err
	}
	if valRef == nil {
		return nil, errors.Wrap(unixfs_errors.ErrFsNotFound, unixFsID)
	}
	fsHandle, fsHandleRel, err := val(ctx, nil)
	if err != nil {
		valRef.Release()
		return nil, err
	}
	release := func() {
		fsHandleRel()
		valRef.Release()
	}

	// Look up the package directory and tie the root's lifetime to it.
	pkgHandle, _, err := fsHandle.LookupPath(ctx, pkgPath)
	if err != nil {
		release()
		return nil, err
	}
	pkgHandle.AddReleaseCallback(release)
	return pkgHandle, nil
}

// pinnedAssetBasePath returns the immutable URL prefix that serves webPkgID, or
// empty when unixFsID is not a manifest-bound plugin assets filesystem.
func pinnedAssetBasePath(unixFsID, unixFsPrefix, webPkgID string) (string, error) {
	// Only plugin assets filesystems have an immutable URL.
	artifactID, ok := strings.CutPrefix(unixFsID, bldr_plugin.PluginAssetsFsIdPrefix)
	if !ok {
		return "", nil
	}

	// An unpinned plugin binding serves its files directly.
	_, manifestRoot, err := bldr_plugin.ParsePluginArtifactID(artifactID, false)
	if err != nil || manifestRoot == "" {
		return "", err
	}
	return bldr_plugin.PluginAssetsHttpPrefix + artifactID + "/" + strings.Trim(unixFsPrefix, "/") + "/" + webPkgID + "/", nil
}
