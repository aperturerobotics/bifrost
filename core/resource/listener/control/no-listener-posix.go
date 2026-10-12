//go:build !js && !windows

package control

import "syscall"

// errRefused is the errno of a Unix socket dial that no listener accepts.
var errRefused error = syscall.ECONNREFUSED
