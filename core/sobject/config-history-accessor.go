package sobject

import "context"

// SharedObjectConfigHistoryAccessor reads the retained native lineage of an accepted head.
// Missing history cannot be reconstructed from client observations.
type SharedObjectConfigHistoryAccessor interface {
	// ReadSharedObjectConfigHistory returns the held checkpoint and verified changes through target.
	ReadSharedObjectConfigHistory(ctx context.Context, target *SharedObjectConfig) (*SharedObjectConfig, []*SOConfigChange, error)
}
