package block

import "context"

// readAheadKey identifies the operation's preferred range-fetch minimum.
type readAheadKey struct{}

// WithReadAhead requests at least bytes per remote range-cache miss for this
// operation. Stores may limit the request to an uncovered gap, the end of a
// file, or their transport and memory bounds. Resident reads need no fetch.
// The hint does not change a shared store's default policy.
func WithReadAhead(ctx context.Context, bytes int) context.Context {
	return context.WithValue(ctx, readAheadKey{}, max(0, bytes))
}

// ReadAhead returns the requested range-fetch minimum, or zero when unset.
func ReadAhead(ctx context.Context) int {
	bytes, _ := ctx.Value(readAheadKey{}).(int)
	return bytes
}
