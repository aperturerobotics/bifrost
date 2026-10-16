package forge_world

import (
	"slices"

	forge_cluster "github.com/s4wave/spacewave/forge/cluster"
	forge_execution_tx "github.com/s4wave/spacewave/forge/execution/tx"
	forge_pass_tx "github.com/s4wave/spacewave/forge/pass/tx"
	forge_task_tx "github.com/s4wave/spacewave/forge/task/tx"
	forge_worker "github.com/s4wave/spacewave/forge/worker"
)

// Types lists the Forge World operations.
var Types = slices.Concat(
	forge_cluster.ClusterTypes,
	forge_execution_tx.TxTypes,
	forge_pass_tx.TxTypes,
	forge_task_tx.TxTypes,
	forge_worker.WorkerTypes,
)
