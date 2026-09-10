import assert from 'node:assert/strict'
import { mkdtemp, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import type { StandardSchemaV1 } from '@standard-schema/spec'
import { createServer } from '../../core/sync/server.js'
import { openNodeEngine } from '../../core/sync/node/host.js'

import { Application } from '../../core/sync/application.js'
import { defineSchema, type Principal } from '../../sdk/sync/schema.js'
import { SyncError } from '../../sdk/sync/errors.js'
import { collectionKey } from '../../sdk/sync/keys.js'
import { getObjectType, setObjectType } from '../../sdk/world/types/types.js'

function validator<T>(
  check: (value: unknown) => value is T,
): StandardSchemaV1<T> {
  return {
    '~standard': {
      version: 1,
      vendor: 'test',
      validate: (value) =>
        check(value) ? { value } : { issues: [{ message: 'invalid' }] },
    },
  }
}

const number = validator(
  (value): value is number =>
    typeof value === 'number' && Number.isFinite(value),
)
const nullable = validator(
  (value): value is string | null =>
    value === null || typeof value === 'string',
)
const schema = defineSchema({
  id: 'sync-acceptance',
  version: 1,
  collections: { counts: number, audit: number, notes: nullable },
  mutations: {
    increment: { input: number, output: number },
    fail: { input: number, output: number },
  },
})
const principal: Principal = { subject: 'alice', scope: 'team-a' }

test(
  'internal server attachment preserves a supplied engine and existing ObjectTypes',
  { timeout: 60_000 },
  async () => {
    const directory = await mkdtemp(join(tmpdir(), 'spacewave-supplied-'))
    const owner = await openNodeEngine(directory)
    try {
      const publicSchema = defineSchema({
        id: 'public-api',
        version: 1,
        collections: { notes: nullable },
        mutations: {},
      })
      const server = await createServer({
        schema: publicSchema,
        engine: owner.engine,
        mutations: {},
        authenticate: async () => principal,
        authorize: ({ principal }) => principal.subject === 'alice',
      })
      try {
        await server
          .as(principal)
          .collection('notes')
          .put('\ufeffkey', null, { requestId: 'null' })
        assert.equal(
          await server.as(principal).collection('notes').get('\ufeffkey'),
          null,
        )
        assert.deepEqual(
          await server.as(principal).collection('notes').scan(),
          [{ key: '\ufeffkey', value: null }],
        )
        await assert.rejects(
          server
            .as({ subject: 'denied', scope: 'team-a' })
            .collection('notes')
            .put('x', 'secret'),
          { code: 'DENIED' },
        )
        await server.admin('team-b').collection('notes').put('admin', 'allowed')
        assert.equal(
          await server
            .as({ ...principal, scope: 'team-b' })
            .collection('notes')
            .get('admin'),
          'allowed',
        )
        const key = collectionKey(publicSchema.id, 'wrong-type', 'notes')
        const tx = await owner.engine.newTransaction(true)
        try {
          const object = await tx.createObject(key, {})
          object.release()
          await setObjectType(tx, key, 'different/type')
          await tx.commit()
        } finally {
          await tx.discard()
          tx.release()
        }
        await assert.rejects(
          server.admin('wrong-type').collection('notes').put('key', 'value'),
          { code: 'SCHEMA_MISMATCH' },
        )
        const read = await owner.engine.newTransaction(false)
        try {
          assert.equal(await getObjectType(read, key), 'different/type')
        } finally {
          await read.discard()
          read.release()
        }
      } finally {
        await server.close()
      }
      assert.ok(await owner.engine.getSeqno())
    } finally {
      await owner.close()
      await rm(directory, { recursive: true, force: true })
    }
  },
)

async function attach(
  directory: string,
  version = 1,
  migration?: Parameters<
    Application<typeof schema, Principal>['initialize']
  >[0],
) {
  const owner = await openNodeEngine(directory)
  let handlerRuns = 0
  const app = new Application({
    schema: { ...schema, version },
    engine: owner.engine,
    authorize: ({ principal }) => principal.subject !== 'denied',
    mutations: {
      increment: async (tx, amount) => {
        handlerRuns++
        const counts = tx.collection('counts')
        const value = ((await counts.get('total')) ?? 0) + amount
        await counts.put('total', value)
        await tx.collection('audit').put('total', value)
        return value
      },
      fail: async (tx, amount) => {
        await tx.collection('counts').put('total', amount)
        await tx.collection('audit').put('total', amount)
        throw new SyncError(
          'VALIDATION',
          'Business rule rejected the operation',
        )
      },
    },
  })
  try {
    await app.initialize(migration)
  } catch (error) {
    await app.close()
    await owner.close()
    throw error
  }
  return {
    app,
    owner,
    runs: () => handlerRuns,
    close: async () => {
      await app.close()
      await owner.close()
    },
  }
}

test(
  'compiled application atomically changes two ordinary collections and retains receipts',
  { timeout: 60_000 },
  async () => {
    const directory = await mkdtemp(join(tmpdir(), 'spacewave-application-'))
    try {
      let server = await attach(directory)
      try {
        const request = { kind: 'mutate', name: 'increment', input: 2 } as const
        assert.equal(
          await server.app.execute(principal, request, { requestId: 'once' }),
          2,
        )
        assert.equal(
          await server.app.execute(principal, request, { requestId: 'once' }),
          2,
        )
        assert.equal(server.runs(), 1)
        await assert.rejects(
          server.app.execute(
            principal,
            { ...request, input: 3 },
            { requestId: 'once' },
          ),
          { code: 'CONFLICT' },
        )
        await assert.rejects(
          server.app.execute(principal, {
            kind: 'mutate',
            name: 'fail',
            input: 99,
          }),
          { code: 'VALIDATION' },
        )
        for (const collection of ['counts', 'audit']) {
          assert.deepEqual(
            await server.app.execute(principal, {
              kind: 'get',
              collection,
              key: 'total',
            }),
            { found: true, value: 2 },
          )
        }
        const accepted = await Promise.all([
          server.app.execute(principal, {
            kind: 'mutate',
            name: 'increment',
            input: 1,
          }),
          server.app.execute(principal, {
            kind: 'mutate',
            name: 'increment',
            input: 1,
          }),
        ])
        assert.deepEqual(accepted.sort(), [3, 4])
        await server.app.close()
        assert.ok(await server.owner.engine.getSeqno())
      } finally {
        await server.close()
      }
      server = await attach(directory)
      try {
        assert.equal(
          await server.app.execute(
            principal,
            { kind: 'mutate', name: 'increment', input: 2 },
            { requestId: 'once' },
          ),
          2,
        )
        assert.equal(server.runs(), 0)
        assert.deepEqual(
          await server.app.execute(principal, {
            kind: 'get',
            collection: 'counts',
            key: 'total',
          }),
          { found: true, value: 4 },
        )
      } finally {
        await server.close()
      }
    } finally {
      await rm(directory, { recursive: true, force: true })
    }
  },
)

test(
  'compiled application distinguishes null, validates writes, and isolates authorized scopes',
  { timeout: 60_000 },
  async () => {
    const directory = await mkdtemp(join(tmpdir(), 'spacewave-policy-'))
    try {
      const server = await attach(directory)
      try {
        await server.app.execute(principal, {
          kind: 'put',
          collection: 'notes',
          key: 'nullable',
          value: null,
        })
        assert.deepEqual(
          await server.app.execute(principal, {
            kind: 'get',
            collection: 'notes',
            key: 'nullable',
          }),
          { found: true, value: null },
        )
        assert.deepEqual(
          await server.app.execute(principal, {
            kind: 'get',
            collection: 'notes',
            key: 'missing',
          }),
          { found: false },
        )
        assert.deepEqual(
          await server.app.execute(
            { ...principal, scope: 'team-b' },
            { kind: 'scan', collection: 'notes', prefix: '' },
          ),
          [],
        )
        await assert.rejects(
          server.app.execute(
            { ...principal, subject: 'denied' },
            { kind: 'get', collection: 'notes', key: 'nullable' },
          ),
          { code: 'DENIED' },
        )
        await assert.rejects(
          server.app.execute(principal, {
            kind: 'put',
            collection: 'notes',
            key: 'bad',
            value: 42,
          }),
          { code: 'VALIDATION' },
        )
        await assert.rejects(
          server.app.execute(principal, {
            kind: 'get',
            collection: 'unknown',
            key: 'x',
          }),
          { code: 'MISSING_COLLECTION' },
        )
        const expired = { ...principal, expiresAt: Date.now() - 1 }
        await assert.rejects(
          server.app.execute(expired, {
            kind: 'scan',
            collection: 'notes',
            prefix: '',
          }),
          { code: 'AUTHENTICATION' },
        )
        const revoke = new AbortController()
        revoke.abort()
        await assert.rejects(
          server.app.execute(
            { ...principal, signal: revoke.signal },
            { kind: 'scan', collection: 'notes', prefix: '' },
          ),
          { code: 'AUTHENTICATION' },
        )
      } finally {
        await server.close()
      }
    } finally {
      await rm(directory, { recursive: true, force: true })
    }
  },
)

test(
  'maintenance migration commits data and version together and recovers after failure',
  { timeout: 60_000 },
  async () => {
    const directory = await mkdtemp(join(tmpdir(), 'spacewave-migrate-'))
    try {
      const first = await attach(directory)
      try {
        await first.app.execute(principal, {
          kind: 'put',
          collection: 'counts',
          key: 'total',
          value: 1,
        })
      } finally {
        await first.close()
      }
      await assert.rejects(attach(directory, 2), { code: 'SCHEMA_MISMATCH' })
      await assert.rejects(
        attach(directory, 2, {
          from: 1,
          run: async (tx) => {
            await tx
              .scope(principal.scope)
              .collection('counts')
              .put('total', 99)
            throw new Error('interrupted migration')
          },
        }),
        /interrupted migration/,
      )
      const afterFailure = await attach(directory)
      try {
        assert.deepEqual(
          await afterFailure.app.execute(principal, {
            kind: 'get',
            collection: 'counts',
            key: 'total',
          }),
          { found: true, value: 1 },
        )
      } finally {
        await afterFailure.close()
      }
      const migrated = await attach(directory, 2, {
        from: 1,
        run: async (tx) => {
          assert.deepEqual(await tx.scopes(), [principal.scope])
          await tx.scope(principal.scope).collection('counts').put('total', 10)
        },
      })
      try {
        assert.deepEqual(
          await migrated.app.execute(principal, {
            kind: 'get',
            collection: 'counts',
            key: 'total',
          }),
          { found: true, value: 10 },
        )
      } finally {
        await migrated.close()
      }
      const reopened = await attach(directory, 2)
      await reopened.close()
    } finally {
      await rm(directory, { recursive: true, force: true })
    }
  },
)
