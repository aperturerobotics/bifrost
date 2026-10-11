// Package flowgraph_registry carries plugin node declarations and compiler RPCs.
package flowgraph_registry

import (
	"github.com/pkg/errors"
)

// Validate requires identifiers, a display name and distinct config declarations.
func (n *NodeType) Validate() error {
	// Require metadata that the Device uses to resolve and label the node.
	if n.GetTypeId() == "" || n.GetDisplayName() == "" {
		return errors.New("node type ID and display name are required")
	}

	// Keep the allowed controller set unambiguous.
	ids := make(map[string]struct{}, len(n.GetConfigIds()))
	for _, id := range n.GetConfigIds() {
		if _, exists := ids[id]; id == "" || exists {
			return errors.New("node type config IDs must be nonempty and distinct")
		}
		ids[id] = struct{}{}
	}
	return nil
}
