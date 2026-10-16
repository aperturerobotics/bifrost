package block_store_http

import (
	"github.com/s4wave/spacewave/db/block"
)

// Validate validates the put response.
func (o *PutResponse) Validate() error {
	if o == nil {
		return nil
	}
	// allow empty ref only if err is not empty
	return o.GetRef().Validate(o.GetErr() != "")
}

// _ is a type assertion
var _ block.Block = (*PutResponse)(nil)
