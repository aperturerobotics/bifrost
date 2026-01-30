# Pub/Sub Events

Event broadcasting over a peer-to-peer mesh network.

## What This Shows

A decentralized pub/sub system that works without Kafka, RabbitMQ, or any brokers. Events propagate through the mesh to reach all subscribers.

## Features

- **No Brokers** - Direct peer-to-peer event distribution
- **Auto-Propagation** - Events route through the mesh automatically
- **Scalable** - Add peers without central bottlenecks
- **Secure** - All events are signed and verifiable

## Quick Start

### Start Publisher Node
```bash
cd pubsub-events
go run main.go --listen :5000 --topic "news"
```

Note: All flags can also be set via environment variables (e.g., `LISTEN=:5000 TOPIC=news`).

Note the Peer ID.

### Start Subscriber Node
```bash
go run main.go --listen :5001 --topic "news" --dial <PUBLISHER_PEER_ID>@127.0.0.1:5000
```

### Broadcast Events
Type messages in either terminal:
```
> Breaking: New release announced!
> System update scheduled
```

All peers subscribed to "news" receive the events.

## Architecture

```
        ┌─────────┐
        │ Publisher│
        │  :5000   │
        └────┬────┘
             │ Event: "news"
    ┌────────┼────────┐
    ▼        ▼        ▼
┌──────┐ ┌──────┐ ┌──────┐
│Sub 1 │ │Sub 2 │ │Sub 3 │
└──────┘ └──────┘ └──────┘
```

## How It Works

1. **Subscription**: Peers subscribe to topics via FloodSub
2. **Propagation**: Subscriptions flood through the network
3. **Publishing**: Events broadcast to all interested peers
4. **Relaying**: Peers relay events they receive to connected peers

## Testing

Run the end-to-end tests:

```bash
go test -tags test_examples .
```

Tests verify:
- Event broadcast to multiple subscribers
- Multi-hop event relaying (A → B → C → D)
- Message delivery across mesh topologies

## Use Cases

- **IoT Sensor Network** - Sensors publish readings, consumers subscribe
- **Game State Sync** - Game events propagate to all players
- **Chat Rooms** - Messages broadcast to room participants
- **Market Data** - Stock prices distributed to trading bots
- **Config Updates** - Push configuration changes to services

## Comparison: Traditional vs Bifrost Pub/Sub

### Traditional (Kafka)
```
Producers → Kafka Brokers → Consumers
              ↑
         Single point of failure
         Operational complexity
```

### Bifrost (Decentralized)
```
Peer A ←────→ Peer B ←────→ Peer C
   ↓              ↓              ↓
Subscriber   Subscriber    Subscriber

No brokers, no single point of failure
```
