package block_store_writeback

import "time"

// Status is the upload state of a Store.
type Status struct {
	// Enabled reports whether writes are queued for upload.
	Enabled bool
	// Pending is the number of blocks awaiting upload.
	Pending int
	// PendingBytes is the size of the blocks awaiting upload.
	PendingBytes int64
	// Oldest is the write time of the first block awaiting upload, while
	// Pending is not zero. Every block written before it has uploaded.
	Oldest time.Time
	// Err is the last upload failure, cleared by the next successful batch.
	Err error
	// Target is the remote every block stored before it was chosen is queued
	// for, empty until a backfill completes.
	Target string
	// Policy sets when queued blocks upload.
	Policy Policy
}
