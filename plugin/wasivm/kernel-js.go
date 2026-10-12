//go:build js

package wasivm_plugin

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/world"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	"github.com/sirupsen/logrus"
)

// startKernel reports that the browser cannot run a WasiVM yet.
func startKernel(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	rootfsKey string,
	config *s4wave_vm.WasiVmConfig,
	exited func(runErr error),
) (*console, error) {
	return nil, errors.New("WebAssembly VMs run only in the desktop app")
}
