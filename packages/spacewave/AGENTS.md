# Building with Spacewave

Use this guide when adding the `spacewave` npm package to an application. Start with the [README](README.md) for setup and the [API reference](API.md) for signatures and error codes. The [task board](examples/task-board) is a complete integration with vanilla TypeScript and React clients sharing one server.

## Package and runtime

Install with `npm install spacewave`, or use `spacewave@next` for a prerelease. Use ESM and Node.js `>=24.15.0 <25` for the server. The package contains its compiled engine; consuming it requires no Go compiler or installation scripts.

| Import | Use |
| --- | --- |
| `spacewave` | `defineSchema`, `connect`, `SyncError`, and public types |
| `spacewave/server` | `createServer` and server configuration types; Node only |
| `spacewave/react` | Optional `createSyncContext(schema)` bindings for React `>=19.2.0 <20` |

Use these public entry points. Package internals and generated engine files are implementation details.

For server type checking, install `@types/node` 24.x and include `"node"` in `compilerOptions.types`. React applications also need their matching React types. Check the application against the installed package so repository aliases cannot hide missing dependencies or exports.

## Shared schema

Keep one browser-safe schema module containing the application ID, version, collection validators, and declared mutation inputs and outputs. Import it from clients and the server. Keep token verification, authorization, and mutation handlers in server modules.

Use a Standard Schema v1 validator, such as Zod. Collection writes take validator inputs; reads return stored validator outputs. Store strict JSON: null, booleans, finite numbers, strings, arrays, and plain objects. A missing record is `undefined`; a stored `null` is a value.

Treat the schema ID and version as the persisted application's compatibility contract. To change an existing dataset's schema, supply an explicit migration from its stored version. `createServer` completes that maintenance transaction before admitting clients. See [Migration](API.md#migration).

## Server and access

Create one server per dataset directory. `createServer({ directory, schema, authenticate, authorize, mutations })` opens durable storage and returns after startup and maintenance finish.

- Return a stable `subject` and `scope` from the application's real token verifier. The scope selects the caller's dataset. Supply `expiresAt` or a revocation `signal` when available.
- Grant collection actions through `authorize({ principal, scope, collection, action })`. Read and write permissions are separate, and subscription deliveries recheck read permission. A prefix is a query selector, not an authorization boundary; this RC has no row-level authorization.
- Use `server.as(principal)` for server-side work that follows the same policy. Reserve `server.admin(scope)` for deliberate privileged operations; it bypasses collection authorization.
- Use `server.listen()` for a listener Spacewave owns, or `server.attach(http)` to add sync to an existing HTTP server. Set `allowedOrigins` for the frontend and use `wss://` with HTTPS. Pass tokens through `getAccessToken`, keeping them out of URLs.
- Call `await server.close()` during shutdown. It closes its attachments, application, and storage. When using `attach(http)`, close the supplied HTTP server separately.

The example's `local-demo` verifier is only for local development. Replace it before serving other users.

## Client and React lifetime

Await `connect({ url, schema, getAccessToken })` before using the database. Admission includes authentication and schema compatibility checks. `getAccessToken(signal)` runs for each connection attempt; obtain the current token there. The client manages reconnection.

Keep a database for the application's intended connection lifetime. Use `collection(name).subscribe()` or `.watch()` for changing state, and render freshness alongside the data:

| Status | Meaning |
| --- | --- |
| `loading` | Waiting for the first snapshot |
| `current` | Latest received snapshot on a live connection |
| `stale` | Last received snapshot while reconnecting |
| `error` | Subscription ended with a failure |

Snapshots contain all matching `{ key, value }` records, ordered by key. They replace the previous snapshot. A query supports only `{ prefix?: string }`; it has no SQL, pagination, row filters, or deltas.

Forward cancellation signals to operations. Release subscriptions when views leave and close the database with `await db.close()` when its application lifetime ends.

For React, call `createSyncContext(schema)` once at module scope. Its `SyncProvider` receives an existing database; `useCollection` manages the component's subscription and `useDatabase` exposes the database for writes. The application closes the database. Keep top-level calls to `connect()` in client startup modules; importing a package entry alone is safe during server rendering.

## Writes, transactions, and recovery

Use collection `put` and `delete` for single-record changes. For an operation spanning records, declare a named mutation and implement its handler with `context.collection(name)` so its reads and writes share one transaction. The server validates mutation inputs and outputs.

Keep external side effects out of mutation handlers. Transaction rollback and acceptance receipts cover collection data and the returned result; they cannot roll back an email, payment, or unrelated HTTP request.

Each write has a request ID. When recovery matters, choose the ID before dispatch and keep the exact input available. Handle `SyncError` by its stable `code`:

- On `UNCERTAIN`, acceptance is unknown. Offer a retry of the same operation with the same request ID and exact input. An accepted duplicate returns the original result, including after a server restart.
- On `CONFLICT`, the request ID was reused with different input. Correct the caller's request handling.
- On `UNAVAILABLE`, the server or storage cannot currently serve the operation. Preserve the user's intent and let the application decide when to retry.

Receipts remain for the dataset lifetime. The client has no durable offline replica, offline write queue, or optimistic updates. Render accepted server snapshots and keep any pending form input separate from them. See [Uncertain writes](API.md#uncertain-writes) for a concrete retry example.

## Verify the integration

Run the application against the installed artifact and check the behavior the change affects:

- Type-check shared schema, server code, and client code with strict TypeScript.
- Connect two clients in the same scope and a server writer. Confirm accepted changes reach both clients, then close and reopen the server to verify persistence.
- Exercise rejected tokens, denied operations, and separate scopes. Confirm each caller sees only its authorized dataset.
- Interrupt and restore the connection. Check stale state, recovery, and reuse of the original request ID for an uncertain write.
- Unmount subscribed views and close clients and servers. Confirm subscriptions and processes finish.

Keep queries sized for the view: defaults cap them at 10,000 records and 8 MiB per snapshot, with 256 KiB per record. Changes produce complete snapshots, so record count alone does not describe the workload. Measure scan and delivery time, memory, and storage growth with representative data before raising limits. See the [README's current scope](README.md#current-scope).
