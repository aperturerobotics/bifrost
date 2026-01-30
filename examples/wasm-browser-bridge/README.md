# WASM Browser Bridge

The Bifrost networking stack compiled to WebAssembly and running in a web browser. The browser connects directly to native Go backends over WebRTC or WebSocket.

## Features

- **Browser-Native Bridge** - WASM connects to Go backends
- **Encrypted** - All browser-native traffic uses TLS 1.3
- **Bidirectional** - Stream data both ways
- **Universal** - Works over WebRTC (P2P) or WebSocket

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  HTML/JS                                             │   │
│  │    │                                                 │   │
│  │    ▼                                                 │   │
│  │  WebAssembly (Go compiled to WASM)                   │   │
│  │    │                                                 │   │
│  │    ▼                                                 │   │
│  │  ┌──────────────┐    ┌──────────────┐                │   │
│  │  │ Bifrost Stack│◄──►│ WebRTC/WS    │                │   │
│  │  │ - Peer ID    │    │ Transport    │                │   │
│  │  │ - QUIC/mTLS  │    │              │                │   │
│  │  │ - Streams    │    │              │                │   │
│  │  └──────────────┘    └──────────────┘                │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ WebRTC DataChannel
                              │ or WebSocket
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Native Backend (Go)                     │
│                    ┌─────────────────┐                      │
│                    │  Bifrost Stack  │                      │
│                    │  - Same code    │                      │
│                    └─────────────────┘                      │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Start Backend Server
```bash
cd wasm-browser-bridge
go run main.go --listen :5000 --http :8080
```

Note: All flags can also be set via environment variables (e.g., `LISTEN=:5000 HTTP=:8080`).

Note the Peer ID that is printed.

### 2. Build Browser Client
```bash
# Build WASM module
GOOS=js GOARCH=wasm go build -o client.wasm main_wasm.go

# Serve browser client
python3 -m http.server 8000
```

### 3. Connect from Browser
1. Open http://localhost:8000
2. Enter the backend Peer ID
3. Click "Connect"
4. Make HTTP requests through the encrypted WebRTC tunnel

## How It Works

1. **Compile**: Bifrost + your app compiled to WebAssembly
2. **Connect**: Browser establishes WebRTC or WebSocket to native peer
3. **Encrypt**: QUIC/TLS 1.3 encrypts all traffic
4. **Stream**: Full bidirectional streaming between browser and backend
5. **Request**: Browser makes HTTP requests that tunnel through WebRTC

## Testing

Run the end-to-end tests:

```bash
go test -tags test_examples .
```

Tests verify:
- Browser-native connectivity (simulated)
- Multiple backend connections from one browser
- Message delivery through the bridge

## Comparison: Traditional vs Bifrost Browser

### Traditional Web App
```
Browser → HTTPS → Load Balancer → App Server → Database

Multiple hops, complex infrastructure, single point of failure
```

### Bifrost Browser App
```
Browser ←──WebRTC──→ Backend (Go)
        (encrypted P2P)

Direct connection, no infrastructure, P2P
```
