//go:build (!darwin && !linux && !windows) || js

package hostmem

import "errors"

// errUnsupported reports a platform without a memory probe.
var errUnsupported = errors.New("host memory detection is unsupported on this platform")

// Total reports that this platform has no memory probe.
func Total() (uint64, error) {
	return 0, errUnsupported
}

// Available reports that this platform has no memory probe.
func Available() (uint64, error) {
	return 0, errUnsupported
}
