//go:build !js

package bldr_plugin_compiler_go

import (
	configset_proto "github.com/aperturerobotics/controllerbus/controller/configset/proto"
	"github.com/aperturerobotics/util/enabled"
	bldr_web_bundler "github.com/s4wave/spacewave/bldr/web/bundler"
)

// BuildPluginOpts is the configuration for one invocation of BuildPlugin.
type BuildPluginOpts struct {
	// JsMinification enables minification of bundled JavaScript output.
	JsMinification bool
	// JsSourcemaps enables source maps for bundled JavaScript output.
	JsSourcemaps bool
	// GoScriptCodeSplitting enables code splitting for GoScript builds.
	GoScriptCodeSplitting bool

	// OutBinName is the file name of the compiled plugin binary.
	OutBinName string
	// WorkingPath is the working directory used for the build.
	WorkingPath string
	// SourcePath is the path to the plugin source directory.
	SourcePath string
	// DistSourcePath is the source directory of the dist output to bundle.
	DistSourcePath string
	// OutDistPath is the output directory for the dist output.
	OutDistPath string
	// OutAssetsPath is the output directory for the assets output.
	OutAssetsPath string

	// GoPkgs is the list of go packages included in the plugin.
	GoPkgs []string
	// WebPkgs is the list of web package references included in the plugin.
	WebPkgs []*bldr_web_bundler.WebPkgRefConfig
	// WebPluginID is optional; if set, automatically adds controllers to configure the web plugin.
	WebPluginID string
	// DelveAddr optionally attaches a delve debugger listening on this address.
	DelveAddr string
	// LinkerFlags are extra linker flags merged into the build's -ldflags.
	LinkerFlags []string

	// DisableRpcFetch disables rpc fetch controllers in the plugin.
	DisableRpcFetch bool

	// ConfigSet is an additional config set embedded into the plugin.
	ConfigSet map[string]*configset_proto.ControllerConfig
	// HostConfigSet is an additional config set embedded into the plugin host.
	HostConfigSet map[string]*configset_proto.ControllerConfig

	// EnableCgoOpt enables cgo support on default false.
	EnableCgoOpt enabled.Enabled
	// GoCompilerOpt selects the go compiler used for the build.
	GoCompilerOpt GoCompiler
	// EnableImportedFactoryDiscoveryOpt enables imported factory discovery on default false.
	EnableImportedFactoryDiscoveryOpt enabled.Enabled
	// EnableCompressionOpt enables compression of the plugin binary on default depending on release mode.
	EnableCompressionOpt enabled.Enabled

	// BaseEsbuildFlags are extra flags passed to esbuild.
	BaseEsbuildFlags []string
	// DevInfoFile is the path of the dev info file to generate, if set.
	DevInfoFile string
}
