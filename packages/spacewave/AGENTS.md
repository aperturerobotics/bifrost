# Building with Spacewave

Start from `examples/task-board`. Keep one browser-safe schema module containing the application ID, version, Standard Schema validators, and declared mutation inputs and outputs. Import it from the server and clients. Keep authentication, authorization, and mutation handlers in server files.

Use the public entries `spacewave`, `spacewave/server`, and optional `spacewave/react`. Run Node 24.15 or later within 24.x. Install the RC tarball directly; no Go compiler or package install scripts are required. Server type checking needs `@types/node` 24.x and `types: ["node"]`; React requires its corresponding peers and types.

Create one server per dataset directory. `createServer` completes startup and maintenance before returning. Use `server.listen()` to own a listener, or `server.attach(http)` to add sync to an existing HTTP server. Close the returned attachment and server at shutdown. An attachment does not own the supplied HTTP server.

Return a stable `subject` and `scope` from your real token verifier. Provide expiry or a revocation signal when available. Grant collection actions through `authorize`. Read and write permissions are separate; a prefix is a query, not an authorization boundary. Use `wss` when serving an HTTPS application. Keep tokens out of URLs.

Await `connect` before using the database. Render subscription data and freshness together. Unsubscribe when a view leaves; close the database when its application lifetime ends. A React provider receives an existing database and does not close it on unmount.

Choose stable request IDs for writes that may need recovery. On `UNCERTAIN`, preserve the exact input and request ID and offer a retry. Do not generate a replacement ID for that retry. Avoid side effects outside the transaction inside mutation handlers: rollback and accepted receipts cover collection data, not unrelated external systems.

Validate the actual installed package with strict TypeScript, two clients plus a server writer, reconnect, denied operations, and close/reopen persistence. Use the example's vanilla and React pages to exercise the same schema and dataset. Handle typed errors from the API reference. Size prefix queries for the data the view needs; snapshots are complete and bounded.
