package trace

import "time"

// TaskTypeStats holds aggregated statistics for all instances of a task type.
type TaskTypeStats struct {
	Name     string
	Count    int
	Complete int
	Total    time.Duration
	Min      time.Duration
	Max      time.Duration
	P50      time.Duration
	P99      time.Duration
}
