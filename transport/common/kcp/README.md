# KCP Packet-Conn Implementation

This is an implementation of the packet-conn common package that uses Kcp
instead of Quic.

The primary advantage here is that Kcp supports very low MTU values, as well as
custom block cyphers and compression.

Known issues:

 - Unencrypted streams currently do not work (raw streams)
 - The packet switching is buggy and prone to break
 - KCP in general is a bit buggy
 
This package is here as a stop-gap for xbee until we find a way to transmit the
minimum 1000 bytes packet size for Quic over xbee.
