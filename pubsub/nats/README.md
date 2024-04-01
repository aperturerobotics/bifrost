# NATS as a Pub-Sub Protocol Module

Nats is a CNCF-certified pub-sub, streaming, and communications system built in
Go for the cloud, IoT, and edge.

Nats is forked into the aperture-2.0 branch, which removes all listeners and
other complex OS-specific code, significantly simplifying the codebase. It then
implements these functions with directives to the Aperture stack.

Nats uses nkeys (public-private key crypto) with flexible wrappers internally.
This made it simple to import Aperture's flexible public/private key interface
system with some simple glue code. Nats Accounts map directly to Bifrost Peers.
The Nats 2.0 systems for authentication, authorization, and permissions are all
effectively integrated with the Aperture stack.

The result is a highly automated Nats deployment which can work across arbitrary
network transports, firewalls, air-gaps, compute architectures, even in the web
browser. Nats is now communicating over QUIC-over-UDP instead of TCP, but could
communicate over other exotic transports, even sound (like chirp.js).

This fully supports all protocols for Nats. Previously, NATS would require
manual configuration for TLS, accounts, and HTTP listeners. With Bifrost, these
protocols are exposed as stream endpoints bound to the peer's public key.

Bifrost's one-size-fits-all PubSub controller effectively manages and automates
connections between next-hop routing peers. As the system supports the exact
same drop-in protocol compatibility with upstream nats and minimal
modifications, the system also inter-operates with any existing or third-party
nats clusters or clients as expected.


# NATS as a communications backend for Bifrost

An existing NATS deployment or NATS-client protocol compatible cluster can also
be used as backing infrastructure for Bifrost. The "NATS Client Controller"
operates as a client to existing NATS infrastructure, importing the nats-go
client and speaking the protocol over HTTP/s.

Bifrost imports the following capabilities from the NATS 2.0 client:

 - Account: the controller has a private key which is used to authenticate.
   - Bifrost Peer maps to Nkeys PKI and uses the PKI challenge auth method.
 - Stream: peers communicate over a Link via Transport backed by a NATS Stream
   - Provides full two-way communication connectivity between peers.
   - Requires permission to import/export streams + services.
   - Peer A publishes a "Session Establish" service to NATS.
   - Peer B discovers the Service and sends a request to dial the Transport.
   - Peer A acks with a unique identifier for the session.
   - The unique identifiers from both peers are mixed to form a shared secret S.
   - Secret S is used to generate the session ID and stream ID prefix.
   - Peer A publishes a Stream {session-id} to peer B and vise-versa.
   - Packet-Conn protocol (currently Quic) is used to TLS handshake.
   - Session continues as usual.
   - If session is closed for any reason the streams are unpublished.
   - Embedded NATS hosts can communicate over this encrypted Link.
 - PubSub: use existing NATS.io infrastructure for pub-sub.
   - Requires permission to publish/subscribe to arbitrary topics below a
     user-specified prefix.
   - The Bifrost pub-sub topic is transformed w/ the prefix and passed to Nats.
   - Trusts the underlying NATS infrastructure to not forge messages.

This allows existing deployment tools for NATS to be used as a backing
infrastructure for Bifrost and all other systems to communicate with.
