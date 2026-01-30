//go:build test_examples

package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/aperturerobotics/bifrost/pubsub"
	floodsub_controller "github.com/aperturerobotics/bifrost/pubsub/floodsub/controller"
	"github.com/aperturerobotics/bifrost/sim/graph"
	"github.com/aperturerobotics/bifrost/sim/simulate"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/sirupsen/logrus"
)

// TestPubSubEvents demonstrates topic-based broadcast across a mesh network.
//
// This test creates a mesh of 4 peers and broadcasts events from one peer
// to all others via the pub/sub system. Shows automatic message propagation
// through the network.
//
// Topology:
//
//	  p0
//	 /  \
//	p1--p2
//	 \  /
//	  p3
func TestPubSubEvents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	le := logrus.NewEntry(log)

	// Create mesh topology
	g := graph.NewGraph()

	peers := make([]*graph.Peer, 4)
	for i := range peers {
		peers[i] = addPeerWithPubSub(t, g, "events")
	}

	// Create a fully connected mesh
	lan := graph.AddLAN(g)
	for _, p := range peers {
		lan.AddPeer(g, p)
	}

	// Start simulator
	sim := initSimulator(t, ctx, le, g)
	defer sim.Close()

	// Verify all peers can communicate
	for i := 0; i < len(peers); i++ {
		for j := i + 1; j < len(peers); j++ {
			if err := simulate.TestConnectivity(ctx, sim.GetPeerByID(peers[i].GetPeerID()), sim.GetPeerByID(peers[j].GetPeerID())); err != nil {
				t.Fatalf("connectivity between peer %d and %d failed: %v", i, j, err)
			}
		}
	}
	le.Info("All peers connected in mesh")

	// Subscribe all peers to "events" topic
	topic := "events"
	subscriptions := make([]pubsub.BuildChannelSubscriptionValue, len(peers))
	receivers := make([]chan pubsub.Message, len(peers))

	for i, p := range peers {
		peerInstance := sim.GetPeerByID(p.GetPeerID())
		tb := peerInstance.GetTestbed()

		// Subscribe to topic
		subVal, _, subRef, err := bus.ExecOneOff(
			ctx,
			tb.Bus,
			pubsub.NewBuildChannelSubscription(topic, tb.PrivKey),
			nil,
			nil,
		)
		if err != nil {
			t.Fatalf("peer %d: failed to subscribe: %v", i, err)
		}
		defer subRef.Release()

		subscriptions[i] = subVal.GetValue().(pubsub.BuildChannelSubscriptionValue)

		// Create message receiver channel
		receivers[i] = make(chan pubsub.Message, 10)
		subscriptions[i].AddHandler(func(m pubsub.Message) {
			select {
			case receivers[i] <- m:
			default:
			}
		})
	}

	le.Info("All peers subscribed to topic 'events'")

	// Allow time for subscriptions to propagate
	time.Sleep(100 * time.Millisecond)

	// Publish event from peer 0
	eventData := []byte("Important event from peer 0!")
	le.Infof("Peer 0 publishing event: %s", string(eventData))

	if err := subscriptions[0].Publish(eventData); err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Verify all other peers receive the event
	for i := 1; i < len(peers); i++ {
		select {
		case msg := <-receivers[i]:
			if !bytes.Equal(msg.GetData(), eventData) {
				t.Errorf("peer %d: received wrong data: %s", i, msg.GetData())
			} else {
				le.Infof("Peer %d received event", i)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("peer %d: timeout waiting for event", i)
		}
	}

	le.Info("Pub/Sub event broadcast test passed")
}

// addPeerWithPubSub adds a peer with pub/sub enabled.
func addPeerWithPubSub(t *testing.T, g *graph.Graph, topic string) *graph.Peer {
	ctx := context.Background()
	p, err := graph.GenerateAddPeer(ctx, g)
	if err != nil {
		t.Fatalf("failed to add peer: %v", err)
	}

	// Add floodsub controller
	p.AddFactory(func(b bus.Bus) controller.Factory {
		return floodsub_controller.NewFactory(b)
	})
	p.AddConfig("pubsub", &floodsub_controller.Config{})

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
