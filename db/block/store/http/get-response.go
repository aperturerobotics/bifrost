package block_store_http

import (
	"github.com/s4wave/spacewave/db/block"
)

// _ is a type assertion
var _ block.Block = (*GetResponse)(nil)
