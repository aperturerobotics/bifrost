//go:build tinygo

package provider_local

import "github.com/s4wave/spacewave/core/transport"

// newLocalSessionNetwork leaves browser transport to its host-owned network runtime.
func newLocalSessionNetwork() transport.SessionTransportOption { return nil }
