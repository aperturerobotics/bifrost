package prom_plugin

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	s4wave_prom "github.com/s4wave/spacewave/sdk/prom"
)

// maxRangeSteps bounds the steps of one range query, as the Prometheus HTTP
// API does.
const maxRangeSteps = 11_000

// newQueryEngine returns a PromQL engine with Prometheus defaults.
func newQueryEngine() *promql.Engine {
	return promql.NewEngine(promql.EngineOpts{
		MaxSamples:    50_000_000,
		Timeout:       2 * time.Minute,
		LookbackDelta: 5 * time.Minute,
	})
}

// queryService serves PromQL over every target in one World.
type queryService struct {
	// engine evaluates PromQL.
	engine *promql.Engine
	// q reads the World's targets.
	q storage.Queryable
}

// Query evaluates one PromQL expression at an instant or over a range.
func (s *queryService) Query(ctx context.Context, req *s4wave_prom.QueryRequest) (*s4wave_prom.QueryResponse, error) {
	// Build an instant or a range query.
	var query promql.Query
	var err error
	if step := time.Duration(req.GetStepMs()) * time.Millisecond; step > 0 {
		start, end := time.UnixMilli(req.GetStartMs()), time.UnixMilli(req.GetEndMs())
		if end.Before(start) {
			return nil, errors.New("range query ends before it starts")
		}
		if end.Sub(start)/step > maxRangeSteps {
			return nil, errors.Errorf("range query exceeds %d steps", maxRangeSteps)
		}
		query, err = s.engine.NewRangeQuery(ctx, s.q, nil, req.GetExpr(), start, end, step)
	} else {
		ts := time.Now()
		if req.GetTimeMs() != 0 {
			ts = time.UnixMilli(req.GetTimeMs())
		}
		query, err = s.engine.NewInstantQuery(ctx, s.q, nil, req.GetExpr(), ts)
	}
	if err != nil {
		return nil, err
	}
	defer query.Close()

	// Evaluate it and convert the value before Close returns its slices.
	res := query.Exec(ctx)
	if res.Err != nil {
		return nil, res.Err
	}
	resp := convertValue(res.Value)
	for _, w := range res.Warnings.AsErrors() {
		resp.Warnings = append(resp.Warnings, w.Error())
	}
	return resp, nil
}

// convertValue converts a PromQL value to its wire form. Histogram samples
// are left out; targets store float samples only.
func convertValue(v any) *s4wave_prom.QueryResponse {
	resp := &s4wave_prom.QueryResponse{}
	switch v := v.(type) {
	case promql.Vector:
		resp.ResultType = s4wave_prom.ResultType_RESULT_TYPE_VECTOR
		for _, smp := range v {
			resp.Series = append(resp.Series, &s4wave_prom.Series{
				Labels:  NewLabelSet(smp.Metric),
				Samples: []*s4wave_prom.Sample{{TimeMs: smp.T, Value: smp.F}},
			})
		}
	case promql.Matrix:
		resp.ResultType = s4wave_prom.ResultType_RESULT_TYPE_MATRIX
		for _, series := range v {
			out := &s4wave_prom.Series{Labels: NewLabelSet(series.Metric)}
			for _, p := range series.Floats {
				out.Samples = append(out.Samples, &s4wave_prom.Sample{TimeMs: p.T, Value: p.F})
			}
			resp.Series = append(resp.Series, out)
		}
	case promql.Scalar:
		resp.ResultType = s4wave_prom.ResultType_RESULT_TYPE_SCALAR
		resp.Series = []*s4wave_prom.Series{{Samples: []*s4wave_prom.Sample{{TimeMs: v.T, Value: v.V}}}}
	case promql.String:
		resp.ResultType = s4wave_prom.ResultType_RESULT_TYPE_STRING
		resp.StringValue = v.V
	}
	return resp
}

// _ is a type assertion
var _ s4wave_prom.SRPCPromQueryResourceServiceServer = (*queryService)(nil)
