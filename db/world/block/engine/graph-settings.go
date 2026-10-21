package world_block_engine

import (
	"context"

	world_block "github.com/s4wave/spacewave/db/world/block"
	"github.com/sirupsen/logrus"
)

// graphSettingsStepMax bounds the links one graph settings step indexes or
// the postings it deletes, keeping each step's commit short next to writers.
const graphSettingsStepMax = 2048

// driveGraphSettings steps the engine's World to the target graph settings
// until the change finishes, ctx ends, or a step fails. A stopped change
// resumes from the stored progress the next time it runs.
func driveGraphSettings(ctx context.Context, le *logrus.Entry, engine *world_block.Engine, target *world_block.GraphSettings) {
	// Step until the World reaches the target.
	steps := 0
	for {
		done, err := engine.StepGraphSettings(ctx, target, graphSettingsStepMax)
		if err != nil {
			if ctx.Err() == nil {
				le.WithError(err).Warn("graph settings change stopped")
			}
			return
		}
		if done {
			break
		}
		steps++
	}

	// Report a change that took steps.
	if steps != 0 {
		le.WithField("steps", steps).Info("graph settings applied")
	}
}
