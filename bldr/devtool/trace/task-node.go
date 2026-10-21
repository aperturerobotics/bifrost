package trace

import (
	"time"

	parser "github.com/s4wave/spacewave/bldr/devtool/trace/parser"
)

// TaskNode represents a single task instance in the critical path tree.
type TaskNode struct {
	Name      string
	Duration  time.Duration
	ParentPct float64       // duration / parent.Duration * 100
	TaskID    parser.TaskID // original task ID
	Children  []*TaskNode
}
