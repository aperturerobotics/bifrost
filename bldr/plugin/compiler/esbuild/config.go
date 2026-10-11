// Package esbuild builds JavaScript plugins from a source filesystem and
// integrity-pinned npm tarballs, without a package manager or disk checkout.
package esbuild

import (
	"io/fs"
	"math"
	"path"
	"strings"

	"github.com/pkg/errors"
	bundler "github.com/s4wave/spacewave/bldr/web/bundler"
	bundle "github.com/s4wave/spacewave/bldr/web/bundler/esbuild"
)

// Config selects a plugin entrypoint and optional separately served assets.
// The plugin entrypoint exports the host's BackendEntrypoint function. Assets
// use Bldr's existing bundle entrypoint records and frontend exclusion policy.
type Config struct {
	// Entrypoint is the source path of the plugin's default-exported function.
	Entrypoint string
	// Assets are additional source entrypoints emitted into the manifest assets.
	Assets []*bundle.EsbuildBundleEntrypoint
	// WebPkgs marks packages supplied by other plugins with exclude=true.
	WebPkgs []*bundler.WebPkgRefConfig
	// MaxExpandedSize bounds uncompressed bytes across npm packages in one build.
	MaxExpandedSize int64
}

// Validate checks source paths and the bound before compiling untrusted inputs.
func (c *Config) Validate() error {
	// Require a plugin entrypoint and a finite archive expansion allowance.
	if !fs.ValidPath(c.Entrypoint) || c.Entrypoint == "." {
		return errors.New("esbuild plugin requires a relative entrypoint")
	}
	if c.MaxExpandedSize <= 0 || c.MaxExpandedSize == math.MaxInt64 {
		return errors.New("esbuild plugin requires a positive finite archive expansion bound")
	}

	// Keep every asset output inside the manifest's assets filesystem.
	for _, entry := range c.Assets {
		if !fs.ValidPath(entry.GetInputPath()) || entry.GetInputPath() == "." {
			return errors.New("esbuild asset requires a relative input path")
		}
		output := path.Clean(entry.GetOutputPath())
		if !fs.ValidPath(output) || output == "." {
			return errors.New("esbuild asset requires a relative output path")
		}
	}

	// Spacewave code travels with the plugin while Bldr supplies its singletons.
	for id := range bundler.ExcludedWebPkgIDs(c.WebPkgs) {
		if id == "spacewave" || strings.HasPrefix(id, "spacewave/") {
			return errors.New("spacewave must remain bundled in the plugin")
		}
	}
	return nil
}
