package prom_plugin

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
)

// newTestEngine returns a PromQL engine with Prometheus defaults.
func newTestEngine() *promql.Engine {
	return promql.NewEngine(promql.EngineOpts{
		MaxSamples:    10_000_000,
		Timeout:       30 * time.Second,
		LookbackDelta: 5 * time.Minute,
	})
}

// instantQuery runs qs at time ts, in milliseconds, against q.
func instantQuery(ctx context.Context, t testing.TB, q storage.Queryable, qs string, ts int64) *promql.Result {
	t.Helper()
	query, err := newTestEngine().NewInstantQuery(ctx, q, nil, qs, time.UnixMilli(ts))
	if err != nil {
		t.Fatalf("parse %q: %v", qs, err)
	}
	// Close would return the result's point slices to the engine's pool,
	// where a later query overwrites them, so the result stays unclosed.
	return query.Exec(ctx)
}

// rangeQuery runs qs from start to end, in milliseconds, at step against q.
func rangeQuery(ctx context.Context, t testing.TB, q storage.Queryable, qs string, start, end int64, step time.Duration) *promql.Result {
	t.Helper()
	query, err := newTestEngine().NewRangeQuery(ctx, q, nil, qs, time.UnixMilli(start), time.UnixMilli(end), step)
	if err != nil {
		t.Fatalf("parse %q: %v", qs, err)
	}
	// Close would return the result's point slices to the engine's pool,
	// where a later query overwrites them, so the result stays unclosed.
	return query.Exec(ctx)
}

// requireSameResult fails t unless got and want hold the same value.
func requireSameResult(t *testing.T, qs string, got, want *promql.Result) {
	// Fail on an error from either side.
	t.Helper()
	if got.Err != nil || want.Err != nil {
		t.Fatalf("%s: got error %v, reference error %v", qs, got.Err, want.Err)
	}

	// Compare the printed values line by line.
	gotLines := strings.Split(got.Value.String(), "\n")
	wantLines := strings.Split(want.Value.String(), "\n")
	for i := range max(len(gotLines), len(wantLines)) {
		var g, w string
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if g != w {
			t.Fatalf("%s: line %d differs from reference (%d vs %d lines)\n got: %s\nwant: %s", qs, i, len(gotLines), len(wantLines), g, w)
		}
	}
}
