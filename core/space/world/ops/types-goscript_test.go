//go:build goscript

package space_world_ops

import (
	"reflect"
	"testing"

	forge_dashboard "github.com/s4wave/spacewave/core/forge/dashboard"
	forge_job_ops "github.com/s4wave/spacewave/core/forge/job"
	forge_task_ops "github.com/s4wave/spacewave/core/forge/task"
	s4wave_git "github.com/s4wave/spacewave/core/git"
	"github.com/s4wave/spacewave/core/git/opids"
	"github.com/s4wave/spacewave/db/world"
	forge_cluster "github.com/s4wave/spacewave/forge/cluster"
	forge_execution_tx "github.com/s4wave/spacewave/forge/execution/tx"
	s4wave_device "github.com/s4wave/spacewave/sdk/device"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// TestGoScriptTypesResolveSpaceOps checks that the GoScript build schema
// declares the operations GoScript Spaces apply, one per contributing list.
func TestGoScriptTypesResolveSpaceOps(t *testing.T) {
	// Build the schema once and resolve one operation from each list.
	schema := world.NewSchema(Types)
	for _, tc := range []struct {
		id   string
		want any
	}{
		{s4wave_device.CreateComputersDashboardOpTypeID, (*s4wave_device.CreateComputersDashboardOp)(nil)},
		{forge_dashboard.InitForgeQuickstartOpTypeID, (*forge_dashboard.InitForgeQuickstartOp)(nil)},
		{forge_job_ops.ForgeJobCreateOpTypeID, (*forge_job_ops.ForgeJobCreateOp)(nil)},
		{forge_task_ops.ForgeTaskCreateOpTypeID, (*forge_task_ops.ForgeTaskCreateOp)(nil)},
		{forge_cluster.ClusterCreateOpTypeID, (*forge_cluster.ClusterCreateOp)(nil)},
		{forge_execution_tx.TxTypeID, (*forge_execution_tx.Tx)(nil)},
		{s4wave_vm.CreateVmV86OpTypeID, (*s4wave_vm.CreateVmV86Op)(nil)},
		{opids.CreateRepoWizard, (*s4wave_git.CreateGitRepoWizardOp)(nil)},
	} {
		// Resolve the type ID, then check the constructed operation type.
		op, err := schema.LookupOp(t.Context(), tc.id)
		if err != nil {
			t.Fatalf("LookupOp(%s): %v", tc.id, err)
		}
		if reflect.TypeOf(op) != reflect.TypeOf(tc.want) {
			t.Fatalf("LookupOp(%s) = %T, want %T", tc.id, op, tc.want)
		}
	}
}
