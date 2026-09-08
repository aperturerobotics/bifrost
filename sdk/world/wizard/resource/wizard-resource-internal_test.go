//go:build !js

package wizard_resource

import (
	"context"
	"errors"
	"testing"

	wizard "github.com/s4wave/spacewave/sdk/world/wizard"
)

func TestWizardResourceClosePreventsLateStatePublish(t *testing.T) {
	resource := NewWizardResource(nil, nil, "", &wizard.WizardState{Name: "Initial"})
	resource.Close()

	resource.setWizardStateWatchState(&wizard.WizardState{Name: "Late"}, 1)

	var snap *wizardStateWatchSnapshot
	resource.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		snap = resource.snapshotWizardStateWatchLocked()
	})
	if !errors.Is(snap.err, context.Canceled) {
		t.Fatalf("expected context.Canceled after late state publish, got %v", snap.err)
	}
}

func TestWizardResourceRejectsStaleWorldStatePublish(t *testing.T) {
	resource := NewWizardResource(nil, nil, "", &wizard.WizardState{Name: "Initial"})

	resource.setWizardStateWatchState(&wizard.WizardState{Name: "Current"}, 2)
	resource.setWizardStateWatchState(&wizard.WizardState{Name: "Stale"}, 1)

	var snap *wizardStateWatchSnapshot
	resource.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		snap = resource.snapshotWizardStateWatchLocked()
	})
	if snap.state.GetName() != "Current" {
		t.Fatalf("expected stale state to be ignored, got %q", snap.state.GetName())
	}
}
