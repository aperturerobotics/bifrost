package volume_controller

import (
	"context"
	"time"

	block_gc "github.com/s4wave/spacewave/db/block/gc"
)

// gcDelegator is implemented by a volume that views another controller's
// storage. That controller sees every owner's references and sweeps through the
// storage's own block store, so this controller tracks references and never
// collects.
type gcDelegator interface {
	// DelegatesGC marks the volume as a view of collected storage.
	DelegatesGC()
}

// storageMaintainer is implemented by a volume whose storage reclaims the
// space of removed blocks in a separate step.
type storageMaintainer interface {
	// Maintenance reclaims the space of removed blocks.
	Maintenance(ctx context.Context) error
}

// markIntervals is the number of collection intervals between block passes.
const markIntervals = 15

// runGCSweep runs the periodic GC sweep goroutine. It waits for the volume to
// become ready, then collects its reference graph on each interval.
func (c *Controller) runGCSweep(ctx context.Context) error {
	// Choose the configured collection interval or disable volume collection.
	interval, err := c.config.ParseGCIntervalDur()
	if err != nil {
		c.le.WithError(err).Warn("invalid gc_interval_dur, using default")
		interval = defaultGCInterval
	}
	if interval == 0 {
		c.le.Debug("gc disabled (interval=0)")
		return nil
	}

	// Wait for the volume before choosing its collection mechanism.
	vol, err := c.GetVolume(ctx)
	if err != nil {
		return err
	}

	// Leave collection to the controller of the storage a view reads.
	if _, ok := vol.(gcDelegator); ok {
		c.le.Debug("volume delegates gc, sweep disabled")
		return nil
	}

	// Require a reference graph for collection.
	rg := vol.GetRefGraph()
	if rg == nil {
		c.le.Debug("volume has no ref graph, gc sweep disabled")
		return nil
	}

	// Construct the collector for the volume reference graph. A block pass
	// reads the whole graph, so it runs every markIntervals intervals.
	collector := block_gc.NewCollector(rg, vol, nil)
	collector.SetMarkInterval(markIntervals * interval)
	c.le.WithField("interval", interval.String()).Debug("gc sweep routine started")

	// Collect unreferenced volume nodes on each maintenance deadline, for at
	// most one interval at a time, then a due block pass, which sweeps all it
	// marked however long that takes. A collection that stops with candidates
	// left continues at once.
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		// Wait for the next collection deadline or volume cancellation.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}

		// Reap expired volume root pins before collecting orphaned nodes.
		if reaper, ok := vol.(interface{ ReapRootPins(context.Context) error }); ok {
			if err := reaper.ReapRootPins(ctx); err != nil {
				return err
			}
		}

		// Collect unreferenced nodes, retrying failures at the next deadline.
		stats, err := collector.CollectFor(ctx, interval)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.le.WithError(err).Warn("gc sweep failed")
			timer.Reset(interval)
			continue
		}

		// Report each collection that found candidates or marked, warn of a
		// pass that blocks of unknown references stopped, and continue a
		// collection the budget stopped.
		if stats.UnreferencedNodeCount != 0 || stats.BlockPasses != 0 {
			c.le.
				WithField("swept", stats.NodesSwept).
				WithField("candidates", stats.Candidates).
				WithField("live", stats.LiveNodes).
				WithField("unreached", stats.UnreachedBlocks).
				WithField("blocks-swept", stats.BlocksSwept).
				WithField("mark-duration", stats.MarkDuration.String()).
				WithField("more", stats.More).
				WithField("duration", stats.Duration.String()).
				Info("gc sweep completed")
		}
		if stats.UnknownBlocks != 0 {
			c.le.
				WithField("unknown", stats.UnknownBlocks).
				WithField("sample", stats.UnknownSample).
				Warn("gc block pass reached blocks of unknown references and swept no block")
		}
		// Let the storage reclaim the space of the swept blocks.
		if stats.NodesSwept != 0 || stats.BlocksSwept != 0 {
			if m, ok := vol.(storageMaintainer); ok {
				if err := m.Maintenance(ctx); err != nil && ctx.Err() == nil {
					c.le.WithError(err).Warn("storage maintenance failed")
				}
			}
		}
		if stats.More {
			timer.Reset(0)
		} else {
			timer.Reset(interval)
		}
	}
}
