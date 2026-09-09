//go:build !js

package s4wave_core_e2e

import (
	"context"

	plugin_host_default "github.com/s4wave/spacewave/bldr/plugin/host/default"
	plugin_host_scheduler "github.com/s4wave/spacewave/bldr/plugin/host/scheduler"
	"github.com/s4wave/spacewave/bldr/testbed"
	"github.com/sirupsen/logrus"
)

// NewNativeTestbed builds the native E2E host with QuickJS reserved for the
// TypeScript fixture and frontend plugins. Go plugins run as native processes.
func NewNativeTestbed(ctx context.Context, le *logrus.Entry) (*testbed.Testbed, error) {
	return testbed.BuildTestbedWithSchedulerConfig(ctx, le, func(
		engineID, objectKey, volumeID, peerID string,
	) *plugin_host_scheduler.Config {
		return plugin_host_default.NewNativeDesktopSchedulerConfig(
			"", engineID, objectKey, volumeID, peerID, true, false, false,
			[]string{"spacewave-e2e", "spacewave-web", "spacewave-app"},
		)
	})
}
