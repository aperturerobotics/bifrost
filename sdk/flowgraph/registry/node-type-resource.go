package flowgraph_registry

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_proto "github.com/aperturerobotics/controllerbus/controller/configset/proto"
	"github.com/pkg/errors"
	s4wave_flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
)

// NodeTypeResource serves one plugin node type's compiler.
type NodeTypeResource struct {
	// typeID selects the sole type this Resource compiles.
	typeID string
	// nodeType implements compilation inside the plugin.
	nodeType s4wave_flowgraph.FlowgraphNodeType
}

// NewNodeTypeResource constructs a compiler Resource for one node type.
func NewNodeTypeResource(typeID string, nodeType s4wave_flowgraph.FlowgraphNodeType) *NodeTypeResource {
	return &NodeTypeResource{typeID: typeID, nodeType: nodeType}
}

// Compile encodes the plugin's configs without assigning Device keys or revisions.
func (r *NodeTypeResource) Compile(ctx context.Context, req *CompileRequest) (*CompileResponse, error) {
	// Keep the granted compiler restricted to its declared node type.
	if req.GetNode().GetTypeId() != r.typeID {
		return nil, errors.New("node type does not match compiler")
	}
	compiled, err := r.nodeType.Compile(ctx, req.PlacedNode())
	if err != nil {
		return nil, err
	}

	// Use ControllerBus's binary config encoding for the language boundary.
	entries := make(map[string]*configset_proto.ControllerConfig, len(compiled))
	for name, conf := range compiled {
		encoded, err := configset_proto.NewControllerConfig(configset.NewControllerConfig(0, conf), false)
		if err != nil {
			return nil, err
		}
		entries[name] = encoded
	}
	return &CompileResponse{Entries: entries}, nil
}

// NewCompileRequest copies a placement into its RPC representation.
func NewCompileRequest(node *s4wave_flowgraph.PlacedFlowgraphNode) *CompileRequest {
	// Copy the placement and all remote connection identities.
	req := &CompileRequest{FlowgraphKey: node.FlowgraphKey, NodeId: node.NodeID, Node: node.Node.CloneVT(), DevicePeerId: node.DevicePeerID, Connections: make([]*PlacedConnection, len(node.Connections))}
	for i, conn := range node.Connections {
		req.Connections[i] = &PlacedConnection{Id: conn.ID, Connection: conn.Connection.CloneVT(), RemoteDevicePeerId: conn.RemoteDevicePeerID}
	}
	return req
}

// PlacedNode copies the RPC context into the compiler's local interface.
func (r *CompileRequest) PlacedNode() *s4wave_flowgraph.PlacedFlowgraphNode {
	// Restore the node and all remote connection identities.
	node := &s4wave_flowgraph.PlacedFlowgraphNode{FlowgraphKey: r.GetFlowgraphKey(), NodeID: r.GetNodeId(), Node: r.GetNode().CloneVT(), DevicePeerID: r.GetDevicePeerId(), Connections: make([]*s4wave_flowgraph.PlacedFlowgraphConnection, len(r.GetConnections()))}
	for i, conn := range r.GetConnections() {
		node.Connections[i] = &s4wave_flowgraph.PlacedFlowgraphConnection{ID: conn.GetId(), Connection: conn.GetConnection().CloneVT(), RemoteDevicePeerID: conn.GetRemoteDevicePeerId()}
	}
	return node
}

// _ is a type assertion.
var _ SRPCNodeTypeResourceServiceServer = (*NodeTypeResource)(nil)
