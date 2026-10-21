# bldr trace

`bldr trace` reads a Go runtime trace file and prints text-first timing analysis
for user-defined trace tasks. It was `tracetool` in the instrument repo. The
same command is `spacewave-dev trace` in an app built on spacewave.

It is designed for the cases where the default `go tool trace` UI is too heavy
or too interactive for the question at hand. Instead of hunting through the web
UI, you can get a compact hot-path summary directly in the terminal.

## What It Outputs

`bldr trace` produces two complementary views:

- A hot path table showing task name, count, total wall time, and
  min / p50 / p99 / max.
- A critical path tree showing how the slowest path is nested.

Those two views answer different questions:

- "What consumed the most time overall?"
- "Inside the slow operation, what was it actually blocked on?"

The table is useful for spotting amplification, contention, and heavy singleton
work. The tree is useful for finding the actual blocking chain inside an
end-to-end slow operation.

`bldr trace` streams the trace once and keeps only the recorded tasks, so memory
grows with the task count rather than the trace size. Use `-task` on large traces
to keep that count small.

If a task shows a large `(untraced)` gap, that usually means the instrumentation
inside that region is still too coarse and should be refined.

## Build

From the spacewave repo root:

```bash
go build -o ~/company/bin/bldr ./bldr/cmd/bldr
```

You can also run it directly without building:

```bash
go run ./bldr/cmd/bldr trace <trace-file>
```

## Usage

```text
bldr trace [flags] <trace-file>
```

Current flags:

- `-table`
  Show the hot path table. Default: `true`.
- `-tree`
  Show the critical path tree. Default: `true`.
- `-logs`
  Show runtime/trace log events grouped by category and message. Default: `false`.
- `-task string`
  Record only tasks whose name has the given prefix, plus their descendants.
  Child tasks keep their own names, so `-task "db/"` also reports the tasks a
  `db/` task starts.
- `-min duration`
  Only include task types with `max >= duration`.
- `-top int`
  Show the top N task types in the table. Default: `30`. Use `0` for all.
- `-all`
  Show all task types in the table, overriding the top-N limit.
- `-root string`
  Root the tree at the first task matching the given name prefix. By default,
  the tree is rooted at the trace's slowest critical-path task.

## Common Examples

Show the default summary:

```bash
bldr trace trace.out
```

Show only the table and include every task type:

```bash
bldr trace -all -tree=false trace.out
```

Focus on one subsystem by task prefix:

```bash
bldr trace -task "storage/" trace.out
```

Only show slower task types:

```bash
bldr trace -min 100ms trace.out
```

Show only the tree and root it at a specific task:

```bash
bldr trace -table=false -root "app/request" trace.out
```

Show only runtime/trace log events:

```bash
bldr trace -table=false -tree=false -logs trace.out
```

## How To Read The Output

### Hot path table

The table includes:

- `Count`
- `Total`
- `Min`
- `P50`
- `P99`
- `Max`

Typical patterns:

- High `Count` + high `Total`
  Usually amplification. The task is individually cheap but runs too many
  times.
- Small `P50` + much larger `P99` / `Max`
  Usually contention, queuing, or intermittent blocking.
- Low `Count` + high `P50`
  Usually one expensive operation that needs direct optimization.

### Critical path tree

The tree shows nested tasks with parent-relative percentages. Follow the
highest-percentage chain downward to find the actual blocking path.

This is especially useful when the table tells you a subsystem is expensive,
but you still need to know which nested child actually dominates the slow user
action or request.

## Why Use It

The main benefit of `bldr trace` is iteration speed.

It makes it easy to:

- inspect a trace without opening a browser UI
- compare before/after traces during optimization work
- paste findings into notes, issues, or reviews
- identify whether the problem is amplification, contention, one heavy commit,
  or missing instrumentation

It does not replace other profiling tools. It complements them by making Go
runtime trace data much easier to inspect in normal development workflows.
