//go:build !js && !darwin && !linux

package spacewave_cli

import "net"

// peerProcess returns zero and an empty name: this platform does not report the
// process on the other end of a Unix socket.
func peerProcess(net.Conn) (int, string) {
	return 0, ""
}
