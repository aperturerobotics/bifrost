package trace_test

import (
	"bytes"
	"context"
	stdtrace "runtime/trace"
	"testing"

	bldrtrace "github.com/s4wave/spacewave/bldr/devtool/trace"
)

func TestAnalyzeTaskPrefixKeepsDescendants(t *testing.T) {
	// Record a db task with a child and an unrelated task.
	traceData := captureTrace(t, func() {
		// Start a db task with one child.
		ctx, commit := stdtrace.NewTask(context.Background(), "db/commit")
		_, sync := stdtrace.NewTask(ctx, "fsync")
		sync.End()
		commit.End()

		// Start a task the filter drops.
		_, other := stdtrace.NewTask(context.Background(), "http/request")
		other.End()
	})

	// Analyze with the db prefix and check only db tasks were recorded.
	result, err := bldrtrace.Analyze(bytes.NewReader(traceData), bldrtrace.Options{TaskPrefix: "db/"})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if result.TaskCount != 2 {
		t.Fatalf("TaskCount = %d, want 2", result.TaskCount)
	}
	for _, stats := range result.TypeStats {
		if stats.Name == "http/request" {
			t.Fatalf("filtered task recorded: %#v", stats)
		}
	}

	// The critical path keeps the child.
	root := result.CriticalPath
	if root == nil || root.Name != "db/commit" {
		t.Fatalf("CriticalPath = %#v, want db/commit", root)
	}
	if len(root.Children) != 1 || root.Children[0].Name != "fsync" {
		t.Fatalf("CriticalPath children = %#v, want fsync", root.Children)
	}
}
