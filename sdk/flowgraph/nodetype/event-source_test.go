package flowgraph_nodetype

import (
	"testing"

	s4wave_flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
)

// TestEventSourceNodeType resolves the core Event Source node type and checks
// that it declares one event output and compiles no controller configs.
func TestEventSourceNodeType(t *testing.T) {
	// Look the type up in the core controller.
	nodeType, found := NewController().types[s4wave_flowgraph.EventSourceNodeTypeID]
	if !found {
		t.Fatalf("node type %q is not provided", s4wave_flowgraph.EventSourceNodeTypeID)
	}

	// The type declares one output port for events and runs no controller.
	ports := nodeType.GetPorts()
	if len(ports) != 1 || ports[0].GetName() != "event" || ports[0].GetDirection() != s4wave_flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_OUTPUT {
		t.Fatalf("ports = %v, want one output port event", ports)
	}
	if len(nodeType.GetConfigIDs()) != 0 {
		t.Fatalf("config IDs = %v, want none", nodeType.GetConfigIDs())
	}
	entries, err := nodeType.Compile(t.Context(), &s4wave_flowgraph.PlacedFlowgraphNode{NodeID: "source"})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("compiled %d entries, want none", len(entries))
	}
}
