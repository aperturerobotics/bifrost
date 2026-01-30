# Multi-Transport Bridge

Demonstrates Bifrost's transport abstraction - write protocol logic once, run over any transport.

## What This Shows

The same application protocol works identically over:
- **UDP/QUIC** - High performance, NAT traversal
- **WebSocket** - Firewall friendly, works everywhere
- **WebRTC** - Browser-to-browser, P2P

No code changes when switching transports. Just change the configuration.

## Quick Start

### UDP Transport
```bash
# Terminal 1
go run main.go --transport udp --listen :5000

# Terminal 2
go run main.go --transport udp --listen :5001 --dial <PEER_ID>@127.0.0.1:5000
```

Note: All flags can also be set via environment variables (e.g., `TRANSPORT=udp LISTEN=:5000`).

### WebSocket Transport (same code)
```bash
# Terminal 1
go run main.go --transport websocket --listen :5000

# Terminal 2
go run main.go --transport websocket --listen :5001 --dial <PEER_ID>@ws://127.0.0.1:5000
```

The protocol code is identical. Only the transport config changed.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Application Layer                     │
│              (Your Protocol - Unchanged)                 │
│                    Echo Protocol                        │
└─────────────────────────────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
┌─────────────────┐ ┌─────────────┐ ┌─────────────┐
│  UDP Transport  │ │ WebSocket   │ │  WebRTC     │
│  (QUIC/mTLS)    │ │  Transport  │ │  Transport  │
└─────────────────┘ └─────────────┘ └─────────────┘
           │               │               │
           ▼               ▼               ▼
     ┌──────────┐    ┌──────────┐    ┌──────────┐
     │  UDP/IP  │    │   WS/TCP │    │  ICE/UDP │
     └──────────┘    └──────────┘    └──────────┘
```

## Why This Matters

Traditional networking requires you to:
1. Write different code for UDP vs WebSocket vs WebRTC
2. Handle protocol differences manually
3. Maintain multiple implementations

With Bifrost:
1. Write protocol logic once
2. Choose transport via configuration
3. Same code runs everywhere

## Testing

Run the end-to-end tests:

```bash
go test -tags test_examples .
```

Tests verify:
- Same protocol works across multiple simulated transports
- Message delivery is identical regardless of transport
- Bridge functionality (A → B → C) works transparently

## Use Cases

- **Hybrid Apps**: Use UDP for desktop, WebSocket for web, same protocol
- **Resilient Services**: Auto-fallback from UDP to WebSocket if blocked
- **Edge Computing**: Deploy same code to cloud, IoT, browser
- **Protocol Testing**: Test protocol once, deploy to any transport

## Code Comparison

### Without Bifrost
```go
// UDP version - completely different code
func sendUDP(addr string, msg []byte) { ... }

// WebSocket version - completely different code
func sendWS(url string, msg []byte) { ... }

// WebRTC version - completely different code
func sendWebRTC(peerID string, msg []byte) { ... }
```

### With Bifrost
```go
// One protocol - works over any transport
func sendMessage(peerID peer.ID, msg []byte) {
    // Same code works over UDP, WebSocket, WebRTC
    stream.Write(msg)
}

// Transport selected via config:
// config := &udptpt.Config{...}     // UDP
// config := &wstpt.Config{...}      // WebSocket
// config := &webrtc.Config{...}     // WebRTC
```
