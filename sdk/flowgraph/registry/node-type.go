// Package flowgraph_registry carries plugin node declarations and compiler RPCs.
package flowgraph_registry

import (
	"slices"
	"strings"

	"github.com/pkg/errors"
)

// Validate requires a namespaced type ID, a display name and distinct config declarations.
//
// A plugin type ID is a slash-separated path of nonempty segments, such as
// "flowgraph/stream-echo". Core node type IDs are bare, so the slash keeps a
// plugin from claiming a core ID and having the Device compile a core node
// with the plugin's compiler.
func (n *NodeType) Validate() error {
	// Require metadata that the Device uses to resolve and label the node.
	if n.GetTypeId() == "" || n.GetDisplayName() == "" {
		return errors.New("node type ID and display name are required")
	}

	// Keep plugin IDs disjoint from the bare core IDs.
	segments := strings.Split(n.GetTypeId(), "/")
	if len(segments) < 2 || slices.Contains(segments, "") {
		return errors.Errorf("node type ID %q must be a slash-separated namespaced path", n.GetTypeId())
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
