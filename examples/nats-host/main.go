package main

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	bcli "github.com/aperturerobotics/bifrost/cli"
	nats "github.com/aperturerobotics/bifrost/pubsub/nats"
	nats_controller "github.com/aperturerobotics/bifrost/pubsub/nats/controller"
	"github.com/aperturerobotics/bifrost/sim/tests"
	"github.com/aperturerobotics/controllerbus/bus"

	"github.com/aperturerobotics/bifrost/pubsub"
	pubsub_relay "github.com/aperturerobotics/bifrost/pubsub/relay"
	"github.com/aperturerobotics/bifrost/sim/graph"
	"github.com/aperturerobotics/bifrost/sim/simulate"
	"github.com/aperturerobotics/controllerbus/controller"
)

var logTrace = false

var addPeer = tests.AddPeer
var initSimulator = tests.InitSimulator

var privKeyPath string

var bDaemonArgs bcli.DaemonArgs

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

func main() {
	app := cli.NewApp()
	app.Name = "nats-host"
	app.Usage = "bifrost with embedded nats router example"
	app.HideVersion = true
	app.Action = runNatsExample
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "priv-key",
			Usage:       "path to private key to use",
			Value:       "priv-key.pem",
			Destination: &privKeyPath,
		},
	}
	app.Flags = append(
		app.Flags,
		(&bDaemonArgs).BuildFlags()...,
	)

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err.Error())
	}
}

func runNatsExample(c *cli.Context) error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	g := graph.NewGraph()

	p0 := addPeer(nil, g)
	p1 := addPeer(nil, g)
	p2 := addPeer(nil, g)

	lan1 := graph.AddLAN(g)
	lan1.AddPeer(g, p0)
	lan1.AddPeer(g, p1) // XXX

	// lan2 := graph.AddLAN(g)
	// lan2.AddPeer(g, p1)
	// lan2.AddPeer(g, p2)
	lan1.AddPeer(g, p2)

	// replicate p2 -> [lan2] -> p1 -> [lan1] -> p0

	// Add pubsub configurations.
	topics := []string{"test-topic-1"}
	for _, peer := range g.AllPeers() {
		peer.AddFactory(func(b bus.Bus) controller.Factory { return nats_controller.NewFactory(b) })
		peer.AddConfig("pubsub", &nats_controller.Config{
			PeerId: peer.GetPeerID().Pretty(),
			NatsConfig: &nats.Config{
				LogTrace: logTrace,
			},
		})
		peer.AddFactory(func(b bus.Bus) controller.Factory { return pubsub_relay.NewFactory(b) })
		peer.AddConfig("pubsub-relay", &pubsub_relay.Config{
			TopicIds: topics,
		})
	}

	sim := initSimulator(nil, ctx, le, g)

	var t *testing.T
	assertConnectivity := func(p0, p1 *graph.Peer) {
		px0 := sim.GetPeerByID(p0.GetPeerID())
		px1 := sim.GetPeerByID(p1.GetPeerID())
		if err := simulate.TestConnectivity(ctx, px0, px1); err != nil {
			le.WithError(err).Error("unsuccessful connectivity test from 0 to 1")
			// t.Fatal(err.Error())
		} else {
			le.Infof(
				"successful connectivity test between %s and %s",
				p0.GetPeerID().Pretty(),
				p1.GetPeerID().Pretty(),
			)
		}
	}
	assertConnectivity(p0, p1)
	assertConnectivity(p1, p2)

	// Attempt to open a channel and communicate.
	testingData := []byte("hello world")
	for _, channelID := range topics {
		lp2 := sim.GetPeerByID(p2.GetPeerID())
		lp2tb := lp2.GetTestbed()
		tpv2, tpv2Ref, err := bus.ExecOneOff(
			ctx,
			lp2tb.Bus,
			pubsub.NewBuildChannelSubscription(channelID, lp2tb.PrivKey),
			true,
			nil,
		)
		if err != nil {
			t.Fatal(err.Error())
		}
		s2 := tpv2.GetValue().(pubsub.BuildChannelSubscriptionValue)
		le.Infof("built channel subscription for channel %s on peer p2", channelID)

		lp0 := sim.GetPeerByID(p0.GetPeerID())
		lp0tb := lp0.GetTestbed()
		tpv0, tpv0Ref, err := bus.ExecOneOff(
			ctx,
			lp0tb.Bus,
			pubsub.NewBuildChannelSubscription(channelID, lp0tb.PrivKey),
			true,
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

		// TODO: remove this delay... needs a little time to "settle"
		<-time.After(time.Millisecond * 200)
		testReplicate := func() {
			le.Infof("publishing data on p2 with peer %s", p2.GetPeerID().Pretty())
			s2.Publish(testingData)
			rmsg := <-msgRx
			if bytes.Compare(rmsg.GetData(), testingData) != 0 {
				t.Fatalf("pubsub data mismatch %v != expected %v", rmsg.GetData(), testingData)
			}
			le.Info("successful pubsub replication from p2 -> [lan2] -> p1 -> [lan1] -> p0 ")
		}
		testReplicate()

		le.Info("interrupting connectivity between p2 and p1")
		for _, l := range lp2.GetTransportController().GetPeerLinks(p1.GetPeerID()) {
			le.Infof("closing link %v", l.GetUUID())
			l.Close()
		}

		// re-connect
		le.Info("expecting re-connect between peers")
		assertConnectivity(p2, p1)

		tpv0Ref.Release()
		tpv2Ref.Release()
	}

	le.Info("tests successful")
	return nil
}
