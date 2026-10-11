// Package resource_flowgraph_registry resolves plugin node types through Resource capabilities.
package resource_flowgraph_registry

import (
	"context"

	"github.com/aperturerobotics/starpc/srpc"
	"github.com/aperturerobotics/util/broadcast"
	"github.com/pkg/errors"
	resource_server "github.com/s4wave/spacewave/bldr/resource/server"
	flowgraph_registry "github.com/s4wave/spacewave/sdk/flowgraph/registry"
)

// RegistryResource retains compilers for the lifetime of their registration Resources.
type RegistryResource struct {
	// bcast guards registrations and wakes node type lookups.
	bcast broadcast.Broadcast
	// registrations contains the sole compiler for each type ID under bcast.
	registrations map[string]*nodeTypeRegistration
}

// nodeTypeRegistration binds metadata and its compiler to one Resource lifetime.
type nodeTypeRegistration struct {
	// metadata is an immutable copy of the plugin's declaration.
	metadata *flowgraph_registry.NodeType
	// client invokes the compiler attached by the registering Resource client.
	client srpc.Client
	// ctx ends when the registration releases or the registering client disconnects.
	ctx context.Context
}

// NewRegistryResource constructs an empty node type registry.
func NewRegistryResource() *RegistryResource {
	return &RegistryResource{registrations: make(map[string]*nodeTypeRegistration)}
}

// RegisterNodeType retains a caller-attached compiler until its Resource releases.
func (r *RegistryResource) RegisterNodeType(ctx context.Context, req *flowgraph_registry.RegisterNodeTypeRequest) (*flowgraph_registry.RegisterNodeTypeResponse, error) {
	// Require a complete declaration before exposing a compiler.
	metadata := req.GetNodeType()
	if err := metadata.Validate(); err != nil {
		return nil, err
	}
	if req.GetHandlerResourceId() == 0 {
		return nil, errors.New("node type compiler resource is required")
	}

	// Bind the compiler to the caller's Resource generation.
	client, err := resource_server.MustGetResourceClientContext(ctx)
	if err != nil {
		return nil, err
	}
	compiler, err := client.GetAttachedResource(req.GetHandlerResourceId())
	if err != nil {
		return nil, err
	}
	lifetime, cancel := context.WithCancel(client.Context())
	reg := &nodeTypeRegistration{metadata: metadata.CloneVT(), client: compiler, ctx: lifetime}

	// Reject ambiguous compilers for the same node type.
	var duplicate bool
	r.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		if r.registrations[metadata.GetTypeId()] != nil {
			duplicate = true
			return
		}
		r.registrations[metadata.GetTypeId()] = reg
		broadcast()
	})
	if duplicate {
		cancel()
		return nil, errors.New("node type already registered")
	}

	// Withdraw the declaration and cancel in-flight compiles on Resource release.
	release := func() {
		cancel()
		r.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
			if r.registrations[metadata.GetTypeId()] == reg {
				delete(r.registrations, metadata.GetTypeId())
				broadcast()
			}
		})
	}
	resourceID, err := client.AddResource(srpc.NewMux(), release)
	if err != nil {
		release()
		return nil, err
	}
	return &flowgraph_registry.RegisterNodeTypeResponse{ResourceId: resourceID}, nil
}

// _ is a type assertion.
var _ flowgraph_registry.SRPCNodeTypeRegistryResourceServiceServer = (*RegistryResource)(nil)
