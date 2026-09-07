package bldr_dist_compiler

import (
	"github.com/aperturerobotics/controllerbus/config"
	transform_gzip "github.com/s4wave/spacewave/db/block/transform/gzip"
	transform_s2 "github.com/s4wave/spacewave/db/block/transform/s2"
)

// buildEmbedTransformConf selects compression for newly embedded manifest blocks.
// GoScript browsers decode gzip through their native decompression streams.
func buildEmbedTransformConf(goScriptBrowser bool) []config.Config {
	if goScriptBrowser {
		return []config.Config{&transform_gzip.Config{}}
	}
	return []config.Config{&transform_s2.Config{}}
}
