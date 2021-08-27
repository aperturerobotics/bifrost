# Bifrost

> Peer-to-peer communications framework with pluggable network transports.

## Introduction

Bifrost is a set of components for robust communication between peers. It has:

 - **Cross-platform**: supports web browsers, servers, desktop, mobile, ...
 - **Encryption**: identify, authenticate, and encrypt each Link between peers.
 - **Pluggable Transports**: use multiple modes of communication simultaneously.
 - **Mesh Routing**: multi-hop routing to a desired target peer.
 - **PubSub**: publish/subscribe channels with pluggable implementations.

Bifrost provides [controllerbus] controllers and directives to manage links
between peers, transports, routing, and other higher-level processes. It has
extensive and flexible configuration. Connections are created on-demand.

Wrappers are provided for interop with and support for other networking and
pubsub libraries and systems like [libp2p] and [noise], [nats], and more. These
projects have been simplified by replacing all connections and crypto management
with bindings to the Bifrost abstractions.

In practice, this makes it trivial to mix-and-match communications hardware and
software, often hot-loading support for new functionality without requiring a
restart of the entire process. (see: ControllerBus plugins & hot reload).

[libp2p]: https://libp2p.io/
[noise]: https://github.com/perlin-network/noise
[nats]: https://nats.io

## Overview

Bifrost is designed around the following core concepts:

 - Peer: a routable process or device with an identity.
 - Transport: protocol or hardware for communication Link between two Peer.
 - Link: a connection between two peers over a Transport.
 - Stream: channel of packets between two Peer with a configured protocol type.
 - PubSub: at-least-once delivery of messages to named topics.
 - Route: a multi-hop path through the network between two Peer.

Bifrost is built on the ControllerBus framework, which defines the Config,
Controller, Directive structures and behaviors. All components are implemented
as controllers, and have associated factories.

An EntityGraph controller is provided. EntityGraph exposes the internal state
representation of Bifrost to visualizers and instrumentation via a graph-based
inter-connected entity model.

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
delivery to a network of interested peers. Optimizations of interest are
minimum-spanning-tree transmission path building algorithms, where wasted
bandwidth is minimized. 

Nats is supported as a communications protocol and controllers, and is
protocol-compatible with existing Nats 2.0 clients. There is also a simpler
"floodsub" implementation, and support for libp2p pubsub algorithms.

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
