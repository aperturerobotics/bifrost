# Bifrost

> Peer-to-peer communications framework with pluggable network transports.

## Introduction

Bifrost is a multi-hop mesh networking router. It supports:

 - *Transport multiplexing*: efficient communication over any network transport.
 - *Persistent identity*: identify processes as they move between machines.
 - *Encryption*: encryption and PKI-based authentication. End to end encryption.
 - *Cross platform*: support every platform, including the web browser.
 - *Modular components*: pluggable implementations for flexible configuration.

Internally, Bifrost uses the controllerbus state-machine model to execute and
manage transports, links, routing, and other processes. Declarative state
management controls listening and establishing desired connectivity.

## Examples

From the API and CLI, most Bifrost functionality is exposed:

 - Mount a peer: load a private key into the daemon.
 - Forward incoming streams with a protocol ID to a multiaddress
 - Proxy incoming connections to a listener to a remote peer
 - Open a stream with a remote peer and a given protocol ID
 - Accept a stream for a local peer with a given protocol ID

With the above operations, all stream handling and interaction with the network
is exposed to the API and command line. Some examples:

```
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
which are associated with a local keypair.

Transports are responsible for handshaking their identity and providing stream
multiplexing, encryption, and ordering. The Bifrost codebase contains common
implementations for packet-based and stream-based transports, based primarily on
the [quic-go] implementation of the Quic UDP protocol.

[quic-go]: https://github.com/lucas-clemente/quic-go

Incoming streams have a header with desired protocol ID. The Transport
controller creates a HandleStream singleton directive with the protocol and peer
info. The appropriate controller for the protocol responds to the directive and
handles the stream.

### PubSub

A PubSub is a controller that supports topic-based at-least-once message
delivery to a network of interested peers. Optimizations of interest are
minimum-spanning-tree transmission path building algorithms, where wasted
bandwidth is minimized.
