package kernel

import (
	"context"
	"math"
	"sync"
)

// counterMax is the largest count a counter holds, as on Linux.
const counterMax = math.MaxUint64 - 1

// counter is the 64-bit count an eventfd or timerfd file reads. Writes and
// timer expirations add to it. A read waits for a nonzero count and takes it
// whole, or one at a time for a semaphore, and an add that would pass
// counterMax waits for a read.
type counter struct {
	// semaphore makes each read take one.
	semaphore bool

	// mtx guards the fields below.
	mtx sync.Mutex
	// value is the count.
	value uint64
	// waitCh closes when the count changes.
	waitCh chan struct{}
}

// newCounter returns a counter holding value.
func newCounter(value uint64, semaphore bool) *counter {
	return &counter{semaphore: semaphore, value: value, waitCh: make(chan struct{})}
}

// take waits for a nonzero count and takes it, or one for a semaphore. When
// ctx ends first it returns zero, so a done ctx makes a take that does not
// wait.
func (c *counter) take(ctx context.Context) uint64 {
	for {
		// Take the count when there is one.
		c.mtx.Lock()
		if n := c.value; n != 0 {
			if c.semaphore {
				n = 1
			}
			c.value -= n
			c.notifyLocked()
			c.mtx.Unlock()
			return n
		}
		waitCh := c.waitCh
		c.mtx.Unlock()

		// Wait for an add.
		if !wait(ctx, waitCh) {
			return 0
		}
	}
}

// add waits until n fits below counterMax, then adds it. It reports false
// when ctx ends first.
func (c *counter) add(ctx context.Context, n uint64) bool {
	for {
		// Add when the sum fits.
		c.mtx.Lock()
		if n <= counterMax-c.value {
			c.value += n
			if n != 0 {
				c.notifyLocked()
			}
			c.mtx.Unlock()
			return true
		}
		waitCh := c.waitCh
		c.mtx.Unlock()

		// Wait for a take.
		if !wait(ctx, waitCh) {
			return false
		}
	}
}

// reset sets the count to zero.
func (c *counter) reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.value != 0 {
		c.value = 0
		c.notifyLocked()
	}
}

// state reports whether a take, or an add of one when write is set, would
// proceed without waiting, and returns the channel that closes at the next
// change.
func (c *counter) state(write bool) (ready bool, change <-chan struct{}) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if write {
		return c.value < counterMax, c.waitCh
	}
	return c.value != 0, c.waitCh
}

// notifyLocked wakes the waiters. The caller holds mtx.
func (c *counter) notifyLocked() {
	close(c.waitCh)
	c.waitCh = make(chan struct{})
}
