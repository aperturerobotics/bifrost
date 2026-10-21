package trace

import (
	"fmt"
	"io"
)

// LogStats holds aggregated statistics for runtime/trace log events.
type LogStats struct {
	Category string
	Message  string
	Count    int
}

// WriteLogs writes runtime/trace log events grouped by category and message.
func WriteLogs(w io.Writer, result *TraceResult) {
	fmt.Fprintln(w, "Runtime/trace logs:")
	if len(result.Logs) == 0 {
		fmt.Fprintln(w, "(none)")
		return
	}

	for _, log := range result.Logs {
		fmt.Fprintf(w, "%6d  %q=%q\n", log.Count, log.Category, log.Message)
	}
}
