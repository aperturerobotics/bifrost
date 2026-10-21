package world_block_engine

import (
	"context"

	world_block "github.com/s4wave/spacewave/db/world/block"
	"github.com/sirupsen/logrus"
)

// driveGraphSettings steps the engine's World to the target graph settings
// until the change finishes, ctx ends, or a step fails. A stopped change
// resumes from the stored progress the next time it runs.
func driveGraphSettings(ctx context.Context, le *logrus.Entry, engine *world_block.Engine, target *world_block.GraphSettings) {
	// Step until the World reaches the target.
	steps, err := world_block.ApplyGraphSettings(ctx, engine, target, world_block.DefaultGraphSettingsStepMax)
	if err != nil {
		if ctx.Err() == nil {
			le.WithError(err).Warn("graph settings change stopped")
		}
		return
	}

	// Report a change that took steps.
	if steps != 0 {
		le.WithField("steps", steps).Info("graph settings applied")
	}
}
