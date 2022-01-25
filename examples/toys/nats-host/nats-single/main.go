package main

import (
	"context"
	"os"

	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	bcli "github.com/aperturerobotics/bifrost/cli"
	nats "github.com/aperturerobotics/bifrost/pubsub/nats"
	nats_controller "github.com/aperturerobotics/bifrost/pubsub/nats/controller"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/bifrost/sim/tests"

	"github.com/aperturerobotics/bifrost/pubsub"
	pubsub_relay "github.com/aperturerobotics/bifrost/pubsub/relay"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/bifrost/sim/graph"
)

var addPeer = tests.AddPeer
var initSimulator = tests.InitSimulator

var privKeyPath string

var bDaemonArgs bcli.DaemonArgs

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

func main() {
	app := cli.NewApp()
	app.Name = "nats-single"
	app.Usage = "bifrost with embedded nats client example"
	app.HideVersion = true
	app.Action = runNatsExample
	app.Flags = []cli.Flag{
		cli.StringFlag{
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
	lan1 := graph.AddLAN(g)
	lan1.AddPeer(g, p0)

	// Add pubsub configurations.
	channelID := "test-topic-1"
	for _, peer := range g.AllPeers() {
		peer.AddFactory(func(b bus.Bus) controller.Factory { return nats_controller.NewFactory(b) })
		peer.AddConfig("pubsub", &nats_controller.Config{
			PeerId: peer.GetPeerID().Pretty(),
			NatsConfig: &nats.Config{
				LogTrace: true,
			},
		})
		peer.AddFactory(func(b bus.Bus) controller.Factory { return pubsub_relay.NewFactory(b) })
		peer.AddConfig("pubsub-relay", &pubsub_relay.Config{
			TopicIds: []string{channelID},
		})
	}

	sim := initSimulator(nil, ctx, le, g)

	lp0 := sim.GetPeerByID(p0.GetPeerID())
	lp0tb := lp0.GetTestbed()
	tpv0, tpv0Ref, err := bus.ExecOneOff(
		ctx,
		lp0tb.Bus,
		pubsub.NewBuildChannelSubscription(channelID, lp0tb.PrivKey),
		nil,
	)
	if err != nil {
		return err
	}
	defer tpv0Ref.Release()
	s0 := tpv0.GetValue().(pubsub.BuildChannelSubscriptionValue)
	le.Infof("built channel subscription for channel %s on peer p0", channelID)

	if err := s0.Publish([]byte("testing 1234 HELLo WORLD")); err != nil {
		return err
	}

	le.Info("tests successful")
	<-ctx.Done()
	_ = s0
	return nil
}
