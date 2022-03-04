# NATS Demo

These examples run multiple NATS instances within a single process to
demonstrate how it works with Bifrost. The peers speak Quic to each other and
send the NATS traffic over reliable streams.

The nats-single directory runs a single NATS instance as a test.

## Running

You can run these with "go run ./".
