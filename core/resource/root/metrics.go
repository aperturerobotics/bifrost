package resource_root

import (
	"context"

	"github.com/s4wave/spacewave/core/metrics"
	s4wave_root "github.com/s4wave/spacewave/sdk/root"
)

// GatherMetrics returns the metric families of the core process.
func (s *CoreRootServer) GatherMetrics(
	ctx context.Context,
	req *s4wave_root.GatherMetricsRequest,
) (*s4wave_root.GatherMetricsResponse, error) {
	families, err := metrics.Gather()
	if err != nil {
		return nil, err
	}
	return &s4wave_root.GatherMetricsResponse{MetricFamilies: families}, nil
}
