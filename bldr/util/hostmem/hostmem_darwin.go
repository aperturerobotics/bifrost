//go:build darwin && !js

package hostmem

import "golang.org/x/sys/unix"

// Total returns the physical memory of the host in bytes.
func Total() (uint64, error) {
	return unix.SysctlUint64("hw.memsize")
}

// Available returns the physical memory of the host in bytes. Darwin exposes
// no stable available-memory syscall, so its capacity is the upper bound.
func Available() (uint64, error) {
	return Total()
}
