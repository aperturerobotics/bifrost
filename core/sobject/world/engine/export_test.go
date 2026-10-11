package sobject_world_engine

import "time"

// SetSlowCommitThreshold sets the duration after which a World write logs its
// stage times and returns a function that restores the previous one.
func SetSlowCommitThreshold(threshold time.Duration) func() {
	previous := slowCommitThreshold
	slowCommitThreshold = threshold
	return func() { slowCommitThreshold = previous }
}
