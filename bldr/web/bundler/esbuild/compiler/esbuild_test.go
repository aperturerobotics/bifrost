//go:build !js

package bldr_web_bundler_esbuild_compiler

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	esbuild_api "github.com/aperturerobotics/esbuild/pkg/api"
	bldr_web_bundler "github.com/s4wave/spacewave/bldr/web/bundler"
	bldr_web_bundler_esbuild "github.com/s4wave/spacewave/bldr/web/bundler/esbuild"
	"github.com/sirupsen/logrus"
)

// TestBuildEsbuildBundlePreservesConfiguredPluginContract builds one
// entrypoint and checks its inputs, output metadata and tree shaking.
func TestBuildEsbuildBundlePreservesConfiguredPluginContract(t *testing.T) {
	// Write the entrypoint and sentinel values into a temporary code root.
	root := t.TempDir()
	assets := filepath.Join(root, "assets")
	writeEsbuildControlFile(t, filepath.Join(root, "entry.ts"), `
import { used } from './values'
globalThis.__bldrEsbuildControl = used
`)
	writeEsbuildControlFile(t, filepath.Join(root, "values.ts"), `
export const used = 'retained-sentinel'
export const unused = 'unused-sentinel'
`)

	// Build the bundle with the configured entrypoint and extension override.
	webPkgs, outputs, inputs, err := BuildEsbuildBundle(
		logrus.NewEntry(logrus.New()),
		root,
		"control",
		[]*bldr_web_bundler_esbuild.EsbuildBundleEntrypoint{{
			InputPath:    "entry.ts",
			OutputPath:   "control",
			EntrypointId: "control-entrypoint",
		}},
		&esbuild_api.BuildOptions{OutExtension: map[string]string{".js": ".mjs"}},
		nil,
		assets,
		"/p/control/esb/",
		false,
		true,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the source inputs were reported and no web packages were added.
	if len(webPkgs) != 0 {
		t.Fatalf("web package refs = %v, want none", webPkgs)
	}
	for _, input := range []string{"entry.ts", "values.ts"} {
		if !slices.Contains(inputs, input) {
			t.Fatalf("inputs = %v, missing %q", inputs, input)
		}
	}

	// Validate the output metadata for the configured entrypoint.
	outputIndex := slices.IndexFunc(outputs, func(output *bldr_web_bundler_esbuild.EsbuildOutputMeta) bool {
		return output.GetEntrypointId() == "control-entrypoint"
	})
	if outputIndex == -1 {
		t.Fatalf("outputs = %v, missing configured entrypoint metadata", outputs)
	}
	output := outputs[outputIndex]
	if output.GetEntrypointPath() != "entry.ts" {
		t.Fatalf("entrypoint path = %q, want entry.ts", output.GetEntrypointPath())
	}
	if filepath.IsAbs(output.GetPath()) || strings.HasPrefix(filepath.ToSlash(output.GetPath()), "../") {
		t.Fatalf("output path = %q, want asset-root-relative path", output.GetPath())
	}

	// Read the bundled output and verify the retained and unused exports.
	outputPath := filepath.Join(assets, output.GetPath())
	code, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(code, []byte("retained-sentinel")) {
		t.Fatalf("output missing retained export: %s", code)
	}
	if bytes.Contains(code, []byte("unused-sentinel")) {
		t.Fatalf("output retained unused export: %s", code)
	}
	if output.GetLength() != uint32(len(code)) {
		t.Fatalf("metadata length = %d, output length = %d", output.GetLength(), len(code))
	}

	// Report the compressed output size for the build log.
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	if _, err := gzipWriter.Write(code); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	t.Logf(
		"configured esbuild control: output=%s raw=%d gzip=%d inputs=%d outputs=%d",
		output.GetPath(),
		len(code),
		compressed.Len(),
		len(inputs),
		len(outputs),
	)
}

// TestBuildEsbuildBundleImportsSharedWebPkgs imports a shared web package
// through the bare specifier the document's import map binds, so the bundle
// shares the module instance the host loaded instead of fetching a second copy.
func TestBuildEsbuildBundleImportsSharedWebPkgs(t *testing.T) {
	// Write an entrypoint that imports a package configured as shared.
	root := t.TempDir()
	pkgDir := filepath.Join(root, "node_modules", "shared")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeEsbuildControlFile(t, filepath.Join(pkgDir, "package.json"), `{"name": "shared", "main": "index.js"}`)
	writeEsbuildControlFile(t, filepath.Join(pkgDir, "index.js"), `export const value = 'shared-sentinel'`)
	writeEsbuildControlFile(t, filepath.Join(root, "entry.ts"), `
import { value } from 'shared'
globalThis.__bldrEsbuildShared = value
`)

	// Build the bundle with the package externalized.
	assets := filepath.Join(root, "assets")
	webPkgs, outputs, _, err := BuildEsbuildBundle(
		logrus.NewEntry(logrus.New()),
		root,
		"shared",
		[]*bldr_web_bundler_esbuild.EsbuildBundleEntrypoint{{InputPath: "entry.ts", OutputPath: "shared"}},
		&esbuild_api.BuildOptions{},
		[]*bldr_web_bundler.WebPkgRefConfig{{Id: "shared"}},
		assets,
		"",
		false,
		false,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(webPkgs) != 1 || webPkgs[0].GetWebPkgId() != "shared" {
		t.Fatalf("web package refs = %v, want shared", webPkgs)
	}

	// The output imports the bare specifier and does not inline the package.
	for _, output := range outputs {
		code, err := os.ReadFile(filepath.Join(assets, output.GetPath()))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(code, []byte(`"bldr-web-pkg/shared/index.mjs"`)) {
			t.Fatalf("output does not import the shared specifier: %s", code)
		}
		if bytes.Contains(code, []byte("shared-sentinel")) {
			t.Fatalf("output inlined the shared package: %s", code)
		}
	}
}

// writeEsbuildControlFile writes contents to path or fails the test.
func writeEsbuildControlFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
