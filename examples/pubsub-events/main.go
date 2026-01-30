//go:build !test_examples

// pubsub-events demonstrates real-time event broadcasting across a mesh network.
//
// This example shows how Bifrost's pub/sub system enables event-driven architectures
// over peer-to-peer networks. Events automatically propagate to all interested peers.
//
// Run with:
//
//	Terminal 1: go run main.go -listen :5000 -topic "events"
//	Terminal 2: go run main.go -listen :5001 -topic "events" -dial <peer1>@127.0.0.1:5000
//
// Features demonstrated:
//   - Topic-based message broadcasting
//   - Automatic event propagation through mesh
//   - Decentralized pub/sub without brokers
//
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aperturerobotics/bifrost/daemon"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
	floodsub_controller "github.com/aperturerobotics/bifrost/pubsub/floodsub/controller"
	pubsub_relay "github.com/aperturerobotics/bifrost/pubsub/relay"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	"github.com/aperturerobotics/cli"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	var (
		listenAddr string
		topic      string
		dialAddr   string
	)

	app := cli.NewApp()
	app.Name = "pubsub-events"
	app.Usage = "Real-time event broadcasting across a mesh network demo"
	app.HideHelpCommand = true

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "listen",
			Usage:       "UDP listen address",
			EnvVars:     []string{"LISTEN_ADDR"},
			Value:       ":5000",
			Destination: &listenAddr,
		},
		&cli.StringFlag{
			Name:        "topic",
			Usage:       "Topic to subscribe/publish",
			EnvVars:     []string{"TOPIC"},
			Value:       "events",
			Destination: &topic,
		},
		&cli.StringFlag{
			Name:        "dial",
			Usage:       "Peer to connect to (format: peerID@host:port)",
			EnvVars:     []string{"DIAL_ADDR"},
			Destination: &dialAddr,
		},
	}

	app.Action = func(c *cli.Context) error {
		return run(listenAddr, topic, dialAddr)
	}

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func run(listenAddr, topic, dialAddr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	le := logrus.NewEntry(log)

	// Generate peer identity
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	peerID, _ := peer.IDFromPrivateKey(privKey)

	fmt.Printf("Pub/Sub Events Demo\n")
	fmt.Printf("Peer ID: %s\n", peerID.String())
	fmt.Printf("Topic: %s\n", topic)
	fmt.Printf("Listening on %s\n\n", listenAddr)

	// Create daemon
	d, err := daemon.NewDaemon(ctx, privKey, daemon.ConstructOpts{
		LogEntry: le,
	})
	if err != nil {
		return fmt.Errorf("create daemon: %w", err)
	}

	b := d.GetControllerBus()
	sr := d.GetStaticResolver()

	// Add pubsub controllers
	sr.AddFactory(floodsub_controller.NewFactory(b))
	pubsubCtrl, _, pubsubRef, err := loader.WaitExecControllerRunning(
		ctx, b,
		resolver.NewLoadControllerWithConfig(&floodsub_controller.Config{}),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start pubsub: %w", err)
	}
	defer pubsubRef.Release()

	// Add relay controller to subscribe to topics
	sr.AddFactory(pubsub_relay.NewFactory(b))
	_, _, relayRef, err := loader.WaitExecControllerRunning(
		ctx, b,
		resolver.NewLoadControllerWithConfig(&pubsub_relay.Config{
			PeerId:   peerID.String(),
			TopicIds: []string{topic},
		}),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start relay: %w", err)
	}
	defer relayRef.Release()

	// Start UDP transport
	sr.AddFactory(udptpt.NewFactory(b))
	udpCtrl, _, udpRef, err := loader.WaitExecControllerRunning(
		ctx, b,
		resolver.NewLoadControllerWithConfig(&udptpt.Config{
			ListenAddr: listenAddr,
		}),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start UDP: %w", err)
	}
	defer udpRef.Release()

	// Get transport for dialing
	tptCtrl := udpCtrl.(transport.Controller)
	tpt, err := tptCtrl.GetTransport(ctx)
	if err != nil {
		return fmt.Errorf("get transport: %w", err)
	}
	tptDialer := tpt.(dialer.TransportDialer)

	// Get the pubsub controller
	floodsub := pubsubCtrl.(interface {
		GetPubSub(ctx context.Context) (pubsub.PubSub, error)
	})

	// Get pubsub instance
	ps, err := floodsub.GetPubSub(ctx)
	if err != nil {
		return fmt.Errorf("get pubsub: %w", err)
	}

	// Create a subscription to the topic
	subscription, err := ps.AddSubscription(ctx, privKey, topic)
	if err != nil {
		return fmt.Errorf("add subscription: %w", err)
	}
	defer subscription.Release()

	// Add handler to receive messages
	subscription.AddHandler(func(m pubsub.Message) {
		fmt.Printf("\nReceived: %s\n> ", string(m.GetData()))
	})

	fmt.Printf("Subscribed to topic: %s\n", topic)
	fmt.Println("   You can now publish and receive events on this topic")
	fmt.Println()

	// Dial peer if specified
	if dialAddr != "" {
		parts := strings.Split(dialAddr, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid dial format")
		}
		remotePeerID, err := peer.IDB58Decode(parts[0])
		if err != nil {
			return fmt.Errorf("parse peer ID: %w", err)
		}
		fmt.Printf("Connecting to %s at %s...\n", parts[0], parts[1])
		_, _, err = tptDialer.DialPeer(ctx, remotePeerID, parts[1])
		if err != nil {
			fmt.Printf("Failed to dial: %v\n", err)
		} else {
			fmt.Printf("Connected to %s\n\n", parts[0])
		}
	}

	fmt.Println("Type messages and press Enter to broadcast")
	fmt.Println("   Messages will propagate to all peers subscribed to the topic")
	fmt.Println()

	// Read input and publish
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "/quit" {
			fmt.Println("Goodbye!")
			return nil
		}

		// Publish event
		fmt.Printf("Broadcasting: %s\n", line)
		if err := subscription.Publish([]byte(line)); err != nil {
			fmt.Printf("Failed to publish: %v\n", err)
		} else {
			fmt.Println("Published successfully")
		}
	}
}
