//go:build linux && !js

package bldr_project_controller

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	frontend "github.com/s4wave/spacewave/bldr/frontend"
	bldr_project "github.com/s4wave/spacewave/bldr/project"
	"github.com/sirupsen/logrus"
)

// TestFrontendLifecycle exercises the project service against its managed compiler.
func TestFrontendLifecycle(t *testing.T) {
	// Opt in to the real compiler and its idle-lifetime check.
	if os.Getenv("BLDR_FRONTEND_LIFECYCLE_TEST") != "true" {
		t.Skip("set BLDR_FRONTEND_LIFECYCLE_TEST=true to run the managed compiler lifecycle")
	}

	// Keep the fixture beneath the checkout so Bun finds its installed packages.
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err := os.MkdirTemp(repo, ".frontend-lifecycle-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	})

	// Point the compiler's distribution directory at the current checkout.
	work := filepath.Join(root, ".bldr")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(repo, filepath.Join(work, "src")); err != nil {
		t.Fatal(err)
	}
	writeFrontendFixture(t, root, "Viewer.ts", `import "./style.css"; import.meta.hot.accept(); export const value = "before"`)
	writeFrontendFixture(t, root, "style.css", "body { color: red }")
	writeFrontendFixture(t, root, "vite.config.ts", `import { writeFileSync } from "node:fs"
export default { plugins: [{ name: "process-fixture", configureServer() {
  writeFileSync(new URL("./compiler.pid", import.meta.url), String(process.pid))
} }] }`)

	// Configure the same service and compiler routine used by ordinary web startup.
	project := &bldr_project.ProjectConfig{}
	const projectJSON = `{"id":"fixture","manifests":{"fixture":{"builder":{"id":"bldr/plugin/compiler/js","config":{"modules":[{"kind":"JS_MODULE_KIND_FRONTEND","path":"Viewer.ts"}]}}}}}`
	if err := bldr_project.UnmarshalProjectConfig([]byte(projectJSON), project); err != nil {
		t.Fatal(err)
	}
	conf := &Config{SourcePath: root, WorkingPath: work, ProjectConfig: project}

	// Bound the compiler lifetime and retain it through the service's routine.
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Second)
	defer cancel()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	service := newFrontendService(logrus.NewEntry(logger), nil)
	defer service.Close()
	service.run.SetContext(ctx, true)
	if err := service.configure(conf); err != nil {
		t.Fatal(err)
	}

	// Subscribe before reading the compiler's initial session.
	watchCtx, stopWatch := context.WithCancel(ctx)
	stream := &frontendWatchStream{ctx: watchCtx, events: make(chan *frontend.Event, 64)}
	done := make(chan error, 1)
	go func() { done <- service.Watch(&frontend.WatchRequest{}, stream) }()
	defer stopWatch()
	first := nextFrontendEvent(t, ctx, stream, done)
	if first.GetSession().GetId() == "" {
		t.Fatal("missing initial compiler session")
	}

	// Fetch the initial modules from the admitted compiler capability.
	pid := frontendFixturePID(t, root)
	env, err := service.ready.Await(ctx)
	if err != nil {
		t.Fatal(err)
	}
	address := env.result.GetPrivateUrl() + first.GetSession().GetRoutePrefix()
	readFrontendFixture(t, ctx, address+"Viewer.ts", http.StatusOK, `"before"`)
	readFrontendFixture(t, ctx, address+"style.css", http.StatusOK, "color: red")
	t.Log("compiler ready")

	// Hold the live compiler past the snapshot compiler's release delay, then save.
	idle := time.NewTimer(31 * time.Second)
	defer idle.Stop()
	select {
	case <-idle.C:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	writeFrontendFixture(t, root, "style.css", "body { color: blue }")
	awaitFrontendPayload(t, ctx, stream, done, `"type":"update"`)
	readFrontendFixture(t, ctx, address+"style.css?t=1", http.StatusOK, "color: blue")
	if got := frontendFixturePID(t, root); got != pid {
		t.Fatalf("idle edit replaced compiler %d with %d", pid, got)
	}
	t.Log("idle edit applied")

	// Report invalid source over the update stream and recover on the next save.
	writeFrontendFixture(t, root, "Viewer.ts", "export const value = (")
	awaitFrontendPayload(t, ctx, stream, done, `"type":"update"`)
	readFrontendFixture(t, ctx, address+"Viewer.ts?t=invalid", http.StatusInternalServerError, "Frontend compilation failed")
	awaitFrontendPayload(t, ctx, stream, done, `"type":"error"`)
	writeFrontendFixture(t, root, "Viewer.ts", `import "./style.css"; import.meta.hot.accept(); export const value = "recovered"`)
	awaitFrontendPayload(t, ctx, stream, done, `"type":"update"`, `"type":"full-reload"`)
	readFrontendFixture(t, ctx, address+"Viewer.ts?t=2", http.StatusOK, `"recovered"`)
	t.Log("syntax error reported and next save recovered")

	// Rapid saves must converge to the last completed source write.
	for _, value := range []string{"one", "two", "last"} {
		writeFrontendFixture(t, root, "Viewer.ts", `import "./style.css"; import.meta.hot.accept(); export const value = "`+value+`"`)
	}
	awaitFrontendPayload(t, ctx, stream, done, `"type":"update"`)
	readFrontendFixture(t, ctx, address+"Viewer.ts?t=3", http.StatusOK, `"last"`)
	t.Log("rapid saves served the last source")

	// Cancel one update connection and attach again to the same live compiler.
	stopWatch()
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	// Resume the same compiler session after the interrupted attachment ends.
	stream = &frontendWatchStream{ctx: ctx, events: make(chan *frontend.Event, 64)}
	done = make(chan error, 1)
	go func() { done <- service.Watch(&frontend.WatchRequest{}, stream) }()
	reconnected := nextFrontendEvent(t, ctx, stream, done)
	if reconnected.GetSession().GetId() != first.GetSession().GetId() {
		t.Fatal("connection recovery replaced the compiler session")
	}

	// The resumed subscription must receive the next filesystem edit.
	writeFrontendFixture(t, root, "style.css", "body { color: green }")
	awaitFrontendPayload(t, ctx, stream, done, `"type":"update"`)
	readFrontendFixture(t, ctx, address+"style.css?t=4", http.StatusOK, "color: green")
	t.Log("dropped update stream reattached")

	// Replacing a Vite configuration exits the compiler and rebuilds its graph.
	writeFrontendFixture(t, root, "vite.config.ts", `import { writeFileSync } from "node:fs"
export default { define: { __CONFIG_VALUE__: '"replacement"' }, plugins: [{
  name: "process-fixture", configureServer() {
    writeFileSync(new URL("./compiler.pid", import.meta.url), String(process.pid))
  }
}] }`)
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	// Attach to the replacement graph after readiness changes.
	awaitFrontendReplacement(t, ctx, service, first.GetSession().GetId())
	stream = &frontendWatchStream{ctx: ctx, events: make(chan *frontend.Event, 64)}
	done = make(chan error, 1)
	go func() { done <- service.Watch(&frontend.WatchRequest{}, stream) }()
	replaced := nextFrontendEvent(t, ctx, stream, done)
	if replaced.GetSession().GetId() == first.GetSession().GetId() {
		t.Fatal("configuration replacement retained the old session")
	}

	// Verify the new configuration and the exit of the previous process.
	replacementPID := frontendFixturePID(t, root)
	if replacementPID == pid || syscall.Kill(pid, 0) != syscall.ESRCH {
		t.Fatal("configuration replacement left the old compiler running")
	}
	writeFrontendFixture(t, root, "Viewer.ts", `import.meta.hot.accept(); export const value = __CONFIG_VALUE__`)
	env, err = service.ready.Await(ctx)
	if err != nil {
		t.Fatal(err)
	}
	readFrontendFixture(t, ctx, env.result.GetPrivateUrl()+replaced.GetSession().GetRoutePrefix()+"Viewer.ts", http.StatusOK, `"replacement"`)
	t.Log("Vite configuration replacement rebuilt")

	// Project configuration replacement joins the previous compiler before readiness.
	writeFrontendFixture(t, root, "Other.ts", `import.meta.hot.accept(); export const value = __CONFIG_VALUE__`)
	replacementProject := &bldr_project.ProjectConfig{}
	if err := bldr_project.UnmarshalProjectConfig([]byte(strings.ReplaceAll(projectJSON, "Viewer.ts", "Other.ts")), replacementProject); err != nil {
		t.Fatal(err)
	}
	replacementConfig := conf.CloneVT()
	replacementConfig.ProjectConfig = replacementProject
	if err := service.configure(replacementConfig); err != nil {
		t.Fatal(err)
	}

	// Replace the subscription only after the old graph ends.
	awaitFrontendReplacement(t, ctx, service, replaced.GetSession().GetId())
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	stream = &frontendWatchStream{ctx: ctx, events: make(chan *frontend.Event, 64)}
	done = make(chan error, 1)
	go func() { done <- service.Watch(&frontend.WatchRequest{}, stream) }()
	projectReplaced := nextFrontendEvent(t, ctx, stream, done)

	// Fetch the new entrypoint and confirm the previous compiler was joined.
	env, err = service.ready.Await(ctx)
	if err != nil {
		t.Fatal(err)
	}
	readFrontendFixture(t, ctx, env.result.GetPrivateUrl()+projectReplaced.GetSession().GetRoutePrefix()+"Other.ts", http.StatusOK, `"replacement"`)
	if syscall.Kill(replacementPID, 0) != syscall.ESRCH {
		t.Fatal("project configuration replacement left the old compiler running")
	}
	replacementPID = frontendFixturePID(t, root)
	t.Log("project configuration replacement rebuilt")

	// Project shutdown joins the compiler and ends its update connection.
	service.Close()
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if service.active.Load() != nil || syscall.Kill(replacementPID, 0) != syscall.ESRCH {
		t.Fatal("shutdown left a compiler process or active environment")
	}
	t.Log("shutdown joined every compiler")
}

// awaitFrontendReplacement follows readiness notifications until the graph changes.
func awaitFrontendReplacement(t *testing.T, ctx context.Context, service *FrontendService, previous string) {
	t.Helper()
	for {
		// Capture readiness and its replacement notification together.
		prom, changed := service.ready.GetPromise()
		if prom != nil {
			env, err := prom.AwaitWithCancelCh(ctx, changed)
			if err == nil && env != nil && env.result.GetSession().GetId() != previous {
				return
			}
		}

		// A replaced promise or a canceled test ends this wait.
		select {
		case <-changed:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
}

// writeFrontendFixture saves source through the real filesystem watcher.
func writeFrontendFixture(t *testing.T, root, name, source string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}

// frontendFixturePID reads the process identity published by the fixture plugin.
func frontendFixturePID(t *testing.T, root string) int {
	// Read the compiler's identity only after its readiness event.
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "compiler.pid"))
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		t.Fatal(err)
	}
	return pid
}

// nextFrontendEvent waits on the service stream or its terminal result.
func nextFrontendEvent(t *testing.T, ctx context.Context, stream *frontendWatchStream, done <-chan error) *frontend.Event {
	t.Helper()
	select {
	case event := <-stream.events:
		t.Logf("frontend event: %s", event.GetPayload())
		return event
	case err := <-done:
		t.Fatalf("frontend watch ended: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	return nil
}

// awaitFrontendPayload drains notifications until the requested Vite event arrives.
func awaitFrontendPayload(t *testing.T, ctx context.Context, stream *frontendWatchStream, done <-chan error, contains ...string) {
	t.Helper()
	for {
		event := nextFrontendEvent(t, ctx, stream, done)
		for _, part := range contains {
			if strings.Contains(event.GetPayload(), part) {
				return
			}
		}
	}
}

// readFrontendFixture checks the compiler's HTTP result with request cancellation.
func readFrontendFixture(t *testing.T, ctx context.Context, address string, status int, contains string) {
	// Request the module from the compiler capability used by the Fetch service.
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	// Compare both the response status and its current transformed source.
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status || !strings.Contains(string(body), contains) {
		t.Fatalf("GET %s: status %d, wanted %d containing %q; body %s", address, response.StatusCode, status, contains, body)
	}
}
