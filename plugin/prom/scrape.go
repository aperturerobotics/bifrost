package prom_plugin

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/textparse"
	s4wave_prom "github.com/s4wave/spacewave/sdk/prom"
	"github.com/sirupsen/logrus"
)

// maxScrapeBytes bounds the body of one scrape.
const maxScrapeBytes = 64 << 20

// fallbackContentType parses a page that sends no usable Content-Type as the
// Prometheus text format.
const fallbackContentType = "text/plain"

// scrapeSample is one parsed sample of a scrape, before it is appended.
type scrapeSample struct {
	// lset is the series' labels without the target's labels.
	lset labels.Labels
	// v is the sample value.
	v float64
}

// scraper scrapes one target's exporter into the target's Store on the
// target's interval. It appends samples only; a commitGroup commits them.
type scraper struct {
	// le is the logger.
	le *logrus.Entry
	// store is the target's writer.
	store *Store
	// cfg configures the scrape.
	cfg *s4wave_prom.ScrapeConfig
	// peerID is the Device peer reported in the status.
	peerID string
	// client fetches the metrics page.
	client *http.Client
	// last holds the series the previous successful scrape returned, by hash.
	last map[uint64]labels.Labels
}

// newScraper constructs a scraper of the target store reads into.
func newScraper(le *logrus.Entry, store *Store, cfg *s4wave_prom.ScrapeConfig, peerID string) *scraper {
	return &scraper{
		le:     le,
		store:  store,
		cfg:    cfg,
		peerID: peerID,
		client: &http.Client{Timeout: cfg.Timeout()},
		last:   make(map[uint64]labels.Labels),
	}
}

// run scrapes now and then once per interval until ctx ends.
func (s *scraper) run(ctx context.Context) error {
	t := time.NewTicker(s.cfg.Interval())
	defer t.Stop()
	for {
		s.scrape(ctx, time.Now())
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-t.C:
		}
	}
}

// scrape fetches and appends one scrape stamped at start. A failed scrape
// appends up as 0 and marks every series of the last successful scrape stale;
// a successful one marks stale the series that vanished since then.
func (s *scraper) scrape(ctx context.Context, start time.Time) {
	// Fetch and parse the whole page before appending any of it.
	ts := start.UnixMilli()
	samples, err := s.fetch(ctx)
	duration := time.Since(start)
	if ctx.Err() != nil {
		return
	}

	// Append the scraped samples, renaming labels the target sets.
	seen := make(map[uint64]labels.Labels, len(samples))
	var skipped int
	for _, smp := range samples {
		if err := s.store.Append(smp.lset, ts, smp.v); err != nil {
			skipped++
			continue
		}
		seen[smp.lset.Hash()] = smp.lset
	}
	if skipped != 0 {
		s.le.WithField("skipped", skipped).Debug("skipped out of order or duplicate samples")
	}

	// Mark stale each series the last successful scrape had and this one lacks.
	for h, lset := range s.last {
		if _, ok := seen[h]; !ok {
			_ = s.store.AppendStale(lset, ts)
		}
	}
	if err == nil {
		s.last = seen
	} else {
		s.last = make(map[uint64]labels.Labels)
	}

	// Append the scrape's own series and record its status.
	up := 1.0
	status := &s4wave_prom.TargetStatus{
		Up:             err == nil,
		LastScrapeMs:   ts,
		LastDurationMs: uint32(min(duration.Milliseconds(), int64(^uint32(0)))), //nolint:gosec // clamped above.
		Series:         uint64(len(seen)),
		DevicePeerId:   s.peerID,
	}
	if err != nil {
		up = 0
		status.LastError = err.Error()
		s.le.WithError(err).Debug("scrape failed")
	}
	for name, v := range map[string]float64{
		"up":                      up,
		"scrape_duration_seconds": duration.Seconds(),
		"scrape_samples_scraped":  float64(len(samples)),
	} {
		_ = s.store.Append(labels.FromStrings(model.MetricNameLabel, name), ts, v)
	}
	if err := s.store.SetStatus(status); err != nil {
		s.le.WithError(err).Warn("encode target status")
	}
}

// fetch reads and parses the target's metrics page.
func (s *scraper) fetch(ctx context.Context) ([]scrapeSample, error) {
	// Request the page within the scrape timeout.
	ctx, cancel := context.WithTimeout(ctx, s.cfg.Timeout())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.GetUrl(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/plain;version=0.0.4;q=0.9,*/*;q=0.1")

	// Send it and require a page.
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("server returned HTTP status %s", resp.Status)
	}

	// Read the bounded body.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxScrapeBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxScrapeBytes {
		return nil, errors.Errorf("metrics page exceeds %d bytes", maxScrapeBytes)
	}
	return s.parse(body, resp.Header.Get("Content-Type"))
}

// parse decodes the float samples of a metrics page. Native histograms and
// metadata entries are skipped; exporter timestamps are ignored.
func (s *scraper) parse(body []byte, contentType string) ([]scrapeSample, error) {
	// Choose the parser from the content type.
	p, err := textparse.New(body, contentType, nil, textparse.ParserOptions{FallbackContentType: fallbackContentType})
	if p == nil {
		return nil, errors.Wrap(err, "metrics page format")
	}

	// Collect each float sample with its labels.
	var out []scrapeSample
	for {
		entry, err := p.Next()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, errors.Wrap(err, "parse metrics page")
		}
		if entry != textparse.EntrySeries {
			continue
		}
		_, _, v := p.Series()
		var lset labels.Labels
		p.Labels(&lset)
		out = append(out, scrapeSample{lset: s.exportedLabels(lset), v: v})
	}
}

// exportedLabels renames each label of lset that the target also sets to
// exported_<name>, as Prometheus does without honor_labels.
func (s *scraper) exportedLabels(lset labels.Labels) labels.Labels {
	// Keep the set when no name clashes.
	meta := s.store.meta
	clash := false
	lset.Range(func(l labels.Label) { clash = clash || meta.Has(l.Name) })
	if !clash {
		return lset
	}

	// Rebuild it with the clashing names prefixed.
	b := labels.NewScratchBuilder(lset.Len())
	lset.Range(func(l labels.Label) {
		name := l.Name
		if meta.Has(name) {
			name = "exported_" + name
		}
		b.Add(name, l.Value)
	})
	b.Sort()
	return b.Labels()
}
