package prom_plugin

import (
	"math/rand/v2"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/prometheus/model/labels"
	world_testbed "github.com/s4wave/spacewave/db/world/testbed"
)

const (
	// benchSeries is the number of series one simulated target scrapes.
	benchSeries = 1000
	// benchScrapes is the number of scrapes in one simulated hour at 15 seconds.
	benchScrapes = 240
	// benchWindowScrapes is the number of scrapes one commit covers: one minute.
	benchWindowScrapes = 4
)

// BenchmarkHour commits one simulated hour of a target with 1,000 series
// scraped every 15 seconds, one commit per minute, and then queries six of the
// series over the last 15 minutes. It reports the bytes and blocks per commit
// and the query latency.
func BenchmarkHour(b *testing.B) {
	// Create one target in a real World engine.
	ctx := b.Context()
	wtb := world_testbed.MustDefault(b, ctx)
	target := labels.FromStrings("job", "bench", "instance", "bench:9100")
	st, err := CreateStore(ctx, wtb.Engine, "prom/target/bench", target)
	if err != nil {
		b.Fatal(err)
	}
	series := make([]labels.Labels, benchSeries)
	for i := range series {
		series[i] = labels.FromStrings("__name__", "bench_counter", "series", "s"+strconv.Itoa(1000+i))
	}

	// Scrape the hour, committing every window. Each series is a counter with
	// a random increment per scrape, so the chunks do not compress to nothing.
	rng := rand.New(rand.NewPCG(1, 2))
	counters := make([]float64, benchSeries)
	var commits int
	var firstBytes, sumBytes, maxBytes, sumBlocks, maxBlocks uint64
	for from := 0; from < benchScrapes; from += benchWindowScrapes {
		for k := from; k < from+benchWindowScrapes; k++ {
			ts := testBaseTime + int64(k)*testStep
			for i, lset := range series {
				counters[i] += float64(rng.IntN(1000)) / 8
				if err := st.Append(lset, ts, counters[i]); err != nil {
					b.Fatal(err)
				}
			}
		}
		stats, err := CommitStores(ctx, wtb.Engine, st)
		if err != nil {
			b.Fatal(err)
		}
		if commits == 0 {
			firstBytes = stats.Bytes
		}
		commits++
		sumBytes += stats.Bytes
		sumBlocks += stats.Blocks
		maxBytes = max(maxBytes, stats.Bytes)
		maxBlocks = max(maxBlocks, stats.Blocks)
	}
	b.Logf("commits=%d first commit bytes=%d, bytes/commit mean=%d max=%d blocks/commit mean=%d max=%d blocks written=%d bytes written=%d",
		commits, firstBytes, sumBytes/uint64(commits), maxBytes, sumBlocks/uint64(commits), maxBlocks, sumBlocks, sumBytes)

	// Query six series over the last 15 minutes.
	q := NewQueryable(wtb.Engine, StaticTargets{"prom/target/bench"})
	end := testBaseTime + (benchScrapes-1)*testStep
	start := end - 15*60*1000
	const six = `rate(bench_counter{series=~"s100[0-5]"}[1m])`
	b.Run("range-six-series-15m", func(b *testing.B) {
		for range b.N {
			res := rangeQuery(ctx, b, q, six, start, end, 15*time.Second)
			mat, err := res.Matrix()
			if err != nil || len(mat) != 6 {
				b.Fatalf("got %d series, %v", len(mat), err)
			}
		}
	})
	b.Run("instant-six-series", func(b *testing.B) {
		for range b.N {
			res := instantQuery(ctx, b, q, six, end)
			vec, err := res.Vector()
			if err != nil || len(vec) != 6 {
				b.Fatalf("got %d series, %v", len(vec), err)
			}
		}
	})
}
