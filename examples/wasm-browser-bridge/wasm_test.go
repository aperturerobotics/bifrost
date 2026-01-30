//go:build test_examples

package main

import (
	"context"
	"testing"
	"time"

	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/sim/graph"
	"github.com/aperturerobotics/bifrost/sim/simulate"
	stream_echo "github.com/aperturerobotics/bifrost/stream/echo"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/sirupsen/logrus"
)

// TestWASMBrowserBridge demonstrates browser-to-native communication.
//
// This test simulates the architecture where:
// - A browser runs Bifrost compiled to WebAssembly
// - A native Go peer handles requests
// - They communicate over WebRTC or WebSocket
//
// The test uses in-process transport to simulate the browser-native bridge.
func TestWASMBrowserBridge(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	le := logrus.NewEntry(log)

	// Create browser and native peers
	g := graph.NewGraph()

	// "Browser" peer (simulating WASM)
	browserPeer := addPeerWithEcho(t, g)

	// "Native" peer (backend server)
	nativePeer := addPeerWithEcho(t, g)

	// Connect them (simulates WebRTC/WebSocket link)
	lan := graph.AddLAN(g)
	lan.AddPeer(g, browserPeer)
	lan.AddPeer(g, nativePeer)

	sim := initSimulator(t, ctx, le, g)
	defer sim.Close()

	// Verify connectivity
	if err := simulate.TestConnectivity(ctx, sim.GetPeerByID(browserPeer.GetPeerID()), sim.GetPeerByID(nativePeer.GetPeerID())); err != nil {
		t.Fatalf("browser-native connectivity failed: %v", err)
	}

	le.Info("Browser-Native bridge established")
	le.Info("   This simulates: Browser (WASM) - WebRTC - Native Server")

	// In a real implementation, the browser would make HTTP requests
	// that get forwarded over WebRTC to the native peer
	// For this test, we verify basic connectivity works

	le.Info("WASM Browser Bridge test passed")
}

// TestBrowserToMultipleBackends tests one browser connecting to multiple backend services.
func TestBrowserToMultipleBackends(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	le := logrus.NewEntry(log)

	g := graph.NewGraph()

	// Browser peer
	browser := addPeerWithEcho(t, g)

	// Multiple backend services
	backends := make([]*graph.Peer, 3)
	for i := range backends {
		backends[i] = addPeerWithEcho(t, g)
	}

	// All on same network
	lan := graph.AddLAN(g)
	lan.AddPeer(g, browser)
	for _, b := range backends {
		lan.AddPeer(g, b)
	}

	sim := initSimulator(t, ctx, le, g)
	defer sim.Close()

	// Test connectivity to all backends
	for i, backend := range backends {
		if err := simulate.TestConnectivity(ctx, sim.GetPeerByID(browser.GetPeerID()), sim.GetPeerByID(backend.GetPeerID())); err != nil {
			t.Fatalf("browser-backend%d connectivity failed: %v", i, err)
		}
		le.Infof("Browser connected to backend %d", i+1)
	}

	le.Info("Browser-to-Multiple-Backends test passed")
}

// addPeerWithEcho adds a peer with echo handler.
func addPeerWithEcho(t *testing.T, g *graph.Graph) *graph.Peer {
	ctx := context.Background()
	p, err := graph.GenerateAddPeer(ctx, g)
	if err != nil {
		t.Fatalf("failed to add peer: %v", err)
	}

	proto := protocol.ID("demo/wasm-bridge/v1")

	p.AddFactory(func(b bus.Bus) controller.Factory {
		return stream_echo.NewFactory(b)
	})
	p.AddConfig("echo", &stream_echo.Config{
		ProtocolId: string(proto),
	})

	return p
}

// initSimulator creates a simulator.
func initSimulator(t *testing.T, ctx context.Context, le *logrus.Entry, g *graph.Graph) *simulate.Simulator {
	sim, err := simulate.NewSimulator(ctx, le, g)
	if err != nil {
		t.Fatalf("failed to create simulator: %v", err)
	}
	return sim
}
