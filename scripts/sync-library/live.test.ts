import assert from 'node:assert/strict'
import { mkdtemp, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import { connect, defineSchema } from 'spacewave'
import { createServer } from 'spacewave/server'
import type { StandardSchemaV1 } from '@standard-schema/spec'

const text: StandardSchemaV1<string> = { '~standard': { version: 1, vendor: 'acceptance', validate: (value) => typeof value === 'string' ? { value } : { issues: [{ message: 'Expected text' }] } } }
const schema = defineSchema({ id: 'live-acceptance', version: 1, collections: { todos: text }, mutations: {} })

test('packed clients observe browser-style and server writes and close naturally', { timeout: 30_000 }, async () => {
  const directory = await mkdtemp(join(tmpdir(), 'spacewave-live-'))
  const server = await createServer({ directory, schema, mutations: {}, authenticate: async (token) => ({ subject: token, scope: 'team' }), authorize: () => true })
  const clients: Awaited<ReturnType<typeof connect<typeof schema>>>[] = []
  try {
    const listener = await server.listen({ port: 0 })
    clients.push(await connect({ url: listener.url, schema, getAccessToken: () => 'alice' }))
    clients.push(await connect({ url: listener.url, schema, getAccessToken: () => 'bob' }))
    assert.equal(clients[0].connection.current.status, 'ready')
    const snapshots: unknown[] = []
    const expected = Promise.withResolvers<void>()
    const stop = clients[1].collection('todos').subscribe({}, { next: (state) => {
      snapshots.push(state)
      if (state.status === 'current' && state.data.some((entry) => entry.key === 'server')) expected.resolve()
    }, error: expected.reject })
    await clients[0].collection('todos').put('browser', 'Ship the RC')
    assert.equal(await clients[1].collection('todos').get('browser'), 'Ship the RC')
    await server.as({ subject: 'server', scope: 'team' }).collection('todos').put('server', 'Accepted')
    await expected.promise
    stop()
    assert.ok(snapshots.length >= 2)
    await clients[0].collection('todos').delete('browser')
    assert.equal(await clients[1].collection('todos').get('browser'), undefined)
  } finally {
    await Promise.all(clients.map((client) => client.close()))
    await server.close()
    await rm(directory, { recursive: true, force: true })
  }
})
