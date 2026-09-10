# Spacewave

Typed live collections for TypeScript applications. Define a shared schema, run a Node server, and subscribe from browsers or React. Accepted writes persist in a Spacewave World backed by SQLite; other connected clients receive the resulting state.

This release candidate supports Node 24.15–24.x and ESM. React is optional and requires React 19.2–19.x. Browser clients use the native WebSocket API. The package includes its compiled engine; installation needs neither Go nor lifecycle scripts.

## Try the task board

Install the release-candidate tarball, then copy its complete example:

```sh
npm install ./spacewave-0.1.0-rc.1.tgz
cp -R node_modules/spacewave/examples/task-board ./task-board
cd task-board
npm install ../spacewave-0.1.0-rc.1.tgz
npm start
```

Open `http://127.0.0.1:8787` for vanilla TypeScript or `/react` for React. Open both pages, add a task in one, and complete it in the other. Restart the server to reopen the same `.data` directory. The example's `local-demo` token is a local identity; replace its verifier before exposing the server to other users.

The [shared schema](examples/task-board/schema.ts), [server](examples/task-board/server.ts), [vanilla client](examples/task-board/vanilla.ts), and [React client](examples/task-board/react.tsx) form one application. The server also writes through the same collection API.

## Share a contract

```ts
import { defineSchema } from 'spacewave'
import { z } from 'zod'

export const schema = defineSchema({
  id: 'notes',
  version: 1,
  collections: { notes: z.object({ text: z.string() }) },
  mutations: {},
})
```

Any Standard Schema v1 validator can supply runtime validation and inferred TypeScript types. Writes accept validator inputs; reads return stored validator outputs. Values must be strict JSON. Missing records return `undefined`; stored `null` remains `null`.

```ts
import { connect } from 'spacewave'
import { schema } from './schema.js'

const db = await connect({
  url: 'ws://localhost:8787/sync',
  schema,
  getAccessToken: () => 'local-demo',
})
const notes = db.collection('notes')
const unsubscribe = notes.subscribe({}, {
  next: ({ status, data }) => console.log(status, data),
})
await notes.put('welcome', { text: 'Hello from Spacewave' })
// At application shutdown:
unsubscribe()
await db.close()
```

```ts
import { createServer } from 'spacewave/server'
import { SyncError } from 'spacewave'
import { schema } from './schema.js'

const server = await createServer({
  directory: './data',
  schema,
  authenticate: async (token) => {
    if (token !== 'local-demo') throw new SyncError('AUTHENTICATION', 'Sign in again')
    return { subject: 'local-user', scope: 'local' }
  },
  authorize: ({ scope }) => scope === 'local',
  mutations: {},
})
await server.listen({ port: 8787 })
```

Import `createSyncContext(schema)` from `spacewave/react` for `SyncProvider`, `useDatabase`, and `useCollection`. The provider accepts an existing database; unmounting releases subscriptions, and the application closes its database. Importing any package entry during SSR starts no connections or workers.

For server TypeScript projects, install `@types/node` 24.x and include `node` in `compilerOptions.types`. React projects also need the matching React type packages. Public declarations resolve inside this package and its declared dependencies.

## Behavior and limits

- The server owns accepted state. A subscription delivers complete, ordered prefix snapshots with `loading`, `current`, `stale`, or `error` status. Reconnection refreshes the snapshot. Slow consumers retain the latest pending snapshot.
- Policies grant collection reads or writes within an authenticated scope. Each operation and each subscription delivery checks authority. Tokens travel in the authentication exchange, and never in the URL.
- Collection writes and named mutations commit with a durable acceptance receipt. On `UNCERTAIN`, retain the request ID and exact input. Retrying returns the accepted result without executing the write again; conflicting reuse fails. Receipts remain for the dataset lifetime.
- There is no durable offline replica, offline write queue, optimistic state, row-level authorization, or SQL query interface in this RC. Schema changes require an explicit maintenance migration before admission.
- Defaults bound queries to 10,000 records and 8 MiB of encoded keys and values, and each record to 256 KiB. These are safety bounds, not throughput claims. One server writer owns each dataset directory.

A qualification workload on Node 24.21.0, macOS arm64 used 1,000 numeric records, three subscribers, and ten writes. With a shared producer for identical queries, all-subscriber delivery p50/p95 was 947/1,041 ms and acceptance was 234/242 ms. Ready RSS was 336 MB and RSS after the workload was 1.10 GB. The 28.9 KB snapshot still requires a complete scan per changed revision. Bulk seeding took 13.57 seconds. This RC is suitable for evaluating small live datasets; larger workloads need measurement. RSS includes the compiled runtime and is not a retained-heap measurement. The release artifact's `qualification.json` records its exact sample, sizes, and checksum.

Read the [API reference](API.md), [agent setup guide](AGENTS.md), and [release notes](CHANGELOG.md).
