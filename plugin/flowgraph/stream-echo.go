// Package flowgraph_plugin supplies the Stream Echo Device node and its settings editor.
package flowgraph_plugin

import (
	"context"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/pkg/errors"
	stream_echo "github.com/s4wave/spacewave/net/stream/echo"
	s4wave_flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
)

// StreamEchoNodeTypeID selects the plugin's echo service.
const StreamEchoNodeTypeID = "flowgraph/stream-echo"

// StreamEcho serves a peer stream that returns every byte sent to it.
type StreamEcho struct{}

// NewStreamEcho constructs the plugin's node type.
func NewStreamEcho() *StreamEcho {
	return &StreamEcho{}
}

// GetDisplayName returns the name shown for the node type.
func (n *StreamEcho) GetDisplayName() string {
	return "Stream Echo"
}

// GetPorts declares the stream output that Local Ports may connect to.
func (n *StreamEcho) GetPorts() []*s4wave_flowgraph.FlowgraphPort {
	return []*s4wave_flowgraph.FlowgraphPort{{
		Name:      "stream",
		Direction: s4wave_flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_OUTPUT,
		TypeId:    "stream",
	}}
}

// GetConfigIDs declares the existing stream echo controller.
func (n *StreamEcho) GetConfigIDs() []string {
	return []string{stream_echo.ConfigID}
}

// Compile serves every connected stream, or the explicit protocol_id parameter.
func (n *StreamEcho) Compile(ctx context.Context, node *s4wave_flowgraph.PlacedFlowgraphNode) (map[string]config.Config, error) {
	// Require the Device peer to prevent echoing on another peer's streams.
	if node.DevicePeerID == "" {
		return nil, errors.New("Stream Echo requires a Device peer")
	}

	// Compile graph connections using the protocol shared with Local Port nodes.
	entries := make(map[string]config.Config, len(node.Connections)+1)
	for _, conn := range node.Connections {
		if conn.Connection.GetOutputNode() == node.NodeID && conn.Connection.GetOutputPort() == "stream" {
			entries["connection/"+conn.ID] = &stream_echo.Config{PeerId: node.DevicePeerID, ProtocolId: "spacewave/" + node.FlowgraphKey + "/" + conn.ID}
		}
	}

	// Serve a chosen protocol even when no Local Port is connected.
	if protocolID := node.Node.GetParameters()["protocol_id"]; protocolID != "" {
		entries["protocol"] = &stream_echo.Config{PeerId: node.DevicePeerID, ProtocolId: protocolID}
	}
	for _, conf := range entries {
		if err := conf.Validate(); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

// _ is a type assertion.
var _ s4wave_flowgraph.FlowgraphNodeType = (*StreamEcho)(nil)
