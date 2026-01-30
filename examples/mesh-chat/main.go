//go:build !test_examples

// mesh-chat demonstrates decentralized peer-to-peer messaging.
package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/aperturerobotics/bifrost/daemon"
	"github.com/aperturerobotics/bifrost/keypem/keyfile"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	tptc "github.com/aperturerobotics/bifrost/transport/controller"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	"github.com/aperturerobotics/cli"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"
)

// chatProtocol is the protocol ID for chat streams.
var chatProtocol = protocol.ID("demo/mesh-chat/v1")

func main() {
	var listenAddr, dialAddr, keyPath string

	app := cli.NewApp()
	app.Name = "mesh-chat"
	app.Usage = "P2P chat demo"
	app.HideHelpCommand = true
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "listen",
			Usage:       "UDP listen address",
			EnvVars:     []string{"LISTEN"},
			Value:       ":5000",
			Destination: &listenAddr,
		},
		&cli.StringFlag{
			Name:        "dial",
			Usage:       "Peer to dial (peerID@host:port)",
			EnvVars:     []string{"DIAL"},
			Destination: &dialAddr,
		},
		&cli.StringFlag{
			Name:        "key",
			Aliases:     []string{"k"},
			Usage:       "Path to private key file (default: ./peer-{hash}.pem based on listen addr)",
			EnvVars:     []string{"KEY"},
			Destination: &keyPath,
		},
	}
	app.Action = func(c *cli.Context) error {
		return run(listenAddr, dialAddr, keyPath)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// chatHandler handles incoming chat streams and tracks connected peers.
type chatHandler struct {
	// mu guards remote
	mu sync.Mutex
	// remote is the remote peer ID
	remote peer.ID
}

// GetControllerInfo returns information about the controller.
func (h *chatHandler) GetControllerInfo() *controller.Info {
	return controller.NewInfo("mesh-chat/handler", semver.MustParse("0.0.1"), "chat handler")
}

// Execute executes the controller.
func (h *chatHandler) Execute(ctx context.Context) error { return nil }

// Close closes the controller.
func (h *chatHandler) Close() error { return nil }

// HandleDirective handles directives.
func (h *chatHandler) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	d, ok := di.GetDirective().(link.HandleMountedStream)
	if !ok || d.HandleMountedStreamProtocolID() != chatProtocol {
		return nil, nil
	}
	return directive.Resolvers(&chatResolver{h: h, remote: d.HandleMountedStreamRemotePeerID()}), nil
}

// setRemote sets the remote peer ID if not already set.
func (h *chatHandler) setRemote(p peer.ID) {
	h.mu.Lock()
	if h.remote == "" {
		h.remote = p
	}
	h.mu.Unlock()
}

// getRemote returns the remote peer ID.
func (h *chatHandler) getRemote() peer.ID {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.remote
}

// chatResolver resolves HandleMountedStream directives.
type chatResolver struct {
	// h is the chat handler
	h *chatHandler
	// remote is the remote peer ID
	remote peer.ID
}

// Resolve resolves the directive.
func (r *chatResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	r.h.setRemote(r.remote)
	handler.AddValue(link.MountedStreamHandler(r))
	return nil
}

// HandleMountedStream handles an incoming stream by reading and printing messages.
func (r *chatResolver) HandleMountedStream(ctx context.Context, ms link.MountedStream) error {
	buf := make([]byte, 4096)
	for {
		n, err := ms.GetStream().Read(buf)
		if n > 0 {
			fmt.Printf("\r[%s]: %s\n> ", r.remote.String(), string(buf[:n]))
		}
		if err != nil {
			return nil
		}
	}
}

// run starts the chat node.
func run(listenAddr, dialAddr, keyPath string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	le := logrus.NewEntry(log)

	// Determine key path: use provided path or derive from listen address.
	if keyPath == "" {
		hash := sha256.Sum256([]byte(listenAddr))
		hashHex := hex.EncodeToString(hash[:])
		keyPath = fmt.Sprintf("./peer-%s.pem", hashHex[len(hashHex)-4:])
	}

	// Load or generate the private key.
	privKey, err := keyfile.OpenOrWritePrivKey(le, keyPath)
	if err != nil {
		return fmt.Errorf("load/generate key: %w", err)
	}
	peerID, _ := peer.IDFromPrivateKey(privKey)

	fmt.Printf("Peer ID: %s\n", peerID.String())
	fmt.Printf("Key: %s\n", keyPath)
	fmt.Printf("Listening on %s\n", listenAddr)
	if dialAddr == "" {
		fmt.Printf("\nConnect with: %s --listen :5001 --dial %s@localhost%s\n", os.Args[0], peerID.String(), listenAddr)
	}

	// Create the daemon.
	d, err := daemon.NewDaemon(ctx, privKey, daemon.ConstructOpts{LogEntry: le})
	if err != nil {
		return err
	}

	// Add the chat handler to the controller bus.
	b := d.GetControllerBus()
	handler := &chatHandler{}
	if _, err := b.AddController(ctx, handler, nil); err != nil {
		return err
	}

	// Start the UDP transport.
	udpCtrl, _, udpRef, err := loader.WaitExecControllerRunning(ctx, b,
		resolver.NewLoadControllerWithConfig(&udptpt.Config{ListenAddr: listenAddr}), nil)
	if err != nil {
		return err
	}
	defer udpRef.Release()

	tpt, _ := udpCtrl.(*tptc.Controller).GetTransport(ctx)
	udp := tpt.(*udptpt.UDP)

	// Dial the remote peer if specified.
	if dialAddr != "" {
		parts := strings.Split(dialAddr, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid dial format")
		}
		remotePeerID, err := peer.IDB58Decode(parts[0])
		if err != nil {
			return err
		}
		handler.setRemote(remotePeerID)
		fmt.Printf("Dialing %s...\n", parts[1])
		if _, _, err := udp.DialPeer(ctx, remotePeerID, parts[1]); err != nil {
			return err
		}
		fmt.Println("Connected!")
	}

	fmt.Println("\nType messages and press Enter. /quit to exit.")

	// Read input and send messages.
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "/quit" {
			return nil
		}

		remote := handler.getRemote()
		if remote == "" {
			fmt.Println("No peer connected yet.")
			continue
		}

		// Open a stream and send the message.
		ms, rel, err := link.OpenStreamWithPeerEx(ctx, b, chatProtocol, peerID, remote, 0, stream.OpenOpts{})
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			continue
		}

		_, err = ms.GetStream().Write([]byte(line))
		rel()

		if err != nil {
			fmt.Printf("Write failed: %v\n", err)
		}
	}
}
