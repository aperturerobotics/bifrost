package prom_plugin

import (
	"context"
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/s4wave/spacewave/db/world"
	world_testbed "github.com/s4wave/spacewave/db/world/testbed"
)

// testBaseTime is the timestamp of the first test sample in milliseconds.
const testBaseTime int64 = 1_700_000_000_000

// testStep is the scrape interval of the test samples in milliseconds.
const testStep int64 = 15_000

// fixture is a real World engine with several targets and an in-memory
// reference that receives every sample the stores do.
type fixture struct {
	// t is the running test.
	t *testing.T
	// eng is the World engine holding the targets.
	eng world.Engine
	// stores are the writers of the targets, by object key.
	stores map[string]*Store
	// keys lists the object keys in creation order.
	keys []string
	// ref is the in-memory reference holding the same samples.
	ref *referenceQueryable
}

// newFixture starts a World engine and creates one target per label set.
func newFixture(t *testing.T, targets ...labels.Labels) *fixture {
	// Start the World engine the stores write to.
	t.Helper()
	wtb := world_testbed.MustDefault(t, t.Context())
	f := &fixture{
		t:      t,
		eng:    wtb.Engine,
		stores: make(map[string]*Store, len(targets)),
		ref:    newReferenceQueryable(),
	}

	// Create each target as a World object.
	for i, target := range targets {
		key := "prom/target/" + string(rune('a'+i))
		st, err := CreateStore(t.Context(), f.eng, key, target)
		if err != nil {
			t.Fatal(err)
		}
		f.stores[key] = st
		f.keys = append(f.keys, key)
	}
	return f
}

// appendSamples appends the series' samples for scrape numbers from to to,
// exclusive, to the target's store and to the reference.
func (f *fixture) appendSamples(key string, series []labels.Labels, from, to int) {
	f.t.Helper()
	st := f.stores[key]
	for k := from; k < to; k++ {
		ts := testBaseTime + int64(k)*testStep
		for i, lset := range series {
			v := testValue(i, k)
			if err := st.Append(lset, ts, v); err != nil {
				f.t.Fatal(err)
			}
			f.ref.add(mergeLabels(st.meta, lset), ts, v)
		}
	}
}

// commit commits every store in one transaction.
func (f *fixture) commit(ctx context.Context) CommitStats {
	// Gather the stores in creation order.
	f.t.Helper()
	stores := make([]*Store, 0, len(f.keys))
	for _, key := range f.keys {
		stores = append(stores, f.stores[key])
	}

	// Commit them together.
	stats, err := CommitStores(ctx, f.eng, stores...)
	if err != nil {
		f.t.Fatal(err)
	}
	return stats
}

// queryable returns a new Queryable over all targets.
func (f *fixture) queryable() *Queryable {
	return NewQueryable(f.eng, StaticTargets(f.keys))
}

// testValue returns the value of series i at scrape k: a counter that grows
// by a rate that depends on i and resets once for series 1.
func testValue(i, k int) float64 {
	if i == 1 && k >= 30 {
		return float64((k - 30) * (i + 1) * 3)
	}
	return float64(k * (i + 1) * 3)
}
