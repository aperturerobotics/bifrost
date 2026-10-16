package sobject_world_engine

import (
	"github.com/s4wave/spacewave/db/block"
)

// Validate validates the InitWorldOp configuration.
func (i *InitWorldOp) Validate() error {
	if err := i.GetTransformConf().Validate(); err != nil {
		return err
	}
	return nil
}

// _ is a type assertion
var _ block.Block = (*SOWorldOp)(nil)
