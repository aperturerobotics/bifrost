//go:build !js

package sobject_world_engine_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	sobject_world_engine "github.com/s4wave/spacewave/core/sobject/world/engine"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

// TestWorldCommitLogsStageTimes lowers the slow threshold to zero and checks
// that a World write logs one warning that names every stage with its
// duration, and the writer lock hold, which excludes the receipt wait.
func TestWorldCommitLogsStageTimes(t *testing.T) {
	// Capture the warnings the engine logs, then commit one write.
	w := startLocalWorld(t, t.TempDir(), nil)
	hook := logtest.NewLocal(w.le.Logger)
	defer sobject_world_engine.SetSlowCommitThreshold(0)()
	w.create(t, "stages/0")

	// One warning lists every stage with its duration, in stage order.
	var slow []*logrus.Entry
	for _, entry := range hook.AllEntries() {
		if entry.Message == "slow World commit" {
			slow = append(slow, entry)
		}
	}
	if len(slow) != 1 {
		t.Fatalf("got %d slow World commit warnings, want 1", len(slow))
	}
	elapsed, ok := slow[0].Data["elapsed"].(time.Duration)
	if !ok {
		t.Fatalf("warning lacks the elapsed duration: %v", slow[0].Data)
	}
	names := []string{
		"lock-wait", "fork-replay", "callback", "world-state-commit",
		"block-commit", "publication-admission", "advance", "receipt-wait",
	}
	hold, ok := slow[0].Data["hold"].(time.Duration)
	if !ok {
		t.Fatalf("warning lacks the hold duration: %v", slow[0].Data)
	}
	fields := strings.Fields(fmt.Sprint(slow[0].Data["stages"]))
	if len(fields) != len(names) {
		t.Fatalf("stages field has %d entries, want %d: %v", len(fields), len(names), fields)
	}

	// The stage durations add up to no more than the elapsed time of the
	// write, and the hold is the stages between the lock wait and the receipt
	// wait.
	var sum, held time.Duration
	for i, field := range fields {
		name, value, _ := strings.Cut(field, "=")
		if name != names[i] {
			t.Fatalf("stage %d is %s, want %s", i, name, names[i])
		}
		d, err := time.ParseDuration(value)
		if err != nil {
			t.Fatalf("stage %s has duration %q: %v", name, value, err)
		}
		sum += d
		if i != 0 && i != len(fields)-1 {
			held += d
		}
	}
	if sum > elapsed {
		t.Fatalf("stages sum to %v, more than the elapsed %v", sum, elapsed)
	}
	if held != hold {
		t.Fatalf("hold is %v, want the %v the held stages sum to", hold, held)
	}
}
