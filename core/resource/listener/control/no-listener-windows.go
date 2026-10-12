//go:build windows

package control

import "golang.org/x/sys/windows"

// errRefused is the errno of a Unix socket dial that no listener accepts.
// Windows reports it for an absent socket path as well as a stale one, and it
// differs from syscall.ECONNREFUSED, which Go invents on Windows.
var errRefused error = windows.WSAECONNREFUSED
