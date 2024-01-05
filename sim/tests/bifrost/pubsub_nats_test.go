package bifrost

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/aperturerobotics/bifrost/pubsub/nats"
	nats_controller "github.com/aperturerobotics/bifrost/pubsub/nats/controller"
	"github.com/aperturerobotics/bifrost/sim/graph"
	"github.com/aperturerobotics/bifrost/sim/simulate"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/sirupsen/logrus"
)

var logTrace = false

// TestPubsubNATS performs a simple pubsub test using nats protocol.
func TestPubsubNATS(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	g := graph.NewGraph()

	p0 := addPeer(t, g)
	p1 := addPeer(t, g)

	lan1 := graph.AddLAN(g)
	lan1.AddPeer(g, p0)
	lan1.AddPeer(g, p1)

	// Add pubsub configurations.
	topics := []string{"test-topic-1"}
	for _, peer := range g.AllPeers() {
		peer.AddFactory(func(b bus.Bus) controller.Factory { return nats_controller.NewFactory(b) })
		peer.AddConfig("pubsub", &nats_controller.Config{
			PeerId: peer.GetPeerID().String(),
			NatsConfig: &nats.Config{
				LogTrace: logTrace,
			},
		})
		/*
			peer.AddFactory(func(b bus.Bus) controller.Factory { return pubsub_relay.NewFactory(b) })
			peer.AddConfig("pubsub-relay", &pubsub_relay.Config{
				PeerId:   peer.GetPeerID().String(),
				TopicIds: topics,
			})
		*/
	}

	sim := initSimulator(t, ctx, le, g)

	assertConnectivity := func(p0, p1 *graph.Peer) {
		px0 := sim.GetPeerByID(p0.GetPeerID())
		px1 := sim.GetPeerByID(p1.GetPeerID())
		if err := simulate.TestConnectivity(ctx, px0, px1); err != nil {
			t.Fatal(err.Error())
		}
		le.Infof(
			"successful connectivity test between %s and %s",
			p0.GetPeerID().String(),
			p1.GetPeerID().String(),
		)
	}
	assertConnectivity(p0, p1)

	// Attempt to open a channel and communicate.
	testingData := []byte("hello world")
	for _, channelID := range topics {
		lp1 := sim.GetPeerByID(p1.GetPeerID())
		lp1tb := lp1.GetTestbed()
		tpv1, _, tpv1Ref, err := bus.ExecOneOff(
			ctx,
			lp1tb.Bus,
			pubsub.NewBuildChannelSubscription(channelID, lp1tb.PrivKey),
			nil,
			nil,
		)
		if err != nil {
			t.Fatal(err.Error())
		}
		s1 := tpv1.GetValue().(pubsub.BuildChannelSubscriptionValue)
		le.Infof("built channel subscription for channel %s on peer p1", channelID)

		lp0 := sim.GetPeerByID(p0.GetPeerID())
		lp0tb := lp0.GetTestbed()
		tpv0, _, tpv0Ref, err := bus.ExecOneOff(
			ctx,
			lp0tb.Bus,
			pubsub.NewBuildChannelSubscription(channelID, lp0tb.PrivKey),
			nil,
			nil,
		)
		if err != nil {
			t.Fatal(err.Error())
		}
		s0 := tpv0.GetValue().(pubsub.BuildChannelSubscriptionValue)
		le.Infof("built channel subscription for channel %s on peer p0", channelID)

		msgRx := make(chan pubsub.Message, 1)
		s0.AddHandler(func(m pubsub.Message) {
			select {
			case msgRx <- m:
			default:
			}
		})

		// allow things to "settle"
		<-time.After(time.Millisecond * 100)

		testReplicate := func() {
			ctx, ctxCancel := context.WithCancel(ctx)
			defer ctxCancel()

			errCh := make(chan error, 1)
			go func() {
				le.Infof("publishing data on p1 with peer %s", p1.GetPeerID().String())
				for {
					select {
					case <-ctx.Done():
						return
					case <-time.After(time.Second):
						err := s1.Publish(testingData)
						if err != nil {
							errCh <- err
						}
					}
				}
			}()

			select {
			case rmsg := <-msgRx:
				if !bytes.Equal(rmsg.GetData(), testingData) {
					t.Fatalf("pubsub data mismatch %v != expected %v", rmsg.GetData(), testingData)
				}
			case err := <-errCh:
				t.Fatal(err.Error())
			}

			le.Info("successful pubsub replication from p2 -> [lan2] -> p1")
		}

		testReplicate()

		tpv0Ref.Release()
		tpv1Ref.Release()
	}

	le.Info("tests successful")
}
