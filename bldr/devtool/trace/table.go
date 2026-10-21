package trace

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// WriteTable writes the hot path table to w.
// Shows at most topN task types (0 = no limit).
func WriteTable(w io.Writer, result *TraceResult, topN int) {
	// Keep the top N task types.
	stats := result.TypeStats
	if topN > 0 && len(stats) > topN {
		stats = stats[:topN]
	}

	// Compute column widths.
	nameW := len("Task")
	for _, s := range stats {
		if len(s.Name) > nameW {
			nameW = len(s.Name)
		}
	}
	if nameW > 55 {
		nameW = 55
	}

	// Header.
	fmt.Fprintf(w, "%-*s  %7s  %8s  %8s  %8s  %8s  %8s\n",
		nameW, "Task", "Count", "Total", "Min", "P50", "P99", "Max")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", nameW+2+7+2+8+2+8+2+8+2+8+2+8))

	// Rows.
	for _, s := range stats {
		name := s.Name
		if len(name) > nameW {
			name = name[:nameW-3] + "..."
		}
		if s.Complete == 0 {
			fmt.Fprintf(w, "%-*s  %7d  %8s  %8s  %8s  %8s  %8s\n",
				nameW, name, s.Count, "-", "-", "-", "-", "-")
			continue
		}
		fmt.Fprintf(w, "%-*s  %7d  %8s  %8s  %8s  %8s  %8s\n",
			nameW, name, s.Count,
			fmtDur(s.Total), fmtDur(s.Min), fmtDur(s.P50), fmtDur(s.P99), fmtDur(s.Max))
	}
}

// fmtDur formats a duration with auto-scaled units.
func fmtDur(d time.Duration) string {
	if d == 0 {
		return "0"
	}
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		us := float64(d.Nanoseconds()) / 1000
		if us < 10 {
			return fmt.Sprintf("%.1fus", us)
		}
		return fmt.Sprintf("%.0fus", us)
	case d < time.Second:
		ms := float64(d.Nanoseconds()) / 1e6
		if ms < 10 {
			return fmt.Sprintf("%.1fms", ms)
		}
		return fmt.Sprintf("%.0fms", ms)
	default:
		s := d.Seconds()
		if s < 10 {
			return fmt.Sprintf("%.2fs", s)
		}
		if s < 100 {
			return fmt.Sprintf("%.1fs", s)
		}
		return fmt.Sprintf("%.0fs", s)
	}
}
