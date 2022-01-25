![Bifrost](./doc/img/bifrost-logo.png)

## Introduction

**Bifrost** is a peer-to-peer communications engine with pluggable transports:

 - **Cross-platform**: supports web browsers, servers, desktop, mobile, ...
 - **Efficient**: multiplex many simultaneous streams over a single Link.
 - **Encryption**: identify, authenticate, and encrypt each Link between peers.
 - **Flexible**: use multiple transports, protocols, simultaneously.
 - **Meshing**: supports multi-hop routing to a desired target peer w/ circuits.
 - **PubSub**: publish/subscribe channels with pluggable implementations.
 - **Robust**: uses Quic for reliable connections over lossy transports.

Bifrost provides [ControllerBus] controllers and directives to manage links
between peers, transports, routing, and other higher-level processes. It has
extensive and flexible configuration. Connections are created on-demand.

[ControllerBus]: https://github.com/aperturerobotics/controllerbus

## Overview

[![Go Reference](https://pkg.go.dev/badge/github.com/aperturerobotics/bifrost.svg)](https://pkg.go.dev/github.com/aperturerobotics/bifrost)

Bifrost is designed around the following core concepts:

 - **Peer**: a routable process or device with a keypair.
 - **Transport**: protocol or hardware for communication Link between two Peer.
 - **Link**: a connection between two peers over a Transport.
 - **Stream**: channel of data between two Peer with a protocol type.
 - **PubSub**: at-least-once delivery of messages to named topics.
 - **Route**: a multi-hop path through the network between two Peer.
 - **Circuit**: a type of Link which implements a multi-hop connection.

Integrates with networking, pubsub, and RPC libraries like [libp2p], [noise],
[drpc], and [nats].

[drpc]: https://github.com/storj/drpc
[libp2p]: https://libp2p.io/
[noise]: https://github.com/perlin-network/noise
[nats]: https://nats.io

In practice, this makes it easy to mix-and-match communications hardware and
software in a configurable and reproducible way, often hot-loading support for
new functionality without requiring a restart of the entire process.

The [network simulator], [testbed], and [in-proc transport] can be used to write
end-to-end tests as Go unit tests. The mock transports use identical code to the
real transports, and appear the same to the application code.

[network simulator]: ./sim
[testbed]: ./testbed
[in-proc transport]: ./transport/inproc

[EntityGraph] exposes the internal state representation of Bifrost to
visualizers and instrumentation via a graph-based inter-connected entity model.

[EntityGraph]: https://github.com/aperturerobotics/entitygraph

### Transports and Links

A Link is a packet stream between two Peer. Links are created by Transports,
which are associated with a local private keypair.

Transports are responsible for handshaking their identity and providing stream
multiplexing, encryption, and ordering. The Bifrost codebase contains common
implementations for packet-based and stream-based transports, based primarily on
the [quic-go] implementation of the Quic UDP protocol.

[quic-go]: https://github.com/lucas-clemente/quic-go

The Transport controller creates a HandleStream directive with the protocol and
peer info. The appropriate controller for the protocol responds to the directive
and handles the incoming stream. This decouples the communication from the app.

### PubSub

A PubSub is a controller that supports topic-based at-least-once message
delivery to a network of interested peers.

Nats is also supported as PubSub protocol. There is also a simpler "floodsub"
implementation, and support for libp2p pubsub algorithms.

## Examples

Bifrost can be used as either a Go library or a command-line / daemon.

```bash
GO111MODULE=on go install -v github.com/aperturerobotics/bifrost/cmd/bifrost
```

Access help by adding the "-h" tag or running "bifrost help."

As a basic example, launch the daemon:

```
bifrost daemon --write-config --hold-open-links --pubsub floodsub --api-listen :5110 --udp-listen :5112
```

### YAML Configuration

An example of a ConfigSet in YAML format for the daemon: `bifrost_daemon.yaml`:

```yaml
# Start a UDP listener on port 5112.
my-udp:
  config:
    listenAddr: :5112
  id: bifrost/udp/1
  revision: 1

# Use the floodsub driver for PubSub.
pubsub:
  config: {}
  id: bifrost/floodsub/1
  revision: 1
```

### Client CLI

Most Bifrost functionality is exposed on the client CLI and GRPC API:

 - Mount a peer by loading a private key into the daemon.
 - Forward incoming streams with a protocol ID to a multiaddress
 - Proxy incoming connections to a listener to a remote peer
 - Open a stream with a remote peer and a given protocol ID
 - Accept a stream for a local peer with a given protocol ID

The client CLI has the following help output:

```
bifrost client command [command options] [arguments...]

COMMANDS:
   local-peers           returns local peer info
   identify              Private key will be loaded with a peer controller
   subscribe             Subscribe to a pubsub channel with a private key or mounted peer and publish base64 w/ newline delim from stdin.
   forward               Protocol ID will be forwarded to the target multiaddress
   accept                Single incoming stream with Protocol ID will be accepted
   dial                  Single outgoing stream with Protocol ID will be dialed
   listen                Listen on the multiaddress and forward the connection to a remote stream.
   controller-bus, cbus  ControllerBus system sub-commands.
```

With the above operations, all stream handling and interaction with the network
is exposed to the API and command line. Some examples:

```
  # Note: you can edit bifrost_daemon.yaml to change settings.
  # Once the daemon configuration exists, you can now just run:
  bifrost daemon

  # While the command is executing, the private key will be attached.
  bifrost client identify --peer-priv priv-key.pem

  # While the command is executing, a forwarding controller will be running.
  # Protocol ID will be forwarded to the target multiaddress
  # Handles HandleMountedStream directives by contacting the target.
  # HTTP can be easily proxied through an encrypted stream this way.
  bifrost client forward \
    --peer-id <agent-id> \
    --protocol-id /x/myproto \
    --target /ip4/127.0.0.1/tcp/8000

  # While the command is executing, a proxying controller will be running.
  # Protocol ID will be proxied from the listen multiaddress to the target peer.
  # Calls OpenStream to build a stream from <source-peer-id> to <target-peer-id>.
  # HTTP can be easily proxied through an encrypted stream this way.
  bifrost client listen \
    --peer-id <target-peer-id> \
    --from-peer-id <source-peer-id> \
    --protocol-id /x/myproto \
    --listen /ip4/127.0.0.1/tcp/8001

  # Wait for a stream to be opened to the mounted peer with the protocol ID /x/myproto
  # Standard output is the incoming data stream, standard input is the outgoing data stream.
  # Standard error is used for logging.
  bifrost client accept \
    --local-peer-id <peer-id> \
    --protocol-id /x/myproto 

  # Establish a stream.
  # Standard output is the incoming data stream, standard input is the outgoing data stream.
  # Standard error is used for logging.
  bifrost client dial \
    --peer-id <target-peer-id> \
    --local-peer-id <local-peer-id> \
    --protocol-id /x/myproto
```

## Support

Bifrost is built & supported by Aperture Robotics, LLC.

Please open a [GitHub issue] with any questions / issues.

[GitHub issue]: https://github.com/aperturerobotics/bifrost/issues/new

... or feel free to reach out on [Matrix Chat].

[Matrix Chat]: https://matrix.to/#/#aperture-robotics:matrix.org
