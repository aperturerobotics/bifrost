package bifrost_entitygraph

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/entitygraph/entity"
)

// NodeEntityTypeName is the entitygraph type name for a Bifrost transport
const NodeEntityTypeName = "bifrost/node"

// NodeEntity is a entity implementation backed by a node.
type NodeEntity struct {
	entityID, entityTypeName string
}

// NewNodeEntityRef constructs a new entity ref to a node.
func NewNodeEntityRef(nodeID peer.ID) entity.Ref {
	return entity.NewEntityRefWithID(nodeID.Pretty(), NodeEntityTypeName)
}

// NewNodeEntity constructs a new NodeEntity.
func NewNodeEntity(peerID peer.ID) *NodeEntity {
	nodRef := NewNodeEntityRef(peerID)
	nodID := nodRef.GetEntityRefId()
	return &NodeEntity{
		entityID:       nodID,
		entityTypeName: NodeEntityTypeName,
	}
}

// GetEntityID returns the entity identifier.
func (e *NodeEntity) GetEntityID() string {
	return e.entityID
}

// GetEntityTypeName returns the entity type name.
func (e *NodeEntity) GetEntityTypeName() string {
	return e.entityTypeName
}

var _ entity.Entity = ((*NodeEntity)(nil))
