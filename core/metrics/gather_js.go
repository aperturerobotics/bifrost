//go:build js

package metrics

import (
	io_prometheus_client "github.com/aperturerobotics/go-prometheus-client-lite/client_model/go"
)

// gather reports no families: the browser builds leave the registry out.
func gather() ([]*io_prometheus_client.MetricFamily, error) {
	return nil, nil
}
