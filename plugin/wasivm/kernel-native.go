//go:build !js

package wasivm_plugin

import (
	"context"
	"io"
	"time"

	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-billy/v6/memfs"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/unixfs"
	unixfs_billy "github.com/s4wave/spacewave/db/unixfs/billy"
	unixfs_overlay "github.com/s4wave/spacewave/db/unixfs/overlay"
	unixfs_world "github.com/s4wave/spacewave/db/unixfs/world"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine/native"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	"github.com/sirupsen/logrus"
)

// startKernel boots the kernel for one run on the root filesystem object
// rootfsKey. An in-memory upper layer holds the run's writes, so each run
// starts from the unchanged root filesystem. It returns the run's console once
// the kernel is booting. The kernel runs until ctx is canceled or init exits,
// then releases its resources, ends the console, and calls exited with the
// error that ended the run, or nil when init exited.
func startKernel(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	rootfsKey string,
	config *s4wave_vm.WasiVmConfig,
	exited func(runErr error),
) (*console, error) {
	// Open the root filesystem and the WebAssembly runtime.
	rootfs, releaseRootfs, err := openRootfs(ctx, le, ws, rootfsKey)
	if err != nil {
		return nil, err
	}
	rt := native.New(ctx)

	// Create the kernel with the console as its terminal.
	stdin, input := io.Pipe()
	c := &console{input: input}
	initPath := config.GetInitPath()
	if initPath == "" {
		initPath = s4wave_vm.DefaultWasiVmInitPath
	}
	k, err := kernel.New(ctx, rt, &kernel.Config{
		Stdout:          c,
		Stderr:          c,
		Stdin:           stdin,
		VirtualTerminal: true,
		RootFS:          rootfs,
		InitPath:        initPath,
		InitArgs:        config.GetInitArgs(),
		Env:             config.GetEnv(),
	})
	if err != nil {
		_ = rt.Close(ctx)
		releaseRootfs()
		return nil, errors.Wrap(err, "create the kernel")
	}
	c.resize = k.Resize

	// Run init, then release the run and report how it ended.
	go func() {
		// Run init until it exits or the run ends, then release the run.
		bootErr := k.Boot(ctx)
		_ = input.Close()
		cleanupCtx := context.WithoutCancel(ctx)
		_ = k.Close(cleanupCtx)
		_ = rt.Close(cleanupCtx)
		releaseRootfs()

		// Report how the run ended.
		var code int32
		var runErr error
		var exitErr *kernel.ProcessExitError
		switch {
		case bootErr == nil, ctx.Err() != nil:
		case errors.As(bootErr, &exitErr):
			code = int32(exitErr.Code) //nolint:gosec // an exit code fits in a byte.
		case errors.Is(bootErr, kernel.ErrNoInit):
			runErr = errors.Errorf("init %s is not in the root filesystem", initPath)
		default:
			runErr = bootErr
		}
		c.exit(code, runErr)
		exited(runErr)
	}()
	return c, nil
}

// openRootfs opens the UnixFS object rootfsKey as the root filesystem of a
// run, with an in-memory upper layer over it.
func openRootfs(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	rootfsKey string,
) (billy.Filesystem, func(), error) {
	// Open the object as the lower layer.
	fsType, _, err := unixfs_world.LookupFsType(ctx, ws, rootfsKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "look up the root filesystem %q", rootfsKey)
	}
	lower := unixfs_world.NewFSCursorWithContext(ctx, le, ws, rootfsKey, fsType, nil, false)

	// Overlay an in-memory upper layer and open the result as a filesystem.
	upper := unixfs_billy.NewBillyFSCursor(memfs.New(), "")
	handle, err := unixfs.NewFSHandle(unixfs_overlay.NewOverlayFSCursor(lower, upper))
	if err != nil {
		lower.Release()
		upper.Release()
		return nil, nil, errors.Wrap(err, "open the root filesystem")
	}
	return unixfs_billy.NewBillyFilesystem(ctx, handle, "/", time.Time{}), handle.Release, nil
}
