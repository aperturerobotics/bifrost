package sobject

import "context"

// SharedObjectReadCheckpointAccessor exposes history retained before read access ended.
// The snapshot is immutable and conveys no current membership or write authority.
type SharedObjectReadCheckpointAccessor interface {
	// GetSharedObjectReadCheckpoint returns nil when no history was retained.
	GetSharedObjectReadCheckpoint(ctx context.Context) (*SharedObjectReadCheckpoint, error)
}

// SharedObjectReadCheckpoint binds a historical World to its historical audience.
// Config is descriptive history, never current participant authority.
type SharedObjectReadCheckpoint struct {
	// Snapshot decodes only the root retained at departure.
	Snapshot SharedObjectStateSnapshot
	// Config records the audience at that root.
	Config *SharedObjectConfig
}
