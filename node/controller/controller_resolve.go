package controller

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
// Any fatal error resolving the value is returned.
// When the context is canceled valCh will not be drained anymore.
func (c *getNodeResolver) Resolve(ctx context.Context, valCh chan<- directive.Value) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case valCh <- node.Node(c.controller):
		return nil
	}
}

// resolveGetNode resolves the GetNode directive
func (c *Controller) resolveGetNode(d node.GetNode) directive.Resolver {
	return newGetNodeResolver(d, c)
}

// _ is a type assertion
var _ directive.Resolver = ((*getNodeResolver)(nil))
