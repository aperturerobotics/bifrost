//go:build !js

package bldr_buildbudget

import "github.com/s4wave/spacewave/bldr/util/hostmem"

// availableHostMemoryBytes returns the host memory the budget may claim.
func availableHostMemoryBytes() (uint64, error) {
	return hostmem.Available()
}
