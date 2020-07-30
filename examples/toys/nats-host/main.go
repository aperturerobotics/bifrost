package main

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	bcli "github.com/aperturerobotics/bifrost/cli"
	bcore "github.com/aperturerobotics/bifrost/core"
	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/bifrost/keypem/keyfile"
	"github.com/aperturerobotics/bifrost/peer"
	peer_controller "github.com/aperturerobotics/bifrost/peer/controller"
	nats_controller "github.com/aperturerobotics/bifrost/pubsub/nats/controller"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/hydra/core"
	"github.com/aperturerobotics/network-sim/tests"

	// transform_snappy "github.com/aperturerobotics/hydra/block/transform/snappy"
	"github.com/aperturerobotics/bifrost/pubsub"
	pubsub_relay "github.com/aperturerobotics/bifrost/pubsub/relay"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/network-sim/graph"
	"github.com/aperturerobotics/network-sim/simulate"
)

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

// buildNatsNode builds a new nats node
func buildNatsNode(
	ctx context.Context,
	le *logrus.Entry,
	privKeyPath string,
) (bus.Bus, crypto.PrivKey, directive.Reference, error) {
	var privKey crypto.PrivKey
	var err error
	if privKeyPath != "" {
		le.WithField("priv-key-path", privKeyPath).Debug("opening private key")
		privKey, err = keyfile.OpenOrWritePrivKey(le, privKeyPath)
	} else {
		le.Debug("generating private key")
		privKey, _, err = keypem.GeneratePrivKey()
	}
	if err != nil {
		return nil, nil, nil, err
	}

	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return nil, nil, nil, err
	}
	sr.AddFactory(nats_controller.NewFactory(b))
	bcore.AddFactories(b, sr)

	pid, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, nil, nil, err
	}
	le.WithField("peer-id", pid.Pretty()).Info("starting with peer id")

	peerCtrl, err := peer_controller.NewController(le, privKey)
	if err != nil {
		return nil, nil, nil, err
	}
	go b.ExecuteController(ctx, peerCtrl)

	// execute the pubsub controller
	_, pubsubRef, err := bus.ExecOneOff(
		ctx,
		b,
		resolver.NewLoadControllerWithConfig(&nats_controller.Config{
			PeerId: pid.Pretty(),
		}),
		nil,
	)
	if err != nil {
		return nil, nil, nil, err
	}
	// defer pubsubRef.Release()
	return b, privKey, pubsubRef, nil
}

func runNatsExample(c *cli.Context) error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	/*
		peers := make([]bus.Bus, 3)
		peerPrivs := make([]crypto.PrivKey, len(peers))
		for i := range peers {
			var peerRef directive.Reference
			var err error
			peers[i], peerPrivs[i], peerRef, err = buildNatsNode(
				ctx,
				le.WithField("peer", i),
				fmt.Sprintf("peer-%d.pem", i),
			)
			if err != nil {
				return err
			}
			defer peerRef.Release()
		}
	*/

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
			pubsub.NewBuildChannelSubscription(channelID),
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
			pubsub.NewBuildChannelSubscription(channelID),
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
		<-time.After(time.Millisecond * 100)
		testReplicate := func() {
			le.Infof("publishing data on p2 with peer %s", p2.GetPeerID().Pretty())
			s2.Publish(p2.GetPeerPriv(), testingData)
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
