package kernel

import (
	"sync"
	"time"
)

// intervalTimer expires once after a value and then every interval, as a
// timerfd's timer or a process's ITIMER_REAL does. Each expiration calls
// expire with the expirations since the last call: one, plus the periods a
// periodic timer missed, as on Linux.
type intervalTimer struct {
	// k is the kernel, whose clocks absolute times follow.
	k *Kernel
	// clockID is the clock absolute times follow.
	clockID uint32
	// expire receives each expiration's count. It runs with mtx held.
	expire func(n uint64)
	// clear drops what earlier expirations left when the timer is set
	// again, or is nil. It runs with mtx held.
	clear func()

	// mtx guards the fields below.
	mtx sync.Mutex
	// timer fires at due, or is nil while disarmed.
	timer *time.Timer
	// gen counts the host timers started, so a timer replaced while firing
	// expires nothing.
	gen uint64
	// due is the next expiration on the host clock.
	due time.Time
	// interval is the period, or zero for a timer that expires once.
	interval time.Duration
}

// newIntervalTimer returns a disarmed timer following clockID that reports
// its expirations to expire and drops them with clear, which may be nil.
func newIntervalTimer(k *Kernel, clockID uint32, expire func(n uint64), clear func()) *intervalTimer {
	return &intervalTimer{k: k, clockID: clockID, expire: expire, clear: clear}
}

// set arms the timer to expire after value, or at value on the timer's clock
// when abs is set, and then every interval. A zero value disarms it. It
// drops the earlier expirations and returns the old setting as get does.
func (t *intervalTimer) set(value, interval time.Duration, abs bool) (oldValue, oldInterval time.Duration) {
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
	if t.clear != nil {
		t.clear()
	}
	t.interval = interval
	if value != 0 {
		t.due = time.Now().Add(value)
		t.arm()
	}
	return oldValue, oldInterval
}

// get returns the time until the next expiration and the period, both zero
// while disarmed.
func (t *intervalTimer) get() (value, interval time.Duration) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	return t.getLocked()
}

// getLocked is get for a caller holding mtx.
func (t *intervalTimer) getLocked() (value, interval time.Duration) {
	if t.timer == nil {
		return 0, t.interval
	}
	return max(time.Until(t.due), 1), t.interval
}

// stop disarms the timer for good when its owner ends.
func (t *intervalTimer) stop() {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
}

// arm starts the host timer for due. The caller holds mtx.
func (t *intervalTimer) arm() {
	t.gen++
	gen := t.gen
	t.timer = time.AfterFunc(time.Until(t.due), func() { t.fire(gen) })
}

// fire reports the expirations of host timer gen, unless a set or stop
// replaced it, and rearms a periodic timer for its next period.
func (t *intervalTimer) fire(gen uint64) {
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
	t.expire(n)
}
