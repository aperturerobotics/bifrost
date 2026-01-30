//go:build !test_examples

// wasm-browser-bridge demonstrates browser-to-native communication.
//
// This example shows how Bifrost enables a web browser (running WebAssembly)
// to communicate directly with a native Go backend over WebRTC or WebSocket.
//
// Architecture:
//
//	Browser (WASM) <--WebRTC/WebSocket--> Native Server (Go)
//
// The browser can:
//   - Make HTTP requests through the WebRTC tunnel
//   - Stream data bidirectionally
//   - Connect to multiple backend services
//
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aperturerobotics/bifrost/daemon"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	stream_echo "github.com/aperturerobotics/bifrost/stream/echo"
	"github.com/aperturerobotics/bifrost/transport"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	"github.com/aperturerobotics/cli"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

var (
	bridgeProtocol = protocol.ID("demo/wasm-bridge/v1")
	log            = logrus.New()
)

func init() {
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	var (
		listenAddr string
		httpAddr   string
	)

	app := cli.NewApp()
	app.Name = "wasm-browser-bridge"
	app.Usage = "Browser-to-native communication demo"
	app.HideHelpCommand = true

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "listen",
			Usage:       "Listen address",
			EnvVars:     []string{"LISTEN_ADDR"},
			Value:       ":5000",
			Destination: &listenAddr,
		},
		&cli.StringFlag{
			Name:        "http",
			Usage:       "HTTP service to expose",
			EnvVars:     []string{"HTTP_ADDR"},
			Value:       ":8080",
			Destination: &httpAddr,
		},
	}

	app.Action = func(c *cli.Context) error {
		return run(listenAddr, httpAddr)
	}

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func run(listenAddr, httpAddr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	le := logrus.NewEntry(log)

	// Generate peer identity
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	peerID, _ := peer.IDFromPrivateKey(privKey)

	fmt.Printf("WASM Browser Bridge Demo\n")
	fmt.Printf("Backend Peer ID: %s\n", peerID.String())
	fmt.Printf("Listening on %s\n", listenAddr)
	fmt.Printf("HTTP Service at %s\n\n", httpAddr)

	// Create daemon
	d, err := daemon.NewDaemon(ctx, privKey, daemon.ConstructOpts{
		LogEntry: le,
	})
	if err != nil {
		return fmt.Errorf("create daemon: %w", err)
	}

	b := d.GetControllerBus()
	sr := d.GetStaticResolver()

	// Add echo handler (represents API endpoint)
	sr.AddFactory(stream_echo.NewFactory(b))
	_, _, echoRef, err := loader.WaitExecControllerRunning(
		ctx, b,
		resolver.NewLoadControllerWithConfig(&stream_echo.Config{
			ProtocolId: string(bridgeProtocol),
		}),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start echo handler: %w", err)
	}
	defer echoRef.Release()

	// Start UDP transport (accepts WebRTC or WebSocket upgrade)
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

	// Get transport controller
	tptCtrl := udpCtrl.(transport.Controller)
	_, err = tptCtrl.GetTransport(ctx)
	if err != nil {
		return fmt.Errorf("get transport: %w", err)
	}

	// Start a simple HTTP service that also bridges to Bifrost
	go func() {
		mux := http.NewServeMux()

		// Standard HTTP endpoint
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello from Bifrost WASM Bridge!\nPath: %s\nPeerID: %s\n",
				r.URL.Path, peerID.String())
		})

		// Bridge endpoint - forwards HTTP requests over Bifrost stream
		mux.HandleFunc("/bridge/", func(w http.ResponseWriter, r *http.Request) {
			// Parse target peer from query parameter
			targetPeer := r.URL.Query().Get("peer")
			if targetPeer == "" {
				http.Error(w, "Missing 'peer' query parameter", http.StatusBadRequest)
				return
			}

			remotePeerID, err := peer.IDB58Decode(targetPeer)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid peer ID: %v", err), http.StatusBadRequest)
				return
			}

			// Open stream to peer
			streamCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			ms, msRel, err := link.OpenStreamWithPeerEx(
				streamCtx,
				b,
				bridgeProtocol,
				peerID,
				remotePeerID,
				0,
				stream.OpenOpts{},
			)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to open stream: %v", err), http.StatusServiceUnavailable)
				return
			}
			defer msRel()

			// Read request body
			body, _ := io.ReadAll(r.Body)

			// Forward HTTP request over Bifrost stream
			reqMsg := fmt.Sprintf("%s %s HTTP/1.1\r\nHost: %s\r\n\r\n%s",
				r.Method, r.URL.Path, r.Host, string(body))

			if _, err := ms.GetStream().Write([]byte(reqMsg)); err != nil {
				http.Error(w, fmt.Sprintf("Failed to write to stream: %v", err), http.StatusInternalServerError)
				return
			}

			// Read response from peer
			respBuf := make([]byte, 4096)
			ms.GetStream().SetReadDeadline(time.Now().Add(5 * time.Second))
			n, err := ms.GetStream().Read(respBuf)
			if err != nil && err != io.EOF {
				http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
				return
			}

			// Return response
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write(respBuf[:n])
		})

		// Info endpoint for browser to get PeerID
		mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"peer_id": "%s", "protocol": "%s"}`, peerID.String(), bridgeProtocol)
		})

		fmt.Printf("HTTP service running at http://localhost%s\n", httpAddr)
		if err := http.ListenAndServe(httpAddr, mux); err != nil {
			le.WithError(err).Error("HTTP server error")
		}
	}()

	fmt.Println("\nBridge server ready")
	fmt.Println("   Browser can now connect via WebRTC/WebSocket")
	fmt.Println("   and access the HTTP service through the encrypted tunnel")
	fmt.Printf("\n   Share this Peer ID with the browser: %s\n", peerID.String())
	fmt.Println("\n   Endpoints:")
	fmt.Printf("     - http://localhost%s/         - Status page\n", httpAddr)
	fmt.Printf("     - http://localhost%s/info     - Get Peer ID (JSON)\n", httpAddr)
	fmt.Printf("     - http://localhost%s/bridge/  - Bridge to peer (use ?peer=<peer_id>)\n", httpAddr)

	<-ctx.Done()
	return nil
}
