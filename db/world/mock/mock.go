package world_mock

import "github.com/s4wave/spacewave/db/world"

// LookupMockOp looks up the mock operations.
var LookupMockOp = world.NewSchema(MockTypes).LookupOp
