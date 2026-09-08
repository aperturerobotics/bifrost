# Local CDN startup measurement

Run the cold Chromium landing-to-Drive scenario with:

```sh
bun run test:release:web:startup
```

The harness incrementally builds unminified release artifacts, exports the
startup plugins into real CDN KVF packs, and serves them on
`http://127.0.0.1:30772`. The production bootstrap embeds only the launcher and
materializer. The browser fetches a signed distribution fixture, the CDN root
pointer, and plugin packs through the production loading path. Each run starts
with an empty browser context and reports navigation-to-file and click-to-file
times after reaching `getting-started.md` and checking the invitation dialog.

Build and publication state live in `.bldr-startup`. Bldr checks source changes
on every run and reuses unchanged manifests. The build keeps the release
manifest selection rules, with JavaScript and entrypoint minification disabled.
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
