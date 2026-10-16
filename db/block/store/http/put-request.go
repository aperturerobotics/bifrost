package block_store_http

import (
	"github.com/s4wave/spacewave/db/block"
)

// Validate validates the put request.
func (o *PutRequest) Validate() error {
	if o == nil {
		return nil
	}
	if len(o.GetData()) == 0 {
		return block.ErrEmptyBlock
	}
	if err := o.GetPutOpts().Validate(); err != nil {
		return err
	}
	return nil
}

// _ is a type assertion
var _ block.Block = (*PutRequest)(nil)
