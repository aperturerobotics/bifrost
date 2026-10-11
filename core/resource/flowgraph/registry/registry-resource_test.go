package resource_flowgraph_registry

import (
	"testing"

	flowgraph_registry "github.com/s4wave/spacewave/sdk/flowgraph/registry"
)

// TestRegisterNodeTypeRejectsUnnamespacedID verifies a plugin cannot claim a bare core ID.
func TestRegisterNodeTypeRejectsUnnamespacedID(t *testing.T) {
	for _, typeID := range []string{"tcp-port", "flowgraph/", "/stream-echo", "flowgraph//stream-echo"} {
		// Register a declaration that fails validation before any compiler attaches.
		_, err := NewRegistryResource().RegisterNodeType(t.Context(), &flowgraph_registry.RegisterNodeTypeRequest{
			NodeType:          &flowgraph_registry.NodeType{TypeId: typeID, DisplayName: "Impostor"},
			HandlerResourceId: 1,
		})
		if err == nil {
			t.Fatalf("type ID %q was accepted", typeID)
		}
	}
}
