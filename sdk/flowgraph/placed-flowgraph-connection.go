package s4wave_flowgraph

// PlacedFlowgraphConnection is a connection that ends at a placed node.
type PlacedFlowgraphConnection struct {
	// ID is the graph-local ID of the connection.
	ID string
	// Connection is the authored connection.
	Connection *FlowgraphConnection
	// RemoteDevicePeerID is the peer ID of the Device running the other end.
	RemoteDevicePeerID string
}
