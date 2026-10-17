package sdk_world_engine

import "time"

// SetSlowWriteHold sets the write hold after which a transaction logs its
// opener, and returns a function that restores the previous value.
func SetSlowWriteHold(d time.Duration) func() {
	prev := slowWriteHold
	slowWriteHold = d
	return func() { slowWriteHold = prev }
}
