package sobject

import (
	"context"

	"github.com/aperturerobotics/util/ccontainer"
	"github.com/s4wave/spacewave/db/block"
)

// StatePublisher is implemented by a SharedObject whose state shares the
// atomic publication domain of its block store. When the context passed to
// QueueOperation carries a PublishStateFunc, the provider calls it exactly
// once before the operation's state is admitted or committed.
type StatePublisher interface {
	// LocalStateHead returns a head update that sets key to value in the local
	// state store storeID, for a publication through the block store.
	LocalStateHead(storeID string, key, value []byte) *block.AtomicHeadUpdate
	// AccessAdmittedState adds a reference to the admitted state and returns
	// its container. The admitted state holds every operation whose
	// publication was admitted and has not failed, ahead of the durable state
	// of AccessSharedObjectState. Only the World engine of this device reads
	// it. Returns a release function; released is called if the Watchable
	// becomes invalid.
	AccessAdmittedState(ctx context.Context, released func()) (ccontainer.Watchable[SharedObjectStateSnapshot], func(), error)
}

// PublishStateFunc writes the World commit that queued an operation. Given
// the head of the operation's state write, it submits the head atomically
// with the commit and returns the publication's receipt; the state is then
// admitted, and becomes durable when the receipt succeeds. Given nil, or when
// the block store cannot publish, it writes the commit's blocks and returns a
// nil receipt; the provider then commits the state itself. The caller keeps
// ownership of the receipt.
type PublishStateFunc func(ctx context.Context, head *block.AtomicHeadUpdate) (*block.PublicationReceipt, error)

// publishStateKey is the context key of a PublishStateFunc.
type publishStateKey struct{}

// WithPublishState returns ctx carrying fn for the operation queued with it.
func WithPublishState(ctx context.Context, fn PublishStateFunc) context.Context {
	return context.WithValue(ctx, publishStateKey{}, fn)
}

// GetPublishState returns the PublishStateFunc ctx carries, or nil.
func GetPublishState(ctx context.Context) PublishStateFunc {
	fn, _ := ctx.Value(publishStateKey{}).(PublishStateFunc)
	return fn
}
