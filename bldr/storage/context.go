package storage

import "context"

type hostStorageKey struct{}

// WithHostStorageID carries the runtime's explicit plugin storage selection.
// The ID is resolved on the runtime bus that will host child plugins.
func WithHostStorageID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, hostStorageKey{}, id)
}

// GetHostStorageID returns the explicit child-plugin storage selection, if any.
func GetHostStorageID(ctx context.Context) string {
	id, _ := ctx.Value(hostStorageKey{}).(string)
	return id
}
