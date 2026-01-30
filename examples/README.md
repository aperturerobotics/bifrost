# Bifrost Examples

This directory contains demos of Bifrost's peer-to-peer capabilities.

## Quick Start

Run any example's test with:

```bash
go test -tags test_examples ./...
```

## Examples

### 1. `mesh-chat/` - Terminal Chat
A terminal-based chat that works across any transport (UDP, WebSocket, WebRTC).

### 2. `p2p-filedrop/` - Peer-to-Peer File Transfer
Transfer files directly between peers with end-to-end encryption.

### 3. `multi-transport-bridge/` - Transport Abstraction Demo
Shows how Bifrost bridges between UDP, WebSocket, and WebRTC transports.

### 4. `pubsub-events/` - Event Broadcasting
Broadcast events across a mesh network with automatic propagation.

### 5. `wasm-browser-bridge/` - Browser-to-Native Bridge
WebAssembly example showing browser-to-server communication over WebRTC.

