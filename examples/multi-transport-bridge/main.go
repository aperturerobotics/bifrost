//go:build !test_examples

// multi-transport-bridge demonstrates Bifrost's transport abstraction.
//
// This example shows how the same application protocol works unchanged across
// different underlying transports (UDP, WebSocket, WebRTC).
//
// Run with:
//
//	Terminal 1: go run main.go -transport udp -listen :5000
//	Terminal 2: go run main.go -transport websocket -listen :5001 -dial <peerID>@127.0.0.1:5000
//
// The chat protocol works identically regardless of transport!
//
// Features demonstrated:
//   - Transport abstraction layer
//   - Same protocol code over any transport
//   - Transport switching
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aperturerobotics/bifrost/daemon"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	stream_echo "github.com/aperturerobotics/bifrost/stream/echo"
	bifrosttpt "github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	websocket "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/cli"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

var (
	bridgeProtocol = protocol.ID("demo/bridge/v1")
	log            = logrus.New()
)

func init() {
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	var (
		transportType string
		listenAddr    string
		dialAddr      string
	)

	app := cli.NewApp()
	app.Name = "multi-transport-bridge"
	app.Usage = "Bifrost transport abstraction demo"
	app.HideHelpCommand = true

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "transport",
			Usage:       "Transport type (udp, websocket, webrtc)",
			EnvVars:     []string{"TRANSPORT_TYPE"},
			Value:       "udp",
			Destination: &transportType,
		},
		&cli.StringFlag{
			Name:        "listen",
			Usage:       "Listen address",
			EnvVars:     []string{"LISTEN_ADDR"},
			Value:       ":5000",
			Destination: &listenAddr,
		},
		&cli.StringFlag{
			Name:        "dial",
			Usage:       "Peer to dial (format: peerID@host:port)",
			EnvVars:     []string{"DIAL_ADDR"},
			Destination: &dialAddr,
		},
	}

	app.Action = func(c *cli.Context) error {
		fmt.Printf("Multi-Transport Bridge Demo\n")
		fmt.Printf("   Transport: %s\n\n", transportType)
		return run(transportType, listenAddr, dialAddr)
	}

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func run(transport, listenAddr, dialAddr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	le := logrus.NewEntry(log)

	// Generate peer identity
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	peerID, _ := peer.IDFromPrivateKey(privKey)

	fmt.Printf("Starting bridge node\n")
	fmt.Printf("Peer ID: %s\n", peerID.String())
	fmt.Printf("Transport: %s\n", transport)
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

	// Add protocol handler
	sr.AddFactory(stream_echo.NewFactory(b))
	_, _, echoRef, err := loader.WaitExecControllerRunning(
		ctx, b,
		resolver.NewLoadControllerWithConfig(&stream_echo.Config{
			ProtocolId: string(bridgeProtocol),
		}),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start protocol handler: %w", err)
	}
	defer echoRef.Release()

	// Start transport based on type
	var tptCtrl bifrosttpt.Controller

	switch transport {
	case "udp":
		sr.AddFactory(udptpt.NewFactory(b))
		ctrl, _, ref, err := loader.WaitExecControllerRunning(
			ctx, b,
			resolver.NewLoadControllerWithConfig(&udptpt.Config{
				ListenAddr: listenAddr,
			}),
			nil,
		)
		if err != nil {
			return fmt.Errorf("start UDP transport: %w", err)
		}
		defer ref.Release()
		tptCtrl = ctrl.(bifrosttpt.Controller)

	case "websocket", "ws":
		sr.AddFactory(websocket.NewFactory(b))
		ctrl, _, ref, err := loader.WaitExecControllerRunning(
			ctx, b,
			resolver.NewLoadControllerWithConfig(&websocket.Config{
				ListenAddr: listenAddr,
				HttpPath:   "/ws",
			}),
			nil,
		)
		if err != nil {
			return fmt.Errorf("start WebSocket transport: %w", err)
		}
		defer ref.Release()
		tptCtrl = ctrl.(bifrosttpt.Controller)
		fmt.Printf("   WebSocket endpoint: ws://localhost%s/ws\n", listenAddr)

	case "webrtc":
		return fmt.Errorf("WebRTC transport requires signaling server setup - see webrtc-forwarding example")

	default:
		return fmt.Errorf("unknown transport: %s", transport)
	}

	tpt, err := tptCtrl.GetTransport(ctx)
	if err != nil {
		return fmt.Errorf("get transport: %w", err)
	}
	tptDialer := tpt.(dialer.TransportDialer)

	// Dial if specified
	var remotePeerID peer.ID
	if dialAddr != "" {
		parts := strings.Split(dialAddr, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid dial format")
		}
		remotePeerID, err = peer.IDB58Decode(parts[0])
		if err != nil {
			return fmt.Errorf("parse peer ID: %w", err)
		}
		fmt.Printf("Dialing %s at %s...\n", remotePeerID.String(), parts[1])
		_, _, err = tptDialer.DialPeer(ctx, remotePeerID, parts[1])
		if err != nil {
			return fmt.Errorf("dial peer: %w", err)
		}
		fmt.Printf("Connected to %s\n", remotePeerID.String())
	}

	fmt.Println("\nBridge node running")
	fmt.Println("   The same protocol works identically regardless of transport!")
	fmt.Println("   Try switching transports - zero code changes needed.")

	// If we have a remote peer, start interactive chat
	if remotePeerID != "" {
		return chatLoop(ctx, b, peerID, remotePeerID)
	}

	// Otherwise, just wait
	fmt.Println("Waiting for incoming connections...")
	<-ctx.Done()
	return nil
}

func chatLoop(ctx context.Context, b bus.Bus, localPeer, remotePeer peer.ID) error {
	fmt.Println("\nChat session started")
	fmt.Println("   Type /quit to exit")
	fmt.Println()

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

		// Send message via stream
		if err := sendChatMessage(ctx, b, localPeer, remotePeer, line); err != nil {
			fmt.Printf("Failed to send: %v\n", err)
		} else {
			fmt.Printf("Sent: %s\n", line)
		}
	}
}

func sendChatMessage(ctx context.Context, b bus.Bus, localPeer, remotePeer peer.ID, msg string) error {

	ms, msRel, err := link.OpenStreamWithPeerEx(
		ctx,
		b,
		bridgeProtocol,
		localPeer,
		remotePeer,
		0,
		stream.OpenOpts{},
	)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	defer msRel()

	_, err = ms.GetStream().Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}

	// Read echo response
	respBuf := make([]byte, len(msg)+100)
	ms.GetStream().SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _ = ms.GetStream().Read(respBuf)

	return nil
}
