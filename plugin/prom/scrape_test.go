package prom_plugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/s4wave/spacewave/db/world"
	world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	s4wave_prom "github.com/s4wave/spacewave/sdk/prom"
	"github.com/sirupsen/logrus"
)

// TestScrapeRunThrough scrapes a live exporter into a target, commits, and
// reads the result back through the status and the PromQL service: a good
// scrape, a failed one that marks its series stale, and a reopened store
// that resumes the history.
func TestScrapeRunThrough(t *testing.T) {
	// Share the context and logger.
	ctx := t.Context()
	le := logrus.NewEntry(logrus.New())

	// Serve an exporter that can be made to fail; its job label clashes with
	// the target's.
	var failing atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if failing.Load() {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte("# TYPE node_load1 gauge\nnode_load1{job=\"inner\"} 0.5\nnode_disk_bytes{dev=\"sda\"} 1024\n"))
	}))
	defer srv.Close()

	// Create the target and its query service on a real World engine.
	wtb := world_testbed.MustDefault(t, ctx)
	eng := wtb.Engine
	key := "prom/target/node"
	cfg := &s4wave_prom.ScrapeConfig{Url: srv.URL}

	// Create the target.
	store, err := CreateStore(ctx, eng, key, labels.FromStrings("job", "node", "instance", "forge"))
	if err != nil {
		t.Fatal(err)
	}

	// Serve PromQL over it, with helpers to query and commit.
	svc := &queryService{engine: newQueryEngine(), q: NewQueryable(eng, worldTargets{eng: eng})}
	query := func(expr string, at time.Time) *s4wave_prom.QueryResponse {
		t.Helper()
		resp, err := svc.Query(ctx, &s4wave_prom.QueryRequest{Expr: expr, TimeMs: at.UnixMilli()})
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}
	commit := func(stores ...*Store) {
		t.Helper()
		if _, err := CommitStores(ctx, eng, stores...); err != nil {
			t.Fatal(err)
		}
	}

	// A good scrape stores the page, the scrape series and an up status.
	t0 := time.Now().Add(-time.Minute).Truncate(time.Second)
	sc := newScraper(le, store, cfg, "peer-a")
	sc.scrape(ctx, t0)
	commit(store)
	_, status := readTarget(ctx, t, eng, key)
	if !status.GetUp() || status.GetSeries() != 2 || status.GetDevicePeerId() != "peer-a" || status.GetLastScrapeMs() != t0.UnixMilli() {
		t.Fatalf("status after good scrape: %v", status)
	}
	if v := singleValue(t, query(`up{job="node",instance="forge"}`, t0)); v != 1 {
		t.Fatalf("up = %v", v)
	}
	if v := singleValue(t, query(`node_load1{exported_job="inner"}`, t0)); v != 0.5 {
		t.Fatalf("node_load1 = %v", v)
	}

	// A failed scrape records the error, sets up to 0, and ends the series.
	failing.Store(true)
	t1 := t0.Add(15 * time.Second)
	sc.scrape(ctx, t1)
	commit(store)
	_, status = readTarget(ctx, t, eng, key)
	if status.GetUp() || status.GetLastError() == "" {
		t.Fatalf("status after failed scrape: %v", status)
	}
	if v := singleValue(t, query("up", t1)); v != 0 {
		t.Fatalf("up after failure = %v", v)
	}
	if n := len(query("node_disk_bytes", t1).GetSeries()); n != 0 {
		t.Fatalf("stale series still answered: %d", n)
	}

	// A reopened store resumes the history after the exporter recovers.
	failing.Store(false)
	reopened, err := OpenStore(ctx, eng, key)
	if err != nil {
		t.Fatal(err)
	}

	// Scrape through it and read up over the whole run.
	t2 := t1.Add(15 * time.Second)
	newScraper(le, reopened, cfg, "peer-b").scrape(ctx, t2)
	commit(reopened)
	resp, err := svc.Query(ctx, &s4wave_prom.QueryRequest{
		Expr:    "up",
		StartMs: t0.UnixMilli(),
		EndMs:   t2.UnixMilli(),
		StepMs:  15_000,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Expect up, down, up.
	var got []float64
	for _, smp := range resp.GetSeries()[0].GetSamples() {
		got = append(got, smp.GetValue())
	}
	if len(got) != 3 || got[0] != 1 || got[1] != 0 || got[2] != 1 {
		t.Fatalf("up over the run = %v", got)
	}
}

// readTarget reads the target's meta and status in a read transaction.
func readTarget(ctx context.Context, t *testing.T, eng world.Engine, key string) (*s4wave_prom.TargetMeta, *s4wave_prom.TargetStatus) {
	// Read both in one transaction.
	t.Helper()
	var meta *s4wave_prom.TargetMeta
	var status *s4wave_prom.TargetStatus
	err := world.ExecTransaction(ctx, eng, false, func(ctx context.Context, tx world.WorldState) error {
		var err error
		meta, status, err = s4wave_prom.ReadTarget(ctx, tx, key)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return meta, status
}

// singleValue returns the value of a one-series instant result.
func singleValue(t *testing.T, resp *s4wave_prom.QueryResponse) float64 {
	t.Helper()
	if len(resp.GetSeries()) != 1 {
		t.Fatalf("want one series, got %v", resp.GetSeries())
	}
	return resp.GetSeries()[0].GetSamples()[0].GetValue()
}
