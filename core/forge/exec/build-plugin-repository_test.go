//go:build !js && !tinygo

package space_exec

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/pkg/errors"
	manifest "github.com/s4wave/spacewave/bldr/manifest"
	builder "github.com/s4wave/spacewave/bldr/manifest/builder"
	"github.com/s4wave/spacewave/bldr/manifest/builder/resultworld"
	manifest_world "github.com/s4wave/spacewave/bldr/manifest/world"
	"github.com/s4wave/spacewave/bldr/plugin/compiler/esbuild"
	js "github.com/s4wave/spacewave/bldr/plugin/compiler/js"
	validate "github.com/s4wave/spacewave/bldr/project/validate"
	"github.com/s4wave/spacewave/db/block"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/s4wave/spacewave/db/unixfs"
	unixfs_block "github.com/s4wave/spacewave/db/unixfs/block"
	unixfs_block_fs "github.com/s4wave/spacewave/db/unixfs/block/fs"
	unixfs_iofs "github.com/s4wave/spacewave/db/unixfs/iofs"
	unixfs_world "github.com/s4wave/spacewave/db/unixfs/world"
	"github.com/s4wave/spacewave/db/world"
	forge_target "github.com/s4wave/spacewave/forge/target"
	forge_value "github.com/s4wave/spacewave/forge/value"
	"github.com/sirupsen/logrus"
)

// repositoryTarball creates an npm package whose install script must never run.
func repositoryTarball(t *testing.T) []byte {
	// Encode the package metadata and module under npm's archive prefix.
	t.Helper()
	var output bytes.Buffer
	compressed := gzip.NewWriter(&output)
	archive := tar.NewWriter(compressed)
	for _, file := range []struct {
		// path is the archive-relative package filename.
		path string
		// content is the package file's complete content.
		content string
	}{
		{"package/package.json", `{"name":"answer","version":"1.0.0","main":"index.js","scripts":{"install":"exit 99"}}`},
		{"package/index.js", `export const answer = "verified-block-only-value"`},
	} {
		if err := archive.WriteHeader(&tar.Header{Name: file.path, Mode: 0o644, Size: int64(len(file.content))}); err != nil {
			t.Fatal(err)
		}
		if _, err := archive.Write([]byte(file.content)); err != nil {
			t.Fatal(err)
		}
	}

	// Finish both encodings before computing the integrity digest.
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

// TestBuildPluginRepository exercises the production handler with a validated
// snapshot and verified npm bytes, with no executable build tools available.
func TestBuildPluginRepository(t *testing.T) {
	// Serve a pinned dependency that can be fetched only from this fixture registry.
	tarball := repositoryTarball(t)
	requests := make(chan string, 1)
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests <- request.URL.Path
		if _, err := w.Write(tarball); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(registry.Close)
	sum := sha512.Sum512(tarball)
	lock := `{"lockfileVersion":1,"packages":{"answer":["answer@1.0.0",` +
		strconv.Quote(registry.URL+"/answer.tgz") + `,{},` +
		strconv.Quote("sha512-"+base64.StdEncoding.EncodeToString(sum[:])) + `]}}`

	// Write the repository into the same World the real Forge execution uses.
	ctx := t.Context()
	handlers := NewDefaultRegistry()
	tb, sender := setupIntegrationTest(t, handlers)
	const sourceKey = "plugin-repos/acme/fixture/workdir"
	files := map[string]string{
		"bldr.star": `project(id="fixture")
manifest("acme-fixture", builder="bldr/plugin/compiler/js", description="Pinned repository", rev=4,
    config={"modules": [js_module("JS_MODULE_KIND_BACKEND", "./backend.ts", entrypoint=True)]})
`,
		"package.json": `{"name":"acme-fixture","dependencies":{"answer":"1.0.0"}}`,
		"bun.lock":     lock,
		"backend.ts":   `import {answer} from "answer"; export const value: string = answer; export default async function plugin() { console.info(value) }`,
	}
	createTestFS(t, ctx, tb.WorldState, sender, sourceKey, "bldr.star", []byte(files["bldr.star"]))
	for name, contents := range files {
		if name == "bldr.star" {
			continue
		}
		obj, err := world.MustGetObject(ctx, tb.WorldState, sourceKey)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = unixfs_world.FsMknodWithContent(ctx, tb.WorldState, obj, sender, unixfs_world.FSType_FSType_FS_NODE,
			[]string{name}, unixfs.NewFSCursorNodeType_File(), int64(len(contents)), strings.NewReader(contents), 0o644, time.Now())
		world.ReleaseObjectState(obj)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Validate and pin the exact source before constructing or running the handler.
	obj, err := world.MustGetObject(ctx, tb.WorldState, sourceKey)
	if err != nil {
		t.Fatal(err)
	}
	source, err := forge_value.NewWorldObjectSnapshot(ctx, obj, tb.WorldState)
	world.ReleaseObjectState(obj)
	if err != nil {
		t.Fatal(err)
	}
	err = tb.WorldState.AccessWorldState(ctx, source.GetRootRef(), func(cursor *bucket_lookup.Cursor) error {
		// Read the pinned tree through the same UnixFS adapter the handler uses.
		filesystem := unixfs_block_fs.NewFS(ctx, unixfs_block.NodeType_NodeType_DIRECTORY, cursor.Clone(), nil)
		handle, err := unixfs.NewFSHandle(filesystem)
		if err != nil {
			filesystem.Release()
			return err
		}
		defer handle.Release()
		validation, err := validate.Validate(ctx, unixfs_iofs.NewFS(ctx, handle))
		if err != nil {
			return err
		}
		if len(validation.GetRefusals()) != 0 {
			return errors.Errorf("fixture refused: %v", validation.GetRefusals())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Observe the real compiler boundary, including handler construction at startup.
	loads := 0
	load := func(ctx context.Context, le *logrus.Entry, b bus.Bus, source *unixfs.FSHandle, deps kvtx.TxOps, plugin *js.Config, bound int64) (builder.Controller, error) {
		loads++
		return esbuild.NewRepositoryCompiler(ctx, le, b, source, deps, plugin, bound)
	}
	config := []byte(`{"manifest_id":"acme-fixture","config_path":"bldr.star","platform_id":"js","milli_cpu":1000,"memory_bytes":1073741824}`)
	inputs := forge_target.InputMap{
		"source": forge_target.NewInputValueInline(forge_value.NewValueWithWorldObjectSnapshot("source", source)),
		"world":  forge_target.NewInputValueWorld(tb.EngineID, tb.Engine, tb.WorldState),
	}
	handler, err := NewBuildPluginHandlerWithRepositoryCompiler(ctx, tb.Bus, tb.Logger, tb.WorldState, &pluginExecHandleStub{}, inputs, config, load)
	if err != nil {
		t.Fatal(err)
	}
	if handler == nil || loads != 0 || len(requests) != 0 {
		t.Fatal("constructing the handler loaded a compiler or fetched a dependency")
	}
	handlers.Register(BuildPluginConfigID, func(ctx context.Context, le *logrus.Entry, ws world.WorldState, handle forge_target.ExecControllerHandle, inputs forge_target.InputMap, data []byte) (Handler, error) {
		return NewBuildPluginHandlerWithRepositoryCompiler(ctx, tb.Bus, le, ws, handle, inputs, data, load)
	})

	// Queue the snapshot without Bun and replace the editable repository source.
	t.Setenv("PATH", t.TempDir())
	createTestExecutionWithValueSet(t, ctx, tb.WorldState, sender, "repository-build", BuildPluginConfigID, config,
		&forge_target.ValueSet{Inputs: forge_value.ValueSlice{forge_value.NewValueWithWorldObjectSnapshot("source", source)}})
	_, _, err = unixfs_world.FsInit(ctx, tb.WorldState, sender, sourceKey, unixfs_world.FSType_FSType_FS_NODE, nil, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	// Run the production handler and check that compilation and fetching are lazy.
	execution := runTestExecution(t, tb, "repository-build", sender)
	assertComplete(t, execution)
	if loads != 1 || len(requests) != 1 {
		t.Fatalf("compiler loads = %d, registry requests = %d; want one each", loads, len(requests))
	}
	output := findOutput(execution, "manifest").GetWorldObjectSnapshot().GetRootRef()
	assertBuildProvenance(t, ctx, tb, output, source.GetRootRef())

	// Read esbuild's shared entrypoint and bundled dependency through the manifest API.
	err = manifest_world.AccessManifest(ctx, tb.Logger, tb.WorldState.AccessWorldState, output,
		func(ctx context.Context, _ *bucket_lookup.Cursor, _ *block.Cursor, built *manifest.Manifest, dist, assets *unixfs.FSHandle) error {
			// Preserve the metadata selected from the pinned Starlark configuration.
			if built.GetMeta().GetDescription() != "Pinned repository" || built.GetMeta().GetRev() != 4 {
				return errors.New("repository metadata was not preserved")
			}

			// Use Bldr's entrypoint wrapper and compile the module from verified blocks.
			entry, err := fs.ReadFile(unixfs_iofs.NewFS(ctx, dist), built.GetEntrypoint())
			if err != nil {
				return err
			}
			module, err := fs.ReadFile(unixfs_iofs.NewFS(ctx, assets), "modules/0.mjs")
			if err != nil {
				return err
			}
			if !strings.Contains(string(entry), "createPluginEntrypoint") || !strings.Contains(string(module), "verified-block-only-value") {
				return errors.New("handler did not publish the esbuild repository output")
			}
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}

	// Retain the reviewed repository files in the production build provenance.
	result, _, err := resultworld.LookupManifestBuildResult(ctx, tb.WorldState, manifest.NewManifestArtifactKey(output))
	if err != nil {
		t.Fatal(err)
	}
	paths := make([]string, 0, len(result.GetInputManifest().GetFiles()))
	for _, file := range result.GetInputManifest().GetFiles() {
		paths = append(paths, file.GetPath())
	}
	for _, want := range []string{"bldr.star", "package.json", "bun.lock", "backend.ts"} {
		if !slices.Contains(paths, want) {
			t.Fatalf("build provenance omitted %q: %v", want, paths)
		}
	}
}
