//go:build js || tinygo

package s4wave_vm_world

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/world"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	"github.com/sirupsen/logrus"
)

// startWasiVmKernel reports that this build cannot run the WasiVM kernel.
func startWasiVmKernel(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	rootfsKey string,
	config *s4wave_vm.WasiVmConfig,
	exited func(code int32, err error),
) (*wasiVmConsole, error) {
	return nil, errors.New("this runtime cannot run WebAssembly VMs yet; start the VM from the desktop app or the spacewave daemon")
}
