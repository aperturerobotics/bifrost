package world

import "context"

// View is a generated message whose content SendViews can compare.
type View[T any] interface {
	// EqualVT reports whether two messages carry the same content.
	EqualVT(T) bool
}

// SendViews sends the first projection and each later unequal projection.
// source and read follow WatchView's contract, including context-carried
// transactions and RerunAt. read must return an immutable message; send must not
// mutate it. Returns when ctx ends or read or send fails.
func SendViews[T View[T]](
	ctx context.Context,
	source WorldWaitSeqno,
	read func(context.Context, WorldState) (T, error),
	send func(T) error,
) error {
	// Retain the last successfully sent projection for this subscription.
	var previous T
	sent := false

	// Evaluate on relevant changes, suppressing equal projections.
	return WatchView(ctx, source, func(ctx context.Context, ws WorldState) error {
		// Compute the next immutable projection.
		next, err := read(ctx, ws)
		if err != nil {
			return err
		}
		if sent && next.EqualVT(previous) {
			return nil
		}

		// Advance the comparison baseline only after sending succeeds.
		if err := send(next); err != nil {
			return err
		}
		previous, sent = next, true
		return nil
	})
}
