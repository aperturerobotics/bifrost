//go:build js

package opfs

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/coord"
	"github.com/s4wave/spacewave/db/coord/conformance"
	coord_inmem "github.com/s4wave/spacewave/db/coord/inmem"
)

// unreadableGenerationSource fails if keyed capability probes read global generations.
type unreadableGenerationSource struct{}

// RefreshGenerationContext rejects an unexpected durable-root read.
func (unreadableGenerationSource) RefreshGenerationContext(context.Context) (uint64, error) {
	return 0, errors.New("generation source must not be read")
}

// WaitGeneration rejects an unexpected durable-root subscription.
func (unreadableGenerationSource) WaitGeneration(context.Context, uint64) (uint64, error) {
	return 0, errors.New("generation source must not be watched")
}

// TestCoordinatorConformance checks the shared coordinator behavior contract.
func TestCoordinatorConformance(t *testing.T) {
	inner := coord_inmem.NewCoordinator()
	conformance.Check(t, func(testing.TB) (coord.Coordinator, coord.Coordinator) {
		return NewCoordinator(nil, "spacewave/test-volume", inner),
			NewCoordinator(nil, "spacewave/test-volume", inner)
	})
}

// TestKeyedCapabilityDoesNotReadGenerationStore keeps keyed leases independent of global revisions.
func TestKeyedCapabilityDoesNotReadGenerationStore(t *testing.T) {
	c := NewCoordinator(unreadableGenerationSource{}, "spacewave/test-volume", nil)
	capability, err := c.Capability(context.Background(), coord.Scope{
		VolumeID: "volume-a",
		Key:      "world-a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !capability.Supported || capability.Generations {
		t.Fatalf("unexpected keyed capability: %+v", capability)
	}
}
