package provider_transfer

// TransferOptions carries the optional collaborators of a Transfer. The zero
// value is valid: no cleanup, no checkpointing, and no space-ID filter.
type TransferOptions struct {
	// Cleanup deletes the source volume after a successful transfer.
	Cleanup CleanupSource
	// Checkpoint persists resumable transfer progress.
	Checkpoint CheckpointStore
	// FilterSpaceIDs restricts the transfer to the listed Space IDs.
	FilterSpaceIDs []string
}
