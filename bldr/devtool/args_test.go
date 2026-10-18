//go:build !js

package devtool

import (
	"testing"

	"github.com/aperturerobotics/cli"
)

// TestBuildManifestRevisionFlag preserves the full manifest revision width.
func TestBuildManifestRevisionFlag(t *testing.T) {
	// Parse the build command without starting a project.
	args := &DevtoolArgs{}
	command := args.BuildBuildCommand()
	var revision uint64
	command.Action = func(c *cli.Context) error {
		revision = c.Uint64("manifest-revision")
		return nil
	}
	app := cli.NewApp()
	app.Commands = []*cli.Command{command}
	if err := app.Run([]string{"bldr", "build", "--manifest-revision", "18446744073709551615"}); err != nil {
		t.Fatal(err)
	}

	// Preserve the full width of the manifest revision representation.
	if revision != ^uint64(0) {
		t.Fatalf("manifest revision = %d, want %d", revision, ^uint64(0))
	}
	if args.ManifestRevision != revision {
		t.Fatalf("build revision = %d, want %d", args.ManifestRevision, revision)
	}
}

func TestResolveWebStartModeDefaultsToGoScript(t *testing.T) {
	args := &DevtoolArgs{}

	mode, err := args.resolveWebStartMode()
	if err != nil {
		t.Fatal(err)
	}
	if mode != webStartModeGoScript {
		t.Fatalf("web start mode = %s, want %s", mode, webStartModeGoScript)
	}
}

func TestResolveWebStartModeWasmOverride(t *testing.T) {
	args := &DevtoolArgs{WebUseWasm: true}

	mode, err := args.resolveWebStartMode()
	if err != nil {
		t.Fatal(err)
	}
	if mode != webStartModeWasm {
		t.Fatalf("web start mode = %s, want %s", mode, webStartModeWasm)
	}
}

func TestResolveWebStartModeRejectsConflictingCompilerFlags(t *testing.T) {
	args := &DevtoolArgs{WebUseWasm: true, WebUseGoScript: true}

	if _, err := args.resolveWebStartMode(); err == nil {
		t.Fatal("resolveWebStartMode() error = nil, want conflict error")
	}
}
