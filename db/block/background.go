package block

import "context"

// backgroundKey identifies an operation marked as background work.
type backgroundKey struct{}

// WithBackground marks ctx as background work, such as a fill no person is
// waiting for. A store that fetches remote data admits a background fetch only
// while no foreground fetch on that store needs the link. Reads served from
// local or resident data do not wait.
func WithBackground(ctx context.Context) context.Context {
	return context.WithValue(ctx, backgroundKey{}, true)
}

// IsBackground reports whether ctx is marked as background work.
func IsBackground(ctx context.Context) bool {
	background, _ := ctx.Value(backgroundKey{}).(bool)
	return background
}
