package trace

import "time"

// TraceResult holds the parsed and analyzed trace data.
type TraceResult struct {
	// Duration is the total trace duration.
	Duration time.Duration
	// TaskCount is the total number of task instances.
	TaskCount int
	// IncompleteCount is the number of tasks without an End event.
	IncompleteCount int
	// TypeStats holds per-task-type aggregated statistics, sorted by TotalTime desc.
	TypeStats []*TaskTypeStats
	// Logs holds runtime/trace log events grouped by category and message.
	Logs []*LogStats
	// CriticalPath is the tree rooted at the slowest complete root task.
	CriticalPath *TaskNode
}
