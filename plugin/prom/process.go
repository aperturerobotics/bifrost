package prom_plugin

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/world"
	s4wave_process "github.com/s4wave/spacewave/sdk/process"
	s4wave_prom "github.com/s4wave/spacewave/sdk/prom"
	"github.com/sirupsen/logrus"
)

// targetProcess is the scrape process of one target on the Device whose
// ProcessBinding approves it. It is the only writer of the target.
type targetProcess struct {
	// le is the logger.
	le *logrus.Entry
	// eng is the Space's World engine.
	eng world.Engine
	// group commits the Device's targets in the Space together.
	group *commitGroup
	// objectKey is the target's object key.
	objectKey string
	// peerID is the Device peer reported in the target's status.
	peerID string
}

// Execute scrapes the target until the Space stops the process. A missing
// target, or one without a scrape config, stops at once.
func (p *targetProcess) Execute(_ *s4wave_process.ExecuteRequest, stream s4wave_process.SRPCPersistentExecutionService_ExecuteStream) error {
	// Report each state change to the Space.
	send := func(state s4wave_process.ExecutionState, err error) error {
		status := &s4wave_process.ExecuteStatus{State: state}
		if err != nil {
			status.Error = err.Error()
		}
		return stream.Send(status)
	}
	ctx := stream.Context()
	if err := send(s4wave_process.ExecutionState_ExecutionState_STARTING, nil); err != nil {
		return err
	}

	// Read the scrape config; stop when there is nothing to scrape.
	cfg, err := p.readScrapeConfig(ctx)
	if err != nil {
		return send(s4wave_process.ExecutionState_ExecutionState_ERROR, err)
	}
	if cfg == nil {
		return send(s4wave_process.ExecutionState_ExecutionState_STOPPED, nil)
	}

	// Resume the target's writer after its committed history.
	store, err := OpenStore(ctx, p.eng, p.objectKey)
	if err != nil {
		return send(s4wave_process.ExecutionState_ExecutionState_ERROR, err)
	}

	// Scrape into the Device's commit group until the process stops.
	// Join the Device's commit group and report running.
	leave := p.group.join(store, p.eng)
	defer leave()
	if err := send(s4wave_process.ExecutionState_ExecutionState_RUNNING, nil); err != nil {
		return err
	}

	// Scrape until the process stops; a stop is not an error.
	p.le.WithField("url", cfg.GetUrl()).WithField("interval", cfg.Interval().String()).Info("scraping target")
	err = newScraper(p.le, store, cfg, p.peerID).run(ctx)
	if ctx.Err() != nil {
		return nil
	}
	return send(s4wave_process.ExecutionState_ExecutionState_ERROR, err)
}

// readScrapeConfig returns the target's scrape config, or nil when the target
// is missing or scrapes nothing.
func (p *targetProcess) readScrapeConfig(ctx context.Context) (*s4wave_prom.ScrapeConfig, error) {
	// Read the target's meta, if the object still exists.
	tx, err := p.eng.NewTransaction(ctx, false)
	if err != nil {
		return nil, err
	}
	defer tx.Discard()
	found, err := tx.HasObject(ctx, p.objectKey)
	if err != nil || !found {
		return nil, err
	}
	meta, _, err := s4wave_prom.ReadTarget(ctx, tx, p.objectKey)
	if err != nil {
		return nil, err
	}

	// Check its scrape config.
	cfg := meta.GetScrape()
	if cfg == nil {
		return nil, nil
	}
	return cfg, errors.Wrap(cfg.Validate(), "target scrape config")
}

// _ is a type assertion
var _ s4wave_process.SRPCPersistentExecutionServiceServer = (*targetProcess)(nil)
