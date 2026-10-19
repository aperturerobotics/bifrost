//go:build darwin

package spacewave_cli

import (
	"bytes"
	"net"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// peerProcess returns the process ID and executable name of the process on the
// other end of a Unix socket, or zero and an empty name when the peer cannot be
// identified.
func peerProcess(conn net.Conn) (int, string) {
	// Read the peer process ID from the socket credentials.
	pid, ok := unixPeerPID(conn, func(fd int) (int, error) {
		return unix.GetsockoptInt(fd, unix.SOL_LOCAL, unix.LOCAL_PEERPID)
	})
	if !ok {
		return 0, ""
	}

	// kern.procargs2 holds argc and then the full executable path.
	if args, err := unix.SysctlRaw("kern.procargs2", pid); err == nil && len(args) > 4 {
		path, _, _ := bytes.Cut(args[4:], []byte{0})
		if len(path) != 0 {
			return pid, filepath.Base(string(path))
		}
	}

	// The kernel's command name is truncated but needs no extra access.
	proc, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil {
		return pid, ""
	}
	return pid, unix.ByteSliceToString(proc.Proc.P_comm[:])
}
