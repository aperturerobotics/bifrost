//go:build !js && !tinygo

package s4wave_vm_world

import (
	"context"
	"io"
	"time"

	"github.com/aperturerobotics/wasivm/wazero/kernel"
	wazero_runtime "github.com/aperturerobotics/wasivm/wazero/kernel/runtime/wazero"
	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-billy/v6/memfs"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/unixfs"
	unixfs_billy "github.com/s4wave/spacewave/db/unixfs/billy"
	unixfs_overlay "github.com/s4wave/spacewave/db/unixfs/overlay"
	unixfs_world "github.com/s4wave/spacewave/db/unixfs/world"
	"github.com/s4wave/spacewave/db/world"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	"github.com/sirupsen/logrus"
)

// startWasiVmKernel boots the WasiVM kernel for one run on the root
// filesystem object rootfsKey. An in-memory upper layer holds the run's
// writes, so each run starts from the unchanged rootfs. It returns the run's
// console once the kernel is booting. The kernel runs until ctx is canceled or
// init exits, then releases its resources, ends the console, and calls exited
// with init's exit code, or with the error that ended the run.
func startWasiVmKernel(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	rootfsKey string,
	config *s4wave_vm.WasiVmConfig,
	exited func(code int32, err error),
) (*wasiVmConsole, error) {
	// Open the root filesystem and the WebAssembly runtime.
	rootfs, releaseRootfs, err := openWasiVmRootfs(ctx, le, ws, rootfsKey)
	if err != nil {
		return nil, err
	}
	rt, err := wazero_runtime.New(ctx)
	if err != nil {
		releaseRootfs()
		return nil, errors.Wrap(err, "create wasm runtime")
	}

	// Create the kernel with its console as a virtual terminal.
	stdin, input := io.Pipe()
	console := &wasiVmConsole{input: input}
	initPath := config.GetInitPath()
	if initPath == "" {
		initPath = s4wave_vm.DefaultWasiVmInitPath
	}
	k, err := kernel.New(ctx, rt, &kernel.Config{
		Stdout:          console,
		Stderr:          console,
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
		return nil, errors.Wrap(err, "create kernel")
	}
	console.resize = k.Resize

	// Run init, then release the run and report how it ended.
	go func() {
		// Run init until it exits or the run ends, then release the run.
		bootErr := k.Boot(ctx)
		_ = input.Close()
		_ = k.Close(context.WithoutCancel(ctx))
		releaseRootfs()

		// Report how the run ended.
		var code int32
		var runErr error
		var exitErr *kernel.ProcessExitError
		switch {
		case bootErr == nil, ctx.Err() != nil:
		case errors.As(bootErr, &exitErr):
			code = int32(exitErr.Code) //nolint:gosec // a process exit code fits in a byte.
		case errors.Is(bootErr, kernel.ErrNoInit):
			runErr = errors.Errorf("init %s not found in the root filesystem", initPath)
		default:
			runErr = bootErr
		}
		console.exit(code, runErr)
		exited(code, runErr)
	}()
	return console, nil
}

// openWasiVmRootfs opens the UnixFS object rootfsKey as the root filesystem
// of a run, with an in-memory upper layer over it.
func openWasiVmRootfs(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	rootfsKey string,
) (billy.Filesystem, func(), error) {
	// Open the rootfs object as the lower layer.
	fsType, _, err := unixfs_world.LookupFsType(ctx, ws, rootfsKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "lookup rootfs object %q", rootfsKey)
	}
	lower := unixfs_world.NewFSCursorWithContext(ctx, le, ws, rootfsKey, fsType, nil, false)

	// Overlay an in-memory upper layer and open the result as a filesystem.
	upper := unixfs_billy.NewBillyFSCursor(memfs.New(), "")
	handle, err := unixfs.NewFSHandle(unixfs_overlay.NewOverlayFSCursor(lower, upper))
	if err != nil {
		lower.Release()
		upper.Release()
		return nil, nil, errors.Wrap(err, "open rootfs handle")
	}
	return unixfs_billy.NewBillyFilesystem(ctx, handle, "/", time.Time{}), handle.Release, nil
}
