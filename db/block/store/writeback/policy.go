package block_store_writeback

import "time"

// Policy sets when queued blocks upload.
type Policy struct {
	// Settle is how long a block stays queued before it uploads.
	Settle time.Duration
	// MaxPendingBytes uploads the oldest blocks before they settle while the
	// queued bytes exceed it. Zero sets no limit.
	MaxPendingBytes int64
}

// overLimit reports whether pending bytes exceed the policy's limit.
func (p Policy) overLimit(pendingBytes int64) bool {
	return p.MaxPendingBytes > 0 && pendingBytes > p.MaxPendingBytes
}
