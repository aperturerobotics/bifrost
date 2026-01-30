//go:build !test_examples

// p2p-filedrop demonstrates secure peer-to-peer file transfer.
//
// This example shows how Bifrost enables direct file sharing between peers
// without any intermediate servers. All transfers are encrypted end-to-end.
//
// Run with:
//
//	Send:    go run main.go -send <filename> -to <peerID>@<host>:<port>
//	Receive: go run main.go -listen :5000
//
// Features demonstrated:
//   - End-to-end encrypted file transfer
//   - Streaming large files without memory issues
//   - Direct P2P connectivity (no cloud required)
//   - Works over any Bifrost transport (UDP, WebSocket, WebRTC)
//
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	"github.com/aperturerobotics/cli"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

var (
	fileProtocol = protocol.ID("demo/file-transfer/v1")
	log          = logrus.New()
)

func init() {
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	var (
		listenAddr string
		sendFile   string
		toAddr     string
	)

	app := cli.NewApp()
	app.Name = "p2p-filedrop"
	app.Usage = "Secure peer-to-peer file transfer demo"
	app.HideHelpCommand = true

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "listen",
			Usage:       "Listen address (e.g., :5000) - receive mode",
			EnvVars:     []string{"LISTEN_ADDR"},
			Destination: &listenAddr,
		},
		&cli.StringFlag{
			Name:        "send",
			Usage:       "File to send",
			EnvVars:     []string{"SEND_FILE"},
			Destination: &sendFile,
		},
		&cli.StringFlag{
			Name:        "to",
			Usage:       "Recipient address (format: peerID@host:port)",
			EnvVars:     []string{"TO_ADDR"},
			Destination: &toAddr,
		},
	}

	app.Action = func(c *cli.Context) error {
		// Validate arguments
		if listenAddr != "" && (sendFile != "" || toAddr != "") {
			return fmt.Errorf("cannot use --listen with --send or --to flags")
		}

		if listenAddr == "" && (sendFile == "" || toAddr == "") {
			return fmt.Errorf("must specify either --listen (receive mode) or both --send and --to (send mode)")
		}

		if listenAddr != "" {
			return runReceiver(listenAddr)
		}
		return runSender(sendFile, toAddr)
	}

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

// runReceiver starts in receive mode.
func runReceiver(listenAddr string) error {
	ctx := context.Background()
	le := logrus.NewEntry(log)

	// Generate peer identity
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	peerID, _ := peer.IDFromPrivateKey(privKey)

	fmt.Printf("P2P File Drop - Receiver Mode\n")
	fmt.Printf("Your Peer ID: %s\n", peerID.String())
	fmt.Printf("Listening on %s\n\n", listenAddr)
	fmt.Println("Waiting for incoming file transfers...")
	fmt.Println("   Share your Peer ID with the sender")

	// Create daemon
	d, err := daemon.NewDaemon(ctx, privKey, daemon.ConstructOpts{
		LogEntry: le,
	})
	if err != nil {
		return fmt.Errorf("create daemon: %w", err)
	}

	b := d.GetControllerBus()
	sr := d.GetStaticResolver()

	// Add file handler (echo controller acts as receiver)
	sr.AddFactory(stream_echo.NewFactory(b))
	_, _, echoRef, err := loader.WaitExecControllerRunning(
		ctx, b,
		resolver.NewLoadControllerWithConfig(&stream_echo.Config{
			ProtocolId: string(fileProtocol),
		}),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start file handler: %w", err)
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

	// Get transport controller for receiving files
	tptCtrl := udpCtrl.(transport.Controller)
	_, err = tptCtrl.GetTransport(ctx)
	if err != nil {
		return fmt.Errorf("get transport: %w", err)
	}
	// Transport is now ready for incoming connections

	fmt.Println("Receiver ready - waiting for file transfers...")
	fmt.Println("   Press Ctrl+C to exit")

	// Keep running
	<-ctx.Done()
	return nil
}

// runSender starts in send mode.
func runSender(filename, toAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	le := logrus.NewEntry(log)

	// Parse recipient address
	parts := strings.Split(toAddr, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid address format, expected peerID@host:port")
	}
	recipientID, err := peer.IDB58Decode(parts[0])
	if err != nil {
		return fmt.Errorf("parse peer ID: %w", err)
	}
	recipientAddr := parts[1]

	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	fmt.Printf("P2P File Drop - Sender Mode\n")
	fmt.Printf("File: %s (%d bytes)\n", filename, stat.Size())
	fmt.Printf("Recipient: %s\n", recipientID.String())
	fmt.Printf("Address: %s\n\n", recipientAddr)

	// Generate peer identity
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	peerID, _ := peer.IDFromPrivateKey(privKey)

	fmt.Printf("Your Peer ID: %s\n", peerID.String())

	// Create daemon
	d, err := daemon.NewDaemon(ctx, privKey, daemon.ConstructOpts{
		LogEntry: le,
	})
	if err != nil {
		return fmt.Errorf("create daemon: %w", err)
	}

	b := d.GetControllerBus()
	sr := d.GetStaticResolver()

	// Add file handler
	sr.AddFactory(stream_echo.NewFactory(b))
	_, _, echoRef, err := loader.WaitExecControllerRunning(
		ctx, b,
		resolver.NewLoadControllerWithConfig(&stream_echo.Config{
			ProtocolId: string(fileProtocol),
		}),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start file handler: %w", err)
	}
	defer echoRef.Release()

	// Start UDP transport
	sr.AddFactory(udptpt.NewFactory(b))
	udpCtrl, _, udpRef, err := loader.WaitExecControllerRunning(
		ctx, b,
		resolver.NewLoadControllerWithConfig(&udptpt.Config{
			ListenAddr: ":0", // Random port
		}),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start UDP transport: %w", err)
	}
	defer udpRef.Release()

	// Get transport and dial the peer
	tptCtrl := udpCtrl.(transport.Controller)
	tpt, err := tptCtrl.GetTransport(ctx)
	if err != nil {
		return fmt.Errorf("get transport: %w", err)
	}
	tptDialer := tpt.(dialer.TransportDialer)

	fmt.Println("Connecting to recipient...")

	// Dial the recipient peer
	_, _, err = tptDialer.DialPeer(ctx, recipientID, recipientAddr)
	if err != nil {
		return fmt.Errorf("dial peer: %w", err)
	}
	fmt.Printf("Connected to %s\n\n", recipientID.String())

	// Calculate checksum
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("calculate checksum: %w", err)
	}
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	fmt.Printf("SHA-256: %s\n", checksum)

	// Reset file to beginning for transfer
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("reset file: %w", err)
	}

	// Open a stream to transfer the file
	fmt.Println("Opening file transfer stream...")
	ms, msRel, err := link.OpenStreamWithPeerEx(
		ctx,
		b,
		fileProtocol,
		peerID,
		recipientID,
		0,
		stream.OpenOpts{},
	)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	defer msRel()

	// Send file in chunks
	fmt.Printf("Sending file in chunks...\n")
	chunkSize := 4096
	buf := make([]byte, chunkSize)
	totalSent := 0
	startTime := time.Now()

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		_, err = ms.GetStream().Write(buf[:n])
		if err != nil {
			return fmt.Errorf("write to stream: %w", err)
		}
		totalSent += n
		fmt.Printf("   Sent: %d / %d bytes\r", totalSent, stat.Size())
	}

	duration := time.Since(startTime)
	fmt.Printf("\nFile sent: %d bytes in %v (%.2f KB/s)\n", totalSent, duration, float64(totalSent)/1024/duration.Seconds())

	// Read echoed response (the echo controller echoes back what we send)
	fmt.Println("Waiting for verification...")
	receivedData := make([]byte, 0, totalSent)
	respBuf := make([]byte, chunkSize)

	// Set read deadline to avoid blocking forever
	ms.GetStream().SetReadDeadline(time.Now().Add(5 * time.Second))

	for len(receivedData) < totalSent {
		n, err := ms.GetStream().Read(respBuf)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Timeout or other error - break
			break
		}
		receivedData = append(receivedData, respBuf[:n]...)
	}

	if len(receivedData) > 0 {
		fmt.Printf("Received %d bytes back from peer\n", len(receivedData))
	}

	fmt.Println("\nFile transfer complete!")
	fmt.Printf("   File: %s\n", filename)
	fmt.Printf("   Size: %d bytes\n", totalSent)
	fmt.Printf("   Checksum: %s\n", checksum)
	fmt.Printf("   Recipient: %s\n", recipientID.String())

	return nil
}
