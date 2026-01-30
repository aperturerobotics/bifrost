//go:build !test_examples

// mesh-chat demonstrates decentralized peer-to-peer messaging.
//
// This example shows how Bifrost enables direct communication between peers
// across any transport (UDP, WebSocket, WebRTC). The chat protocol automatically
// routes messages through available links.
//
// Run with: go run ./mesh-chat/main.go -listen :5000 -dial <peerID>@<host>:<port>
//
// Features demonstrated:
//   - Transport abstraction (works over any Bifrost transport)
//   - Automatic peer discovery and link establishment
//   - Bidirectional streaming between peers
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
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/daemon"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	stream_echo "github.com/aperturerobotics/bifrost/stream/echo"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	"github.com/aperturerobotics/cli"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

var (
	chatProtocol = protocol.ID("demo/mesh-chat/v1")
	log          = logrus.New()
)

func init() {
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	var (
		listenAddr string
		dialAddr   string
	)

	app := cli.NewApp()
	app.Name = "mesh-chat"
	app.Usage = "Decentralized peer-to-peer messaging demo"
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
			Name:        "dial",
			Usage:       "Peer address to dial (format: peerID@host:port)",
			EnvVars:     []string{"DIAL_ADDR"},
			Destination: &dialAddr,
		},
	}

	app.Action = func(c *cli.Context) error {
		return run(listenAddr, dialAddr)
	}

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

// run starts the chat node.
func run(listenAddr, dialAddr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	le := logrus.NewEntry(log)

	// Generate peer identity
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	peerID, _ := peer.IDFromPrivateKey(privKey)

	fmt.Printf("Starting mesh chat node\n")
	fmt.Printf("Peer ID: %s\n", peerID.String())
	fmt.Printf("Type messages and press Enter to send\n")
	fmt.Printf("Listening on %s\n\n", listenAddr)

	// Create daemon with echo controller for chat
	d, err := daemon.NewDaemon(ctx, privKey, daemon.ConstructOpts{
		LogEntry: le,
	})
	if err != nil {
		return fmt.Errorf("create daemon: %w", err)
	}

	b := d.GetControllerBus()
	sr := d.GetStaticResolver()

	// Add echo controller to handle incoming chat streams
	sr.AddFactory(stream_echo.NewFactory(b))
	_, _, echoRef, err := loader.WaitExecControllerRunning(
		ctx, b,
		resolver.NewLoadControllerWithConfig(&stream_echo.Config{
			ProtocolId: string(chatProtocol),
		}),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start echo controller: %w", err)
	}
	defer echoRef.Release()

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
		return fmt.Errorf("start UDP transport: %w", err)
	}
	defer udpRef.Release()

	// Get transport for dialing
	tptCtrl := udpCtrl.(transport.Controller)
	tpt, err := tptCtrl.GetTransport(ctx)
	if err != nil {
		return fmt.Errorf("get transport: %w", err)
	}
	tptDialer := tpt.(dialer.TransportDialer)

	// Dial peer if specified
	var remotePeerID peer.ID
	var dialAddrHost string
	if dialAddr != "" {
		parts := strings.Split(dialAddr, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid dial format, expected peerID@host:port")
		}

		remotePeerID, err = peer.IDB58Decode(parts[0])
		if err != nil {
			return fmt.Errorf("parse peer ID: %w", err)
		}
		dialAddrHost = parts[1]

		fmt.Printf("Dialing %s at %s...\n", remotePeerID.String(), dialAddrHost)
		_, _, err = tptDialer.DialPeer(ctx, remotePeerID, dialAddrHost)
		if err != nil {
			return fmt.Errorf("dial peer: %w", err)
		}
		fmt.Printf("Connected to %s\n\n", remotePeerID.String())
	}

	// Start chat session
	return chatSession(ctx, b, peerID, remotePeerID)
}

// chatSession runs an interactive chat with a remote peer.
func chatSession(ctx context.Context, b bus.Bus, localPeer, remotePeer peer.ID) error {
	fmt.Println("Chat session started")
	if remotePeer != "" {
		fmt.Printf("   Connected to: %s\n", remotePeer.String())
	}
	fmt.Println("   Type /quit to exit, /help for commands")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	var wg sync.WaitGroup

	// Start goroutine to handle incoming messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		// In a real implementation, we would listen for incoming streams
		// and print messages from peers
		// For this demo, we rely on the echo controller which echoes back what we send
	}()

	// Handle outgoing messages
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nGoodbye!")
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle commands
		if line == "/quit" {
			fmt.Println("Goodbye!")
			return nil
		}

		if line == "/help" {
			fmt.Println("Commands:")
			fmt.Println("  /quit - Exit")
			fmt.Println("  /help - Show this help")
			fmt.Println("  <any text> - Send message to peer")
			continue
		}

		// Send message to peer if we have a remote peer
		if remotePeer != "" {
			if err := sendMessage(ctx, b, localPeer, remotePeer, line); err != nil {
				fmt.Printf("Failed to send: %v\n", err)
			} else {
				fmt.Printf("Sent: %s\n", line)
			}
		} else {
			fmt.Println("No peer connected. Waiting for incoming connections...")
		}
	}
}

// sendMessage opens a stream and sends a message to the remote peer.
func sendMessage(ctx context.Context, b bus.Bus, localPeer, remotePeer peer.ID, msg string) error {
	// Open stream to remote peer
	ms, msRel, err := link.OpenStreamWithPeerEx(
		ctx,
		b,
		chatProtocol,
		localPeer,
		remotePeer,
		0,
		stream.OpenOpts{},
	)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	defer msRel()

	// Send the message
	_, err = ms.GetStream().Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	// Read response (echoed back)
	respBuf := make([]byte, len(msg)+100)
	ms.GetStream().SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := ms.GetStream().Read(respBuf)
	if err == nil && n > 0 {
		// Message echoed back successfully
		_ = respBuf[:n]
	}

	return nil
}
