package node_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/node"
	"github.com/aperturerobotics/controllerbus/directive"
)

// getNodeResolver resolves the GetNode directive
type getNodeResolver struct {
	directive  node.GetNode
	controller *Controller
}

// newGetNodeResolver constructs a new GetNode resolver
func newGetNodeResolver(
	directive node.GetNode,
	controller *Controller,
) *getNodeResolver {
	peerID := directive.GetNodePeerIDConstraint()
	if len(peerID) != 0 {
		npID := controller.GetPeerID()
		if npID != peerID {
			return nil
		}
	}

	return &getNodeResolver{
		directive:  directive,
		controller: controller,
	}
}

// Resolve resolves the values.
func (c *getNodeResolver) Resolve(ctx context.Context, valHandler directive.ResolverHandler) error {
	_, _ = valHandler.AddValue(node.Node(c.controller))
	return nil
}

// resolveGetNode resolves the GetNode directive
func (c *Controller) resolveGetNode(d node.GetNode) directive.Resolver {
	return newGetNodeResolver(d, c)
}

// _ is a type assertion
var _ directive.Resolver = ((*getNodeResolver)(nil))
