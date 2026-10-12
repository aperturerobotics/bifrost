package s4wave_vm_world

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/world"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	"github.com/sirupsen/logrus"
)

// startWasiVmKernel reports that this build cannot run WebAssembly VMs.
func startWasiVmKernel(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	rootfsKey string,
	config *s4wave_vm.WasiVmConfig,
	exited func(code int32, err error),
) (*wasiVmConsole, error) {
	return nil, errors.New("this build cannot run WebAssembly VMs")
}
