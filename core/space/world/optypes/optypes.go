package optypes

import (
	"context"

	space_world_ops "github.com/s4wave/spacewave/core/space/world/ops"
	"github.com/s4wave/spacewave/db/world"
	s4wave_appconnector "github.com/s4wave/spacewave/sdk/appconnector"
	s4wave_kv_world "github.com/s4wave/spacewave/sdk/kv/world"
	s4wave_sshhost "github.com/s4wave/spacewave/sdk/sshhost"
	s4wave_wizard "github.com/s4wave/spacewave/sdk/world/wizard"
)

// schema resolves the available world operation types.
//
// buildTypes adds the operations whose packages differ between builds.
var schema = world.NewSchema(
	space_world_ops.Types,
	s4wave_kv_world.WorldTypes,
	s4wave_appconnector.AppconnectorTypes,
	s4wave_sshhost.SshhostTypes,
	s4wave_wizard.WizardTypes,
	buildTypes,
)

// LookupWorldOp looks up the available world operation types.
func LookupWorldOp(ctx context.Context, opTypeID string) (world.Operation, error) {
	return schema.LookupOp(ctx, opTypeID)
}
