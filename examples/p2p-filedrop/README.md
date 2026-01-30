# P2P File Drop

Peer-to-peer file transfer without servers or cloud storage.

## What This Shows

Transfer files directly between devices with end-to-end encryption. No intermediate servers, no cloud storage, no file size limits.

## Features

- **End-to-End Encryption** - TLS 1.3 via QUIC
- **Direct P2P** - No servers, no cloud, no intermediaries
- **Streaming** - Transfer files of any size without memory issues
- **Universal** - Works over UDP, WebSocket, WebRTC

## Quick Start

### 1. Start Receiver
```bash
cd p2p-filedrop
go run main.go --listen :5000
```

Note: All flags can also be set via environment variables (e.g., `LISTEN=:5000`).

Note the Peer ID that is printed.

### 2. Send File
```bash
go run main.go --send ./myfile.txt --to <RECEIVER_PEER_ID>@127.0.0.1:5000
```

The file is transferred directly peer-to-peer over an encrypted stream.

## Architecture

```
┌──────────┐                          ┌──────────┐
│  Sender  │ ◄── Encrypted Stream ──► │ Receiver │
│          │      (QUIC/TLS 1.3)      │          │
└──────────┘                          └──────────┘
     │                                       │
     ▼                                       ▼
┌──────────┐                          ┌──────────┐
│myfile.txt│                          │myfile.txt│
└──────────┘                          └──────────┘
```

## How It Works

1. **Identity**: Each peer has a cryptographically generated ID
2. **Discovery**: Peers find each other via multiaddrs (addresses)
3. **Connection**: QUIC/TLS 1.3 connection is established
4. **Streaming**: File is chunked and streamed over the encrypted link
5. **Verification**: SHA-256 checksum ensures integrity

## Testing

Run the end-to-end tests:

```bash
go test -tags test_examples .
```

Tests verify:
- Small file transfer (< 1KB)
- Large file transfer (10KB, streamed in chunks)
- Integrity verification via checksums

## Security

- All traffic encrypted with TLS 1.3
- Peer authentication via Ed25519 keys
- Perfect forward secrecy via QUIC
- No data passes through external servers
