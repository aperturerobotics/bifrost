package resource_flowgraph_registry

import (
	"context"
	"slices"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/starpc/srpc"
	"github.com/pkg/errors"
	s4wave_flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
	flowgraph_registry "github.com/s4wave/spacewave/sdk/flowgraph/registry"
)

// compileTimeout bounds one plugin Compile call, because the Device reconciles
// every node on one loop and an unresponsive plugin would stall all of them.
const compileTimeout = 30 * time.Second

// NodeTypeProxy compiles placements through a plugin's retained Resource capability.
type NodeTypeProxy struct {
	// b resolves compiled controller config constructors on the daemon bus.
	b bus.Bus
	// metadata is the immutable node type declaration.
	metadata *flowgraph_registry.NodeType
	// client invokes the registering plugin's compiler.
	client flowgraph_registry.SRPCNodeTypeResourceServiceClient
	// lifetime cancels compilation when the registration Resource releases.
	lifetime context.Context
}

// NewNodeTypeProxy constructs a compiler bound to one registration lifetime.
func NewNodeTypeProxy(b bus.Bus, metadata *flowgraph_registry.NodeType, client srpc.Client, lifetime context.Context) *NodeTypeProxy {
	return &NodeTypeProxy{b: b, metadata: metadata.CloneVT(), client: flowgraph_registry.NewSRPCNodeTypeResourceServiceClient(client), lifetime: lifetime}
}

// GetDisplayName returns the plugin's display name.
func (p *NodeTypeProxy) GetDisplayName() string {
	return p.metadata.GetDisplayName()
}

// GetPorts returns an independent copy of the declared endpoints.
func (p *NodeTypeProxy) GetPorts() []*s4wave_flowgraph.FlowgraphPort {
	return p.metadata.CloneVT().GetPorts()
}

// GetConfigIDs returns an independent copy of the allowed controller config IDs.
func (p *NodeTypeProxy) GetConfigIDs() []string {
	return slices.Clone(p.metadata.GetConfigIds())
}

// Compile invokes the plugin and decodes only its declared controller config IDs.
func (p *NodeTypeProxy) Compile(ctx context.Context, node *s4wave_flowgraph.PlacedFlowgraphNode) (map[string]config.Config, error) {
	// Cancel the call when the Device or the registration ends, or the plugin stalls.
	ctx, cancel := context.WithTimeout(ctx, compileTimeout)
	defer cancel()
	stop := context.AfterFunc(p.lifetime, cancel)
	defer stop()
	if err := p.lifetime.Err(); err != nil {
		return nil, err
	}

	// Send the authored node with the Device and remote placement identities.
	resp, err := p.client.Compile(ctx, flowgraph_registry.NewCompileRequest(node))
	if err != nil {
		return nil, err
	}

	// Check declarations before resolving constructors on the host bus.
	entries := make(map[string]config.Config, len(resp.GetEntries()))
	for name, encoded := range resp.GetEntries() {
		if name == "" || encoded == nil {
			return nil, errors.New("plugin returned an invalid node entry")
		}
		if !slices.Contains(p.metadata.GetConfigIds(), encoded.GetId()) {
			return nil, errors.Errorf("plugin emitted undeclared config ID %q", encoded.GetId())
		}
		resolved, err := encoded.Resolve(ctx, p.b)
		if err != nil {
			return nil, err
		}
		entries[name] = resolved.GetConfig()
	}
	return entries, nil
}

// _ is a type assertion.
var _ s4wave_flowgraph.FlowgraphNodeType = (*NodeTypeProxy)(nil)
