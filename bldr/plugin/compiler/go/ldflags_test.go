//go:build !js && !windows

package bldr_plugin_compiler_go

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	bldr_manifest "github.com/s4wave/spacewave/bldr/manifest"
	bldr_platform "github.com/s4wave/spacewave/bldr/platform"
	"github.com/s4wave/spacewave/bldr/util/gocompiler"
	"github.com/sirupsen/logrus"
)

const testVersionFlag = "-X github.com/example/app/sdk.Version=1.2.3"

// ldflagsValues returns the value of every -ldflags argument in args.
func ldflagsValues(args []string) []string {
	var values []string
	for i, arg := range args {
		if arg == "-ldflags" && i+1 < len(args) {
			values = append(values, args[i+1])
		}
	}
	return values
}

// TestCompilePluginLinkerFlags checks that a native release build passes the
// configured linker flags and -w -s in one -ldflags argument.
func TestCompilePluginLinkerFlags(t *testing.T) {
	// Put a fake go command first on PATH that records its arguments.
	binDir := t.TempDir()
	argsPath := filepath.Join(binDir, "args.txt")
	script := "#!/bin/sh\nfor a in \"$@\"; do printf '%s\\n' \"$a\"; done > " + argsPath + "\n"
	if err := os.WriteFile(filepath.Join(binDir, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv(gocompiler.GoWasmDiagnosticEnv, "")

	// Build for a non-darwin native platform so no code signing runs.
	platform, err := bldr_platform.ParsePlatform("desktop/linux/amd64")
	if err != nil {
		t.Fatal(err)
	}

	// Compile a release plugin with an extra linker flag.
	codegenDir := t.TempDir()
	le := logrus.NewEntry(logrus.New())
	mc, err := NewModuleCompiler(le, codegenDir, "example")
	if err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(codegenDir, "plugin")
	err = mc.CompilePlugin(
		context.Background(),
		le,
		outFile,
		platform,
		bldr_manifest.BuildType_RELEASE,
		false,
		false,
		[]string{testVersionFlag},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Read back the arguments the fake go command received.
	data, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	values := ldflagsValues(args)
	if len(values) != 1 {
		t.Fatalf("want one -ldflags argument, got %v in %v", values, args)
	}
	if want := testVersionFlag + " -w -s"; values[0] != want {
		t.Fatalf("ldflags = %q, want %q", values[0], want)
	}
}

// TestDevWrapperBuildArgsLinkerFlags checks that the dev wrapper build merges
// the Delve address and the configured linker flags into one -ldflags argument.
func TestDevWrapperBuildArgsLinkerFlags(t *testing.T) {
	// Build with both a Delve address and a linker flag.
	args, err := devWrapperBuildArgs("wrapper", ":5000", []string{testVersionFlag})
	if err != nil {
		t.Fatal(err)
	}
	values := ldflagsValues(args)
	if len(values) != 1 {
		t.Fatalf("want one -ldflags argument, got %v in %v", values, args)
	}
	if want := testVersionFlag + " -X 'main.DelveAddr=:5000'"; values[0] != want {
		t.Fatalf("ldflags = %q, want %q", values[0], want)
	}
	if args[len(args)-1] != "." {
		t.Fatalf("last argument = %q, want .", args[len(args)-1])
	}

	// No linker flags and no Delve address leaves -ldflags out.
	args, err = devWrapperBuildArgs("wrapper", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(args, "-ldflags") {
		t.Fatalf("unexpected -ldflags in %v", args)
	}

	// An invalid Delve address is rejected.
	if _, err := devWrapperBuildArgs("wrapper", "bad addr\"", nil); err == nil {
		t.Fatal("expected invalid delve address error")
	}
}

// TestDevPluginBuildFlagsLinkerFlags checks the flags the wrapper uses to
// build the plugin at run time.
func TestDevPluginBuildFlagsLinkerFlags(t *testing.T) {
	// Build the runtime flags with a linker flag.
	flags := devPluginBuildFlags(bldr_manifest.BuildType_DEV, []string{testVersionFlag})
	values := ldflagsValues(flags)
	if len(values) != 1 || values[0] != testVersionFlag {
		t.Fatalf("ldflags = %v, want [%q] in %v", values, testVersionFlag, flags)
	}
	if !slices.Contains(flags, "-tags=build_type_dev") {
		t.Fatalf("missing build type tag in %v", flags)
	}
	if slices.Contains(devPluginBuildFlags(bldr_manifest.BuildType_DEV, nil), "-ldflags") {
		t.Fatal("unexpected -ldflags without linker flags")
	}
}

// TestConfigMergeLinkerFlags checks build type overrides append linker flags.
func TestConfigMergeLinkerFlags(t *testing.T) {
	// Merge an override into a base config.
	c := &Config{LinkerFlags: []string{"-s"}}
	c.Merge(&Config{LinkerFlags: []string{testVersionFlag}})
	if got := c.GetLinkerFlags(); !slices.Equal(got, []string{"-s", testVersionFlag}) {
		t.Fatalf("linker flags = %v", got)
	}
}
