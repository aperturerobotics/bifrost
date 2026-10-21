package sobject_world_engine

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/aperturerobotics/util/csync"
	"github.com/sirupsen/logrus"
)

// slowHoldThreshold is how long a holder may keep the World writer lock
// before its release logs the holder's name and the hold. Every World write
// waits for the lock, so a long hold is what a slow commit's lock-wait
// stage waited on. Tests set it.
var slowHoldThreshold = time.Second

// writerLock is the World writer lock. Each holder names itself, so a slow
// hold is logged with what held it.
type writerLock struct {
	// mtx is the lock.
	mtx csync.Mutex
	// le logs slow holds, or nil to log none.
	le *logrus.Entry
}

// Lock takes the lock for holder and returns the function that releases it.
// The release logs the hold when it exceeded slowHoldThreshold; calls after
// the first do nothing.
func (l *writerLock) Lock(ctx context.Context, holder string) (func(), error) {
	// Take the lock and note when.
	unlock, err := l.mtx.Lock(ctx)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// Release it once, logging a slow hold.
	var released atomic.Bool
	return func() {
		if released.Swap(true) {
			return
		}
		held := time.Since(start)
		unlock()
		if held > slowHoldThreshold && l.le != nil {
			l.le.WithFields(logrus.Fields{
				"holder": holder,
				"held":   held,
			}).Warn("slow World writer lock hold")
		}
	}, nil
}
