package prom_plugin

import (
	"testing"
	"time"

	"github.com/prometheus/prometheus/model/labels"
)

// testTargets are the target labels of the several-target fixtures.
var testTargets = []labels.Labels{
	labels.FromStrings("job", "api", "instance", "a:9100"),
	labels.FromStrings("job", "api", "instance", "b:9100"),
	labels.FromStrings("job", "db", "instance", "c:9100"),
}

// testSeries are the series every test target scrapes.
var testSeries = []labels.Labels{
	labels.FromStrings("__name__", "http_requests_total", "method", "GET"),
	labels.FromStrings("__name__", "http_requests_total", "method", "POST"),
	labels.FromStrings("__name__", "http_requests_total", "method", "PUT"),
	labels.FromStrings("__name__", "temperature", "room", "kitchen"),
}

// TestQueryableMatchesReference checks selector, rate, and aggregation
// queries over several targets against the in-memory reference.
func TestQueryableMatchesReference(t *testing.T) {
	// Scrape three targets for 40 minutes and commit in 5 windows of 8 minutes.
	ctx := t.Context()
	f := newFixture(t, testTargets...)
	const scrapes = 160
	for from := 0; from < scrapes; from += 32 {
		for _, key := range f.keys {
			f.appendSamples(key, testSeries, from, from+32)
		}
		stats := f.commit(ctx)
		if stats.Segments != len(f.keys) {
			t.Fatalf("segments = %d, want one per target", stats.Segments)
		}
	}
	end := testBaseTime + (scrapes-1)*testStep
	start := end - 30*60*1000
	queries := []string{
		`http_requests_total`,
		`http_requests_total{job="api"}`,
		`http_requests_total{method=~"GET|POST", job!="db"}`,
		`{__name__=~"temp.*", instance="b:9100"}`,
		`http_requests_total{method!~"P.*"}`,
		`rate(http_requests_total[1m])`,
		`rate(http_requests_total[5m])`,
		`increase(http_requests_total{method="POST"}[10m])`,
		`irate(http_requests_total[2m])`,
		`sum(rate(http_requests_total[2m]))`,
		`sum by (job) (rate(http_requests_total[2m]))`,
		`sum without (method) (http_requests_total)`,
		`avg(temperature)`,
		`max by (instance) (temperature)`,
		`count by (job) (http_requests_total)`,
		`topk(2, rate(http_requests_total[3m]))`,
		`avg_over_time(temperature[4m])`,
		`temperature > 20`,
		`sum(rate(http_requests_total[2m])) / count(temperature)`,
		`http_requests_total offset 5m`,
	}

	// Run each query as an instant and a range query against both storages.
	q := f.queryable()
	for _, qs := range queries {
		got := instantQuery(ctx, t, q, qs, end)
		want := instantQuery(ctx, t, f.ref, qs, end)
		requireSameResult(t, qs, got, want)
		if got.Value.String() == "" || got.Value.String() == "{}" {
			t.Fatalf("%s: empty result, the query tests nothing", qs)
		}
		gotRange := rangeQuery(ctx, t, q, qs, start, end, 30*time.Second)
		wantRange := rangeQuery(ctx, t, f.ref, qs, start, end, 30*time.Second)
		requireSameResult(t, qs+" (range)", gotRange, wantRange)
	}
}

// TestQueryableLabelMatching checks matcher types and label lookups.
func TestQueryableLabelMatching(t *testing.T) {
	// Commit one window for two targets.
	ctx := t.Context()
	f := newFixture(t, testTargets[0], testTargets[2])
	for _, key := range f.keys {
		f.appendSamples(key, testSeries, 0, 8)
	}
	f.commit(ctx)
	end := testBaseTime + 7*testStep
	q := f.queryable()

	// Count the series each selector returns.
	cases := []struct {
		selector string
		want     int
	}{
		{`http_requests_total`, 6},
		{`http_requests_total{job="api"}`, 3},
		{`http_requests_total{job="db", method="PUT"}`, 1},
		{`http_requests_total{job!="api"}`, 3},
		{`http_requests_total{method=~"GET|PUT"}`, 4},
		{`http_requests_total{method!~"GET|PUT"}`, 2},
		{`{room="kitchen"}`, 2},
		{`{__name__=~".+", job="db"}`, 4},
		{`http_requests_total{room=""}`, 6},
		{`http_requests_total{room!=""}`, 0},
		{`http_requests_total{instance="nope"}`, 0},
		{`nothing`, 0},
	}
	for _, c := range cases {
		res := instantQuery(ctx, t, q, c.selector, end)
		if res.Err != nil {
			t.Fatalf("%s: %v", c.selector, res.Err)
		}
		vec, err := res.Vector()
		if err != nil {
			t.Fatalf("%s: %v", c.selector, err)
		}
		if len(vec) != c.want {
			t.Fatalf("%s: %d series, want %d", c.selector, len(vec), c.want)
		}
	}

	// Look up label names the way a panel editor does.
	querier, err := q.Querier(testBaseTime, end)
	if err != nil {
		t.Fatal(err)
	}
	defer querier.Close()
	names, _, err := querier.LabelNames(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantNames := []string{"__name__", "instance", "job", "method", "room"}
	if !equalStrings(names, wantNames) {
		t.Fatalf("label names = %v, want %v", names, wantNames)
	}

	// Look up label values under a selector.
	methods, _, err := querier.LabelValues(ctx, "method", nil, labels.MustNewMatcher(labels.MatchEqual, "job", "db"))
	if err != nil {
		t.Fatal(err)
	}
	if !equalStrings(methods, []string{"GET", "POST", "PUT"}) {
		t.Fatalf("method values = %v", methods)
	}
}

// TestQueryableMissingTarget checks that a listed target without an object is
// skipped instead of failing the query.
func TestQueryableMissingTarget(t *testing.T) {
	// Commit one target and list it beside a target that does not exist.
	ctx := t.Context()
	f := newFixture(t, testTargets[0])
	f.appendSamples(f.keys[0], testSeries, 0, 4)
	f.commit(ctx)
	q := NewQueryable(f.eng, StaticTargets{"prom/target/missing", f.keys[0]})

	// The query answers from the target that exists.
	res := instantQuery(ctx, t, q, `temperature`, testBaseTime+3*testStep)
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	vec, _ := res.Vector()
	if len(vec) != 1 {
		t.Fatalf("got %d series, want 1", len(vec))
	}
}

// equalStrings reports whether a and b hold the same strings in order.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
