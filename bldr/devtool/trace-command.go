//go:build !js

package devtool

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aperturerobotics/cli"
	"github.com/pkg/errors"
	bldr_trace "github.com/s4wave/spacewave/bldr/devtool/trace"
)

// traceArgs are the flags of the trace command.
type traceArgs struct {
	table   bool
	tree    bool
	logs    bool
	task    string
	minDur  time.Duration
	top     int
	root    string
	showAll bool
}

// BuildTraceCommand builds the trace command.
func BuildTraceCommand() *cli.Command {
	args := &traceArgs{table: true, tree: true, top: 30}
	return &cli.Command{
		Name:      "trace",
		Usage:     "summarize the runtime/trace tasks in a Go runtime trace",
		ArgsUsage: "<trace-file>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "table",
				Usage:       "show the hot path table",
				Value:       args.table,
				Destination: &args.table,
			},
			&cli.BoolFlag{
				Name:        "tree",
				Usage:       "show the critical path tree",
				Value:       args.tree,
				Destination: &args.tree,
			},
			&cli.BoolFlag{
				Name:        "logs",
				Usage:       "show runtime/trace log events",
				Destination: &args.logs,
			},
			&cli.StringFlag{
				Name:        "task",
				Usage:       "record only tasks matching name prefix and their descendants",
				Destination: &args.task,
			},
			&cli.DurationFlag{
				Name:        "min",
				Usage:       "show only task types whose max is at least this duration",
				Destination: &args.minDur,
			},
			&cli.IntFlag{
				Name:        "top",
				Usage:       "top N task types in the table (0 = all)",
				Value:       args.top,
				Destination: &args.top,
			},
			&cli.StringFlag{
				Name:        "root",
				Usage:       "root task name prefix for the tree (default: slowest)",
				Destination: &args.root,
			},
			&cli.BoolFlag{
				Name:        "all",
				Usage:       "show all task types (no top-N limit)",
				Destination: &args.showAll,
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return errors.New("usage: trace [flags] <trace-file>")
			}
			return args.run(os.Stdout, c.Args().First())
		},
	}
}

// run analyzes the trace file and writes the selected views to w.
func (a *traceArgs) run(w io.Writer, path string) error {
	// Analyze the trace in one streaming pass.
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	result, err := bldr_trace.Analyze(f, bldr_trace.Options{TaskPrefix: a.task})
	if err != nil {
		return err
	}
	if a.minDur > 0 {
		result = filterTraceByMinDur(result, a.minDur)
	}

	// Write the header.
	fmt.Fprintf(w, "Trace: %s (%s, %d tasks, %d incomplete)\n\n",
		path, fmtTraceDur(result.Duration), result.TaskCount, result.IncompleteCount)
	if result.TaskCount == 0 {
		fmt.Fprintln(w, "Warning: trace contains no runtime/trace tasks; if the workload forks, capture the child process with -test.trace.")
		fmt.Fprintln(w)
	}

	// Write the selected views.
	top := a.top
	if a.showAll {
		top = 0
	}
	if a.table {
		bldr_trace.WriteTable(w, result, top)
	}
	if a.table && a.tree {
		fmt.Fprintln(w)
	}
	if a.tree {
		root := result.CriticalPath
		if a.root != "" {
			root = findTraceRootByName(result, a.root)
		}
		bldr_trace.WriteTree(w, root)
	}
	if a.logs {
		if a.table || a.tree {
			fmt.Fprintln(w)
		}
		bldr_trace.WriteLogs(w, result)
	}
	return nil
}

// filterTraceByMinDur keeps the task types whose Max is at least minDur.
func filterTraceByMinDur(r *bldr_trace.TraceResult, minDur time.Duration) *bldr_trace.TraceResult {
	// Copy the result with only the slow enough types.
	var filtered []*bldr_trace.TaskTypeStats
	for _, s := range r.TypeStats {
		if s.Max >= minDur {
			filtered = append(filtered, s)
		}
	}
	result := *r
	result.TypeStats = filtered
	return &result
}

// findTraceRootByName returns the first critical path task with the name prefix.
func findTraceRootByName(r *bldr_trace.TraceResult, name string) *bldr_trace.TaskNode {
	if r.CriticalPath == nil {
		return nil
	}
	if strings.HasPrefix(r.CriticalPath.Name, name) {
		return r.CriticalPath
	}
	return findTraceNodeByName(r.CriticalPath, name)
}

// findTraceNodeByName returns the first descendant whose name has the prefix.
func findTraceNodeByName(node *bldr_trace.TaskNode, name string) *bldr_trace.TaskNode {
	for _, child := range node.Children {
		if strings.HasPrefix(child.Name, name) {
			return child
		}
		if found := findTraceNodeByName(child, name); found != nil {
			return found
		}
	}
	return nil
}

// fmtTraceDur formats the trace duration in milliseconds or seconds.
func fmtTraceDur(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1e6)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
