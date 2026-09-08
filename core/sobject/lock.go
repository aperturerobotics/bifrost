package sobject

import (
	"context"

	"github.com/pkg/errors"
)

// SOStateLock is a lock handle to a SOState. The caller serializes all handle use.
type SOStateLock interface {
	// GetSOState returns the SOState as of when the lock was acquired.
	GetSOState() *SOState

	// WriteSOState writes an updated SOState and its accepted configuration changes.
	// Persistent providers commit the supplied lineage with the state atomically.
	// Returns an error if the lock was released already.
	// Only the locked handle should be able to write.
	WriteSOState(ctx context.Context, state *SOState, changes ...*SOConfigChange) error

	// Release releases the lock.
	Release()
}

// soStateLock implements SOStateLock.
type soStateLock struct {
	// initialState is the state observed while acquiring the provider lock.
	initialState *SOState
	// writeFn commits state and accepted lineage through the provider boundary.
	writeFn func(ctx context.Context, state *SOState, changes ...*SOConfigChange) error
	// release relinquishes the provider lock and its retained state reference.
	release func()
	// released prevents writes after relinquishing the provider lock.
	released bool
}

// GetSOState returns the SOState as of when the lock was acquired.
func (l *soStateLock) GetSOState() *SOState {
	return l.initialState
}

// WriteSOState writes an updated SOState.
// Returns an error if the lock was released already.
// Only the locked handle should be able to write.
func (l *soStateLock) WriteSOState(ctx context.Context, state *SOState, changes ...*SOConfigChange) error {
	if l.released {
		return errors.New("shared object state lock is released")
	}
	return l.writeFn(ctx, state, changes...)
}

// Release releases the lock.
func (l *soStateLock) Release() {
	if l.released {
		return
	}
	l.released = true
	if l.release != nil {
		l.release()
	}
}

// NewSOStateLock constructs a SOStateLock with an initial value and callbacks.
func NewSOStateLock(
	initialState *SOState,
	writeFn func(ctx context.Context, state *SOState, changes ...*SOConfigChange) error,
	release func(),
) SOStateLock {
	return &soStateLock{initialState: initialState, writeFn: writeFn, release: release}
}

// _ is a type assertion
var _ SOStateLock = (*soStateLock)(nil)
