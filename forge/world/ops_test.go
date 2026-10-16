package forge_world_test

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/db/world"
	forge_cluster "github.com/s4wave/spacewave/forge/cluster"
	forge_world "github.com/s4wave/spacewave/forge/world"
)

func TestLookupWorldOpReconstructsClusterJobTransitions(t *testing.T) {
	// Verify lookup reconstructs the cluster job start operation.
	start, err := world.NewSchema(forge_world.Types).LookupOp(context.Background(), forge_cluster.ClusterStartJobOpTypeID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := start.(*forge_cluster.ClusterStartJobOp); !ok {
		t.Fatalf("lookup %q returned %T", forge_cluster.ClusterStartJobOpTypeID, start)
	}

	// Verify lookup reconstructs the cluster job completion operation.
	complete, err := world.NewSchema(forge_world.Types).LookupOp(context.Background(), forge_cluster.ClusterCompleteJobOpTypeID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := complete.(*forge_cluster.ClusterCompleteJobOp); !ok {
		t.Fatalf("lookup %q returned %T", forge_cluster.ClusterCompleteJobOpTypeID, complete)
	}
}
