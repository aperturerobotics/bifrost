# Local CDN startup measurement

Run the cold Chromium landing-to-Drive scenario with:

```sh
bun run test:release:web:startup
```

The harness incrementally builds minified release artifacts, exports the
startup plugins into real CDN KVF packs, and serves them on a fresh localhost
port. The production bootstrap embeds only the launcher and
materializer. The browser fetches a signed distribution fixture, the CDN root
pointer, and plugin packs through the production loading path. Each run starts
with an empty browser context and reports navigation-to-file and click-to-file
times after reaching `getting-started.md` and checking the invitation dialog.
The output records Chromium's actual renderer so hardware-accelerated and
software-rendered runs can be distinguished.

Build and publication state live in `.bldr-startup`. Bldr checks source changes
on every run and reuses unchanged manifests. The build keeps the release
manifest selection rules, with JavaScript and entrypoint minification enabled.
Build time is outside the measured browser interval.

Capture the root worker's runtime trace during the same scenario with:

```sh
E2E_RELEASE_WASM_MANIFEST_STARTUP_TRACE=1 bun run test:release:web:startup
```

The trace is written to
`.bldr/e2e-releasewasm/artifacts/local-cdn-startup.trace`. Trace-enabled runs
include instrumentation overhead and should be compared separately from timings
without tracing.

Browser proxy isolation blocks external network access while retaining normal
HTTP caching. The optional public-content catalog is unavailable, and the local
provider has no cloud signaling connection. This scenario measures local
startup and storage costs; it does not measure internet delivery latency.

## Saved Canvas reopening

Run the same-document Changelog-to-Canvas scenario with:

```sh
bun run test:release:web:reopen
E2E_RELEASE_WASM_BROWSER=webkit bun run test:release:web:reopen
```

The default browser is Chromium. Setup creates a local Canvas through
Quickstart and waits for its saved file node. Each of four samples visits
Changelog, waits for the Canvas to disappear, dwells there for six seconds,
then navigates back using the hash router. Setup and dwell are outside the
measurement. The first reopen is reported separately from the next three;
all share the same document, runtime, storage, and browser cache. The dwell
exceeds Quickstart's five-second handoff release grace; the handoff assertion
guards against accidentally benchmarking those retained resources.

Readiness requires the Canvas viewport and its nested file browser displaying
the saved folder's empty state. The scenario rejects Quickstart handoff reuse and new worker construction
during reopening. It records observations without enforcing a speed threshold.
Console output includes elapsed time and server-side pack/root request deltas.

Each run writes `sample-1.json` through `sample-4.json` under a unique
`.bldr/e2e-releasewasm/artifacts/canvas-reopen-<browser>-<timestamp>/` directory.
Samples contain browser-clock boundaries, loading-state transitions, existing
mount Performance marks, and document Resource Timing entries. Document resource
entries do not include all worker requests; use the server counters and optional
runtime trace for that work. A timed-out sample retains partial marks and page
text. No screenshots are captured by this scenario.

In Chromium, set `E2E_RELEASE_WASM_MANIFEST_STARTUP_TRACE=1` for the runtime trace
collector. Its `runtime.trace` includes setup and all four samples; use the
sample boundaries to distinguish intervals, and compare traced runs separately.
This local fixture measures navigation and storage costs, not staging network
latency or the contents of an existing user's Canvas.
