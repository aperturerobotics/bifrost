package kernel

import (
	"context"
	"sync"
	"time"
)

// timerFile is a timerfd's timer. Each expiration adds one to the file's
// counter, and a periodic timer adds the periods it missed, as on Linux.
type timerFile struct {
	// k is the kernel, whose clocks absolute times follow.
	k *Kernel
	// clockID is the clock the timer follows.
	clockID uint32
	// count is the expirations not yet read.
	count *counter

	// mtx guards the fields below.
	mtx sync.Mutex
	// timer fires at due, or is nil while disarmed.
	timer *time.Timer
	// gen counts the host timers started, so a timer replaced while firing
	// counts nothing.
	gen uint64
	// due is the next expiration on the host clock.
	due time.Time
	// interval is the period, or zero for a timer that expires once.
	interval time.Duration
}

// newTimerFile returns a disarmed timer following clockID.
func newTimerFile(k *Kernel, clockID uint32) *timerFile {
	return &timerFile{k: k, clockID: clockID, count: newCounter(0, false)}
}

// set arms the timer to expire after value, or at value on the timer's clock
// when abs is set, and then every interval. A zero value disarms it. It
// drops the unread expirations and returns the old setting as get does.
func (t *timerFile) set(value, interval time.Duration, abs bool) (oldValue, oldInterval time.Duration) {
	// Convert an absolute time to the wait until it.
	if abs && value != 0 {
		now, _ := t.k.nowForClock(t.clockID)
		value = max(time.Unix(0, 0).Add(value).Sub(now), 1)
	}

	// Replace the setting.
	t.mtx.Lock()
	defer t.mtx.Unlock()
	oldValue, oldInterval = t.getLocked()
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
	t.count.reset()
	t.interval = interval
	if value != 0 {
		t.due = time.Now().Add(value)
		t.arm()
	}
	return oldValue, oldInterval
}

// get returns the time until the next expiration and the period, both zero
// while disarmed.
func (t *timerFile) get() (value, interval time.Duration) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	return t.getLocked()
}

// getLocked is get for a caller holding mtx.
func (t *timerFile) getLocked() (value, interval time.Duration) {
	if t.timer == nil {
		return 0, t.interval
	}
	return max(time.Until(t.due), 1), t.interval
}

// stop disarms the timer for good when its file closes.
func (t *timerFile) stop() {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
}

// arm starts the host timer for due. The caller holds mtx.
func (t *timerFile) arm() {
	t.gen++
	gen := t.gen
	t.timer = time.AfterFunc(time.Until(t.due), func() { t.fire(gen) })
}

// fire counts the expirations of host timer gen, unless a set or stop
// replaced it, and rearms a periodic timer for its next period.
func (t *timerFile) fire(gen uint64) {
	// Count the expiration and the periods missed since.
	t.mtx.Lock()
	defer t.mtx.Unlock()
	if t.timer == nil || t.gen != gen {
		return
	}
	n := uint64(1)
	if t.interval == 0 {
		t.timer = nil
	} else {
		missed := time.Since(t.due) / t.interval
		n += uint64(missed)
		t.due = t.due.Add((missed + 1) * t.interval)
		t.arm()
	}

	// Add them without waiting: a full count drops them.
	t.count.add(nonblocking(context.Background()), n)
}
