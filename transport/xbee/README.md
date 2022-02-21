# xbee transport

> A xbee API mode transport for bifrost.

This is an implementation of a bifrost transport that uses xbee radios.

 - Xbees are controlled with the [gobee](https://github.com/pauleyj/gobee) library in API mode.
 - Incoming packets are matched to a connection by transmitter ID
 - Incoming packet without a matching connection triggers an Accept() return
 - If configured, nodes may emit discovery pings, which contain a signed message
   with peer ID, timestamp, public key, device MAC.
 - Not hearing from a peer for a connection timeout length leads to a connection close event.

Note: discovery is not yet implemented in Bifrost.
 
## Xbee Setup

Reset the xbees to default values, then set:

 - AP: 2 (API mode with escapes)
 - Baud: 115200
 
This will be automatic eventually.


## E2E Test

Testing basic communications between two xbees connected via USB or another
serial port is easily done with [xbe2e](./toys/xbe2e). The tool accepts CLI
arguments to set the baudrates and serial ports. It opens + configures both xbee
using the xbserial common code, reads the local addresses, and transmits some
data between them (unencrypted) as a basic test.
