package bldr_dist_compiler

import (
	"testing"

	transform_gzip "github.com/s4wave/spacewave/db/block/transform/gzip"
	transform_s2 "github.com/s4wave/spacewave/db/block/transform/s2"
)

// TestBuildEmbedTransformConf checks the decoder selected for each target.
func TestBuildEmbedTransformConf(t *testing.T) {
	for _, test := range []struct {
		name            string
		goScriptBrowser bool
		configID        string
	}{
		{name: "native", configID: transform_s2.ConfigID},
		{name: "goscript-browser", goScriptBrowser: true, configID: transform_gzip.ConfigID},
	} {
		t.Run(test.name, func(t *testing.T) {
			steps := buildEmbedTransformConf(test.goScriptBrowser)
			if len(steps) != 1 {
				t.Fatalf("expected one compression step, got %d", len(steps))
			}
			if got := steps[0].GetConfigID(); got != test.configID {
				t.Fatalf("compression step = %s, want %s", got, test.configID)
			}
			if err := steps[0].Validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
