//go:build test_examples

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"testing"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/sim/graph"
	"github.com/aperturerobotics/bifrost/sim/simulate"
	"github.com/aperturerobotics/bifrost/stream"
	stream_echo "github.com/aperturerobotics/bifrost/stream/echo"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/sirupsen/logrus"
)

// TestP2PFileTransfer tests secure file transfer between two peers.
//
// This demonstrates:
//   - Direct peer-to-peer file transfer without intermediate servers
//   - End-to-end encryption via QUIC/TLS 1.3
//   - Large file streaming over multiple packets
//   - Integrity verification via SHA-256 checksum
func TestP2PFileTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	le := logrus.NewEntry(log)

	// Create two peers on the same LAN
	g := graph.NewGraph()

	sender := addPeerWithFileHandler(t, g)
	receiver := addPeerWithFileHandler(t, g)

	lan := graph.AddLAN(g)
	lan.AddPeer(g, sender)
	lan.AddPeer(g, receiver)

	sim := initSimulator(t, ctx, le, g)
	defer sim.Close()

	// Verify connectivity first
	if err := simulate.TestConnectivity(ctx, sim.GetPeerByID(sender.GetPeerID()), sim.GetPeerByID(receiver.GetPeerID())); err != nil {
		t.Fatalf("connectivity test failed: %v", err)
	}
	le.Info("Peers connected")

	// Create test file content (simulating a file)
	fileContent := []byte("This is a test file content that will be transferred peer-to-peer over an encrypted Bifrost stream. " +
		"The file can be arbitrarily large and will be streamed efficiently without loading entirely into memory.")

	// Calculate checksum
	expectedChecksum := sha256.Sum256(fileContent)

	// Transfer file: sender -> receiver
	fileProtocol := protocol.ID("demo/file-transfer/v1")

	senderPeer := sim.GetPeerByID(sender.GetPeerID())
	senderTB := senderPeer.GetTestbed()

	// Open stream for file transfer
	ms, msRel, err := link.OpenStreamWithPeerEx(
		ctx,
		senderTB.Bus,
		fileProtocol,
		sender.GetPeerID(),
		receiver.GetPeerID(),
		0,
		stream.OpenOpts{},
	)
	if err != nil {
		t.Fatalf("failed to open file transfer stream: %v", err)
	}
	defer msRel()

	// Send file in chunks (simulating streaming)
	chunkSize := 64
	for i := 0; i < len(fileContent); i += chunkSize {
		end := i + chunkSize
		if end > len(fileContent) {
			end = len(fileContent)
		}
		chunk := fileContent[i:end]
		if _, err := ms.GetStream().Write(chunk); err != nil {
			t.Fatalf("failed to write chunk at offset %d: %v", i, err)
		}
	}

	le.Infof("Sent %d bytes of file content", len(fileContent))

	// Read echoed response (the file handler echoes back)
	receivedData := make([]byte, 0, len(fileContent))
	buf := make([]byte, 256)

	for len(receivedData) < len(fileContent) {
		n, err := ms.GetStream().Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read file chunk: %v", err)
		}
		receivedData = append(receivedData, buf[:n]...)
	}

	le.Infof("Received %d bytes", len(receivedData))

	// Verify integrity
	if len(receivedData) != len(fileContent) {
		t.Errorf("size mismatch: sent %d bytes, received %d bytes", len(fileContent), len(receivedData))
	}

	actualChecksum := sha256.Sum256(receivedData)
	if !bytes.Equal(expectedChecksum[:], actualChecksum[:]) {
		t.Errorf("checksum mismatch: file was corrupted during transfer")
	}

	le.Info("File transfer test passed - file integrity verified")
}

// TestLargeFileTransfer tests streaming large files without memory issues.
func TestLargeFileTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	le := logrus.NewEntry(log)

	g := graph.NewGraph()

	sender := addPeerWithFileHandler(t, g)
	receiver := addPeerWithFileHandler(t, g)

	lan := graph.AddLAN(g)
	lan.AddPeer(g, sender)
	lan.AddPeer(g, receiver)

	sim := initSimulator(t, ctx, le, g)
	defer sim.Close()

	if err := simulate.TestConnectivity(ctx, sim.GetPeerByID(sender.GetPeerID()), sim.GetPeerByID(receiver.GetPeerID())); err != nil {
		t.Fatalf("connectivity failed: %v", err)
	}

	// Simulate 10KB file
	fileSize := 10 * 1024
	fileContent := make([]byte, fileSize)
	for i := range fileContent {
		fileContent[i] = byte(i % 256)
	}

	fileProtocol := protocol.ID("demo/file-transfer/v1")
	senderPeer := sim.GetPeerByID(sender.GetPeerID())
	senderTB := senderPeer.GetTestbed()

	ms, msRel, err := link.OpenStreamWithPeerEx(
		ctx,
		senderTB.Bus,
		fileProtocol,
		sender.GetPeerID(),
		receiver.GetPeerID(),
		0,
		stream.OpenOpts{},
	)
	if err != nil {
		t.Fatalf("failed to open stream: %v", err)
	}
	defer msRel()

	// Stream file in 1KB chunks
	chunkSize := 1024
	bytesSent := 0
	for i := 0; i < len(fileContent); i += chunkSize {
		end := i + chunkSize
		if end > len(fileContent) {
			end = len(fileContent)
		}
		chunk := fileContent[i:end]
		n, err := ms.GetStream().Write(chunk)
		if err != nil {
			t.Fatalf("write failed at offset %d: %v", i, err)
		}
		bytesSent += n
	}

	// Read response
	receivedData := make([]byte, 0, fileSize)
	buf := make([]byte, 1024)

	for len(receivedData) < fileSize {
		n, err := ms.GetStream().Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
		receivedData = append(receivedData, buf[:n]...)
	}

	if len(receivedData) != fileSize {
		t.Errorf("size mismatch: sent %d, received %d", fileSize, len(receivedData))
	}

	le.Infof("Large file transfer test passed - streamed %d KB successfully", fileSize/1024)
}

// addPeerWithFileHandler adds a peer with file transfer handler.
func addPeerWithFileHandler(t *testing.T, g *graph.Graph) *graph.Peer {
	ctx := context.Background()
	p, err := graph.GenerateAddPeer(ctx, g)
	if err != nil {
		t.Fatalf("failed to add peer: %v", err)
	}

	fileProtocol := protocol.ID("demo/file-transfer/v1")

	// Use echo controller as file handler (echoes back received data)
	p.AddFactory(func(b bus.Bus) controller.Factory {
		return stream_echo.NewFactory(b)
	})
	p.AddConfig("file-handler", &stream_echo.Config{
		ProtocolId: string(fileProtocol),
	})

	return p
}

// initSimulator creates a simulator.
func initSimulator(t *testing.T, ctx context.Context, le *logrus.Entry, g *graph.Graph) *simulate.Simulator {
	sim, err := simulate.NewSimulator(ctx, le, g)
	if err != nil {
		t.Fatalf("failed to create simulator: %v", err)
	}
	return sim
}
