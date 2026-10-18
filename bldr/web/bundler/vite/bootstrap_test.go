//go:build !js

package bldr_web_bundler_vite

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aperturerobotics/fastjson"
	"github.com/sirupsen/logrus"
)

const esLitePkg = "@aptre/protobuf-es-lite"

// pinnedVersion returns the exact version a package.json pins for esLitePkg.
func pinnedVersion(t *testing.T, path string) string {
	// Read and parse the manifest.
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var parser fastjson.Parser
	manifest, err := parser.ParseBytes(data)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	// Return the pinned version.
	return string(manifest.GetStringBytes("dependencies", esLitePkg))
}

// TestServiceScriptResolvesPinnedProtobufEsLite checks that the Vite service
// script loads the protobuf-es-lite version the repository pins, even when its
// directory is outside any project. Bun auto-installs a package it cannot
// resolve from its global cache, which can hold a runtime older than the
// generated messages the script loads.
func TestServiceScriptResolvesPinnedProtobufEsLite(t *testing.T) {
	// Locate the Bldr source root and the versions the repository pins.
	distRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	want := pinnedVersion(t, filepath.Join(distRoot, "dist", "deps", "package.json"))
	if root := pinnedVersion(t, filepath.Join(distRoot, "..", "package.json")); root != want {
		t.Fatalf("%s is pinned to %q in package.json but %q in bldr/dist/deps/package.json; raise both together", esLitePkg, root, want)
	}

	// Build the service script into a temporary directory with no node_modules above it.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	outDir := t.TempDir()
	scriptPath := filepath.Join(outDir, "bldr-vite-test.mjs")
	le := logrus.NewEntry(logrus.New())
	if _, err := BuildServiceScript(ctx, le, t.TempDir(), distRoot, distRoot, scriptPath); err != nil {
		t.Fatal(err)
	}

	// Resolve the package from the script's directory the way bun does.
	probePath := filepath.Join(outDir, "probe.mjs")
	probe := `import { existsSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
let dir = dirname(fileURLToPath(import.meta.resolve("` + esLitePkg + `")));
while (!existsSync(join(dir, "package.json")) || JSON.parse(readFileSync(join(dir, "package.json"), "utf8")).name !== "` + esLitePkg + `") {
  dir = dirname(dir);
}
console.log(JSON.parse(readFileSync(join(dir, "package.json"), "utf8")).version);
`
	if err := os.WriteFile(probePath, []byte(probe), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run the probe with bun from the script's directory.
	cmd := exec.CommandContext(ctx, "bun", probePath)
	cmd.Dir = outDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("resolve %s: %v\n%s", esLitePkg, err, stderr.String())
	}

	// Require the resolved version to match the pin.
	if got := strings.TrimSpace(stdout.String()); got != want {
		t.Fatalf("service script resolves %s %s, want pinned %s", esLitePkg, got, want)
	}
}
