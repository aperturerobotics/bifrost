# WebRTC Chat Example

This example demonstrates a simple peer-to-peer chat application using WebRTC transport in Bifrost.

## Overview

The example consists of:
- A signaling server for WebRTC connection establishment
- Two command-line chat clients that connect via WebRTC
- Real-time messaging between peers

## Running the Example

### Generate Private Keys (if not already present)

```
mkdir -p priv
openssl genrsa -out priv/node-1.pem 2048
openssl genrsa -out priv/node-2.pem 2048
openssl genrsa -out priv/node-3.pem 2048
```

### Start the Signaling Server

```
cd server
go run -v ./
```

The signaling server listens on port 2253 for WebSocket connections. Note the peer ID that is displayed when the server starts, as you'll need to update the configuration files with this ID.

### Update Configuration Files

Update the `node-1.yaml` and `node-2.yaml` files with the signaling server's peer ID:

```yaml
signaling-websocket:
  id: bifrost/websocket
  config:
    dialers:
      <signaling-server-peer-id>:
        address: ws://127.0.0.1:2253/bifrost-ws

signaling:
  id: bifrost/signaling/rpc/client
  config:
    signalingId: webrtc-chat
    protocolId: webrtc-chat
    client:
      serverPeerIds:
      - <signaling-server-peer-id>
```

### Build and Run the Chat Clients

First, build the chat client:

```
cd chat
./build.bash
```

In one terminal, start the first chat client in listening mode:

```
./chat-client --username User1
```

Note the peer ID that is displayed when the client starts.

In another terminal, start the second chat client and connect to the first one:

```
./chat-client --peer-id <first-client-peer-id> --username User2
```

### Start Chatting

Once connected, you can type messages in either terminal and they will be sent to the other client via WebRTC.

## How it Works

1. The signaling server facilitates WebRTC connection establishment between peers
2. The two chat clients connect to each other via WebRTC using the signaling server
3. Once connected, messages are sent directly between peers via WebRTC data channels
4. The signaling server is only used for connection establishment, not for message relay

## Implementation Details

This example demonstrates the use of Bifrost's WebRTC transport for direct peer-to-peer communication. The chat functionality is implemented as a simple command-line interface that allows users to send and receive messages in real-time.
