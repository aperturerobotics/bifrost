//go:build !js

package sobject_world_engine_test

import (
	"testing"
	"time"

	sobject_world_engine "github.com/s4wave/spacewave/core/sobject/world/engine"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

// TestWorldCommitLogsStageTimes lowers the slow threshold to zero and checks
// that a World write logs one warning that names every stage with its duration.
func TestWorldCommitLogsStageTimes(t *testing.T) {
	// Capture the warnings the engine logs, then commit one write.
	w := startLocalWorld(t, t.TempDir(), nil)
	hook := logtest.NewLocal(w.le.Logger)
	defer sobject_world_engine.SetSlowCommitThreshold(0)()
	w.create(t, "stages/0")

	// One warning carries every stage as a duration field.
	var slow []*logrus.Entry
	for _, entry := range hook.AllEntries() {
		if entry.Message == "slow World commit" {
			slow = append(slow, entry)
		}
	}
	if len(slow) != 1 {
		t.Fatalf("got %d slow World commit warnings, want 1", len(slow))
	}
	stages := []string{
		"lock-wait", "fork-replay", "callback", "world-state-commit",
		"block-commit", "publication-admission", "receipt-wait", "advance",
	}
	for _, stage := range append(stages, "elapsed") {
		if _, ok := slow[0].Data[stage].(time.Duration); !ok {
			t.Fatalf("warning lacks the %s duration: %v", stage, slow[0].Data)
		}
	}

	// The stage durations add up to the elapsed time of the write.
	var sum time.Duration
	for _, stage := range stages {
		sum += slow[0].Data[stage].(time.Duration)
	}
	if elapsed := slow[0].Data["elapsed"].(time.Duration); sum > elapsed {
		t.Fatalf("stages sum to %v, more than the elapsed %v", sum, elapsed)
	}
}
