package core

import (
	"context"
	"testing"
	"time"

	"github.com/s4wave/spacewave/db/background"
	"github.com/sirupsen/logrus"
)

// TestCoreBusRunsBackgroundJobs verifies the bus loads the background
// controller that the plugin scheduler's manifest copy waits on.
func TestCoreBusRunsBackgroundJobs(t *testing.T) {
	// Build the bus.
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	b, _, err := NewCoreBus(ctx, logrus.NewEntry(logrus.New()))
	if err != nil {
		t.Fatal(err)
	}

	// Run a job to completion.
	if err := background.ExRunJob(ctx, b, "probe", func(context.Context) error { return nil }); err != nil {
		t.Fatalf("background job did not run: %v", err)
	}
}
