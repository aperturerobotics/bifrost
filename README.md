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

Bifrost uses [ControllerBus] controllers and directives to manage links between
peers, transports, routing, and other higher-level processes with extensive and
flexible configuration.

[ControllerBus]: https://github.com/aperturerobotics/controllerbus

## Overview

[![Go Reference Widget]][Go Reference] [![Go Report Card Widget]][Go Report Card]

[Go Reference]: https://pkg.go.dev/github.com/aperturerobotics/bifrost
[Go Reference Widget]:https://pkg.go.dev/badge/github.com/aperturerobotics/bifrost.svg
[Go Report Card Widget]: https://goreportcard.com/badge/github.com/aperturerobotics/bifrost
[Go Report Card]: https://goreportcard.com/report/github.com/aperturerobotics/bifrost

Bifrost is designed around the following core concepts:

 - **Peer**: a routable process or device with a keypair.
 - **Transport**: a protocol which can create Links with other peers.
 - **Link**: a connection between two peers over a Transport.
 - **Stream**: channel of data between two Peer with a protocol type.
 - **RPC**: request/reply and bidirectional streaming remote calls.
 - **PubSub**: at-least-once delivery of messages to named topics.

Integrates with networking, pubsub, and RPC libraries like [libp2p], [noise],
[drpc], [starpc] and [nats].

[drpc]: https://github.com/storj/drpc
[libp2p]: https://libp2p.io/
[noise]: https://github.com/perlin-network/noise
[nats]: https://nats.io
[starpc]: https://github.com/aperturerobotics/starpc

The [network simulator], [testbed], and [in-proc transport] can be used to write
end-to-end tests as Go unit tests. The mock transports use identical code to the
real transports, and appear the same to the application code.

[network simulator]: ./sim
[testbed]: ./testbed
[in-proc transport]: ./transport/inproc

The [http] packages provide a http server and a mechanism for attaching http
handlers to the controller bus and using them to serve requests. There is also
an implementation of attaching and looking up http clients on the bus.

[http]: ./http

[EntityGraph] exposes the internal state representation of Bifrost to
visualizers and instrumentation via a graph-based inter-connected entity model.

[EntityGraph]: https://github.com/aperturerobotics/entitygraph

Configuring each component as an independent controller makes it easy to adapt
application code to different operating environments and protocols.

## Examples

[![Support Server](https://img.shields.io/discord/803825858599059487.svg?label=Discord&logo=Discord&colorB=7289da&style=for-the-badge)](https://discord.gg/KJutMESRsT)

Bifrost can be used as either a Go library or a command-line / daemon.

The [examples](./examples) directory contains yaml files to configure the
daemon, as well as "toys" which are self-contained Go program examples.

To install the CLI and daemon:

```bash
# Clone the repo and install.
git clone https://github.com/aperturerobotics/bifrost
cd ./bifrost/cmd/bifrost
go install -v

# Alternatively:
# Note: this currently fails on go >= 1.16 due to replace directives.
# See: https://github.com/golang/go/issues/44840
GO111MODULE=on go install -v github.com/aperturerobotics/bifrost/cmd/bifrost@master
```

Access help by adding the "-h" tag or running "bifrost help."

As a basic example, launch the daemon:

```
bifrost daemon \
  --write-config \
  --hold-open-links \
  --pubsub nats  \
  --api-listen :5110 \
  --udp-listen :5112
```

### Daemon CLI

The Bifrost daemon is configured with a YAML ConfigSet and/or via the API.

```
NAME:
   bifrost daemon - run a bifrost daemon

OPTIONS:
   --hold-open-links         if set, hold open links without an inactivity timeout [$BIFROST_HOLD_OPEN_LINKS]
   --websocket-listen value  if set, will listen on address for websocket connections, ex :5111 [$BIFROST_WS_LISTEN]
   --udp-listen value        if set, will listen on address for udp connections, ex :5112 [$BIFROST_UDP_LISTEN]
   --xbee-device-path value  xbee device path to open, if set [$BIFROST_XBEE_PATH]
   --xbee-device-baud value  xbee device baudrate to use, defaults to 115200 (default: 115200) [$BIFROST_XBEE_BAUD]
   --establish-peers value   if set, request establish links to list of peer ids [$BIFROST_ESTABLISH_PEERS]
   --xbee-peers value        list of peer-id@address known XBee peers [$BIFROST_XBEE_PEERS]
   --udp-peers value         list of peer-id@address known UDP peers [$BIFROST_UDP_PEERS]
   --websocket-peers value   list of peer-id@address known WebSocket peers [$BIFROST_WS_PEERS]
   --pubsub value            if set, will configure pubsub from options: [floodsub, nats] [$BIFROST_PUBSUB]
   --config value, -c value  path to configuration yaml file (default: "bifrost_daemon.yaml") [$BIFROST_CONFIG]
   --write-config            write the daemon config file on startup [$BIFROST_WRITE_CONFIG]
   --node-priv value         path to node private key, will be generated if doesn't exist (default: "bifrost_daemon.pem") [$BIFROST_NODE_PRIV]
   --api-listen value        if set, will listen on address for API connections, ex :5110 (default: ":5110") [$BIFROST_API_LISTEN]
   --prof-listen value       if set, debug profiler will be hosted on the port, ex :8080 [$BIFROST_PROF_LISTEN]
```

These CLI flags are provided for convenience to quickly configure a daemon, and
the resulting config can be written to a file with `--write-config` for further
adjustments to be made. Note, however, that additional controllers are available
which are not yet exposed via these flags.

### Client CLI

Most Bifrost functionality is exposed on the client CLI and RPC API:

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

```sh
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

### Example: forward HTTP traffic between peers

The following is a basic example of using the CLI to forward encrypted traffic
between a local port and a remote peer port, similar to SSH port forwarding:

```sh
  # note the peer id in the logs
  ./bifrost daemon \
            --write-config \
            --udp-listen :5000 \
            --node-priv daemon_node_priv_2.pem \
            --websocket-listen ":5111" \
            --prof-listen ":6201"

  # forward incoming connections to port 8000
  # example: "python3 -m http.server 8080"
  ./bifrost client forward \
            --protocol-id "test/protocol" \
            --target /ip4/127.0.0.1/tcp/8000

  # replace PEER-ID-HERE with the peer ID from the first daemon.
  # start a second daemon (in a new shell).
  ./bifrost daemon \
            --udp-listen :5001 \
            --udp-peers "PEER-ID-HERE@127.0.0.1:5000" \
            --api-listen ":5112"

  # tell it to listen on port 8082 and forward to the other peer.
  # try browsing to http://localhost:8082
  ./bifrost client --dial-addr 127.0.0.1:5112 listen \
            --peer-id "PEER-ID-HERE" \
            --protocol-id test/protocol \
            --listen /ip4/127.0.0.1/tcp/8002
```

This example shows how to run two daemons with information on how to contact
each other, and then "tell" the second daemon to listen on port 8002 and forward
any incoming connections to the remote peer with the given peer ID. 

When someone connects to port 8002 the EstablishLinkWithPeer directive is added
and the UDP transport opens the connection with the peer (on-demand.) The stream
is then negotiated. The remote daemon uses HandleMountedStream which is handled
by the "forwarding" controller, which forwards the stream to localhost at 8000.

### YAML Configuration

An example of a ConfigSet in YAML format for the daemon: `bifrost_daemon.yaml`:

```yaml
# Start a UDP listener on port 5112.
my-udp:
  id: bifrost/udp
  config:
    listenAddr: :5112

# Use the floodsub driver for PubSub.
pubsub:
  id: bifrost/floodsub
  config: {}
```

### Simulator and End-to-end Testing

```go
// g is the graph of peers and LANs.
g := graph.NewGraph()

// Add two peers to the graph.
p0 := addPeer(t, g)
p1 := addPeer(t, g)

// Connect them together with a LAN.
lan1 := graph.AddLAN(g)
lan1.AddPeer(g, p0)
lan1.AddPeer(g, p1)

// Run the simulator
sim := initSimulator(t, ctx, le, g)

// Test connecting p0 and p1 together!
simulate.TestConnectivity(ctx, p0, p1)
```

The simulator enables writing end-to-end tests of running controllers on peers
and validating that everything works with the full Bifrost stack in the loop. It
uses in-memory pipes to simulate local-area-network connections between peers.

Try the [example end-to-end test](./sim/tests/bifrost/basic_test.go) yourself:

```
go test -v -run Basic github.com/aperturerobotics/bifrost/sim/tests/bifrost
```

You will see the debug and quic logging followed by the success message:

```
INFO successful connectivity test: p0 <-> [lan1] <-> p1
```

## Concepts

### Transports and Links

A Link is a packet stream between two Peer. Links are created by Transports,
which are associated with a local private keypair.

Transports are responsible for handshaking their identity and providing stream
multiplexing, encryption, and ordering. The Bifrost codebase contains common
implementations for packet-based and stream-based transports, based primarily on
the [quic-go] implementation of the Quic UDP protocol.

[quic-go]: https://github.com/quic-go/quic-go

The HandleMountedStream directive contains incoming protocol and peer info. The
appropriate controller for the protocol responds to the directive and handles
the incoming stream. This decouples the transport layers from the protocols.

### PubSub

A PubSub is a controller that supports topic-based at-least-once delivery.

Nats and floodsub are currently supported as PubSub protocols.

## Developing on MacOS

On MacOS, some homebrew packages are required for `yarn gen`:

```
brew install bash make coreutils gnu-sed findutils protobuf
brew link --overwrite protobuf
```

Add to your .bashrc or .zshrc:

```
export PATH="/opt/homebrew/opt/coreutils/libexec/gnubin:$PATH"
export PATH="/opt/homebrew/opt/gnu-sed/libexec/gnubin:$PATH"
export PATH="/opt/homebrew/opt/findutils/libexec/gnubin:$PATH"
export PATH="/opt/homebrew/opt/make/libexec/gnubin:$PATH"
```

## Support

Please open a [GitHub issue] with any questions / issues.

[GitHub issue]: https://github.com/aperturerobotics/bifrost/issues/new

... or feel free to reach out on [Matrix Chat] or [Discord].

[Discord]: https://discord.gg/KJutMESRsT
[Matrix Chat]: https://matrix.to/#/#aperturerobotics:matrix.org

## License

Apache 2.0
