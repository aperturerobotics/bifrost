---
title: Build a live application with Spacewave
section: start
order: 2
summary: Share a typed schema, run a durable Node server, and subscribe from TypeScript or React.
---

The Spacewave sync library gives a TypeScript application typed collections and live updates backed by a durable server. The `spacewave` package contains the browser client and shared contract helpers. `spacewave/server` hosts the data on Node 24.15–24.x; `spacewave/react` adds optional React 19.2 bindings.

Start with the packaged task board. Its `schema.ts` declares task values and a `completeTodo` mutation using Standard Schema validators. The server imports that contract, verifies tokens, grants collection actions within a scope, and implements the mutation in one transaction. The vanilla and React clients import the same schema and subscribe to the resulting collection.

Install the release-candidate tarball, copy `node_modules/spacewave/examples/task-board`, install that example with the same tarball, and run `npm start` using Node 24. Open `http://127.0.0.1:8787` and `http://127.0.0.1:8787/react`. Add a task in one page and complete it in the other. Restarting the example reopens its `.data` directory. Replace the example's local token verifier before exposing it to other users.

`connect` resolves when authentication and compatibility checks complete. `collection(name)` provides typed `get`, `put`, `delete`, `scan`, `watch`, and `subscribe` operations. Subscriptions deliver complete ordered snapshots with loading, current, stale, and error states. Reconnecting refreshes server state. React's `createSyncContext(schema)` supplies a provider and typed hooks; the application retains ownership of the database lifetime.

A successful write includes durable acceptance. When a response is lost, `UNCERTAIN` carries the request ID needed for recovery. Retry with that ID and the same input to receive the original accepted result. Named mutations can change multiple collections atomically, but external side effects remain the application's responsibility.

The RC supports small online datasets with bounded complete snapshots. It has no offline write queue, durable browser replica, row-level policy, or SQL interface. Prefix queries reduce the size of a view; authorization still applies to the whole collection. Schema changes require a maintenance migration before server admission.

The package README includes working setup commands and measured limits. Its `API.md` documents options, typed errors, policies, migrations, and close behavior; `AGENTS.md` provides an implementation guide for coding agents. The release's qualification report identifies the exact artifact and records installed-consumer, browser, React, persistence, and workload checks.
