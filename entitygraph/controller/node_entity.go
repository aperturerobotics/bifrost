package entitygraph_controller

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/entitygraph/entity"
)

// NodeEntityTypeName is the entitygraph type name for a Bifrost transport
const NodeEntityTypeName = "bifrost/node"

// NewNodeEntityRef constructs a new entity ref to a node.
func NewNodeEntityRef(nodeID peer.ID) entity.Ref {
	return entity.NewEntityRefWithID(nodeID.Pretty(), NodeEntityTypeName)
}
