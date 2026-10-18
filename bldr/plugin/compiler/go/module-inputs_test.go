package bldr_plugin_compiler_go

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	builder "github.com/s4wave/spacewave/bldr/manifest/builder"
)

// TestGoModuleInputsTracked keeps dependency selection in the same input
// manifest that controls reuse of compiled plugin binaries.
func TestGoModuleInputsTracked(t *testing.T) {
	// Materialize the module inputs discovered by the native compiler.
	root := t.TempDir()
	want := []string{"go.mod", "go.sum", "vendor/modules.txt"}
	for _, name := range want {
		file := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("dependency selection\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Use the compiler's discovery and filtering path, not hand-built absolute
	// test inputs that bypass the source-relative discovery boundary.
	inputs := &builder.InputManifest{}
	if err := appendInputManifestFiles(inputs, root, InputFileKind_InputFileKind_GO, existingSourceFiles(root, want...), false); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, file := range inputs.GetFiles() {
		got = append(got, filepath.ToSlash(file.GetPath()))
	}
	if !slices.Equal(got, want) {
		t.Fatalf("module dependency inputs = %v, want %v", got, want)
	}
}

// TestAppendInputManifestFilesTracksReplacementSources keeps mutable module sources in startup validation.
func TestAppendInputManifestFilesTracksReplacementSources(t *testing.T) {
	// Create one source root and one external replacement file.
	sourceRoot := t.TempDir()
	replacementRoot := t.TempDir()
	sourcePath := filepath.Join(replacementRoot, "dep.go")
	if err := os.WriteFile(sourcePath, []byte("package dep\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	overridePath := filepath.Join(replacementRoot, "gs", "override.go")
	if err := os.MkdirAll(filepath.Dir(overridePath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(overridePath, []byte("package override\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Record replacement sources and overrides with the compiler's file input producer.
	inputs := &builder.InputManifest{}
	if err := appendInputManifestFiles(inputs, sourceRoot, InputFileKind_InputFileKind_GO, []string{sourcePath}, true); err != nil {
		t.Fatal(err)
	}
	if err := appendInputManifestFiles(inputs, sourceRoot, InputFileKind_InputFileKind_GOSCRIPT_OVERRIDE, []string{overridePath}, true); err != nil {
		t.Fatal(err)
	}

	// Preserve the external source path relative to the project root.
	wantSourcePath, err := filepath.Rel(sourceRoot, sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	wantOverridePath, err := filepath.Rel(sourceRoot, overridePath)
	if err != nil {
		t.Fatal(err)
	}
	gotPaths := []string{inputs.GetFiles()[0].GetPath(), inputs.GetFiles()[1].GetPath()}
	wantPaths := []string{filepath.ToSlash(wantSourcePath), filepath.ToSlash(wantOverridePath)}
	if !slices.Equal(gotPaths, wantPaths) {
		t.Fatalf("replacement input paths = %v, want %v", gotPaths, wantPaths)
	}
}
