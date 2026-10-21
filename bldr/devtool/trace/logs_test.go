package trace_test

import (
	"bytes"
	"context"
	stdtrace "runtime/trace"
	"testing"

	bldrtrace "github.com/s4wave/spacewave/bldr/devtool/trace"
)

func TestAnalyzeAndWriteLogs(t *testing.T) {
	// Record logs inside and outside a task.
	traceData := captureTrace(t, func() {
		// Log inside a task, twice for one message.
		ctx, task := stdtrace.NewTask(context.Background(), "log-task")
		stdtrace.Log(ctx, "startup", "skipped")
		stdtrace.Log(ctx, "startup", "skipped")
		stdtrace.Log(ctx, "startup=phase", "selected\nwith newline\tand tab")
		task.End()
		stdtrace.Log(context.Background(), "taskless", "recorded=without task")
	})

	// Analyze and check the task counts.
	result, err := bldrtrace.Analyze(bytes.NewReader(traceData), bldrtrace.Options{})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if result.TaskCount != 1 || result.IncompleteCount != 0 {
		t.Fatalf("task counts = %d total, %d incomplete; want 1 total, 0 incomplete", result.TaskCount, result.IncompleteCount)
	}

	// The logs are grouped and quoted.
	var output bytes.Buffer
	bldrtrace.WriteLogs(&output, result)
	want := "Runtime/trace logs:\n" +
		"     2  \"startup\"=\"skipped\"\n" +
		"     1  \"startup=phase\"=\"selected\\nwith newline\\tand tab\"\n" +
		"     1  \"taskless\"=\"recorded=without task\"\n"
	if output.String() != want {
		t.Fatalf("WriteLogs output = %q, want %q", output.String(), want)
	}
}

func TestWriteLogsEmpty(t *testing.T) {
	var output bytes.Buffer
	bldrtrace.WriteLogs(&output, &bldrtrace.TraceResult{})
	if got := output.String(); got != "Runtime/trace logs:\n(none)\n" {
		t.Fatalf("WriteLogs output = %q", got)
	}
}

func captureTrace(t *testing.T, fn func()) []byte {
	// Trace fn into memory.
	t.Helper()
	var output bytes.Buffer
	if err := stdtrace.Start(&output); err != nil {
		t.Fatalf("trace.Start: %v", err)
	}
	defer stdtrace.Stop()
	fn()
	stdtrace.Stop()
	return output.Bytes()
}
