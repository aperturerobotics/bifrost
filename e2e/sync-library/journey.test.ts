import assert from 'node:assert/strict'
import { createServer as createHTTPServer } from 'node:http'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { dirname, join, resolve, sep } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'
import { chromium, webkit } from 'playwright'
import type { Page } from 'playwright'
import { SyncError } from 'spacewave'
import { createServer } from 'spacewave/server'
import { schema } from './schema.js'
import type {} from './browser.js'

const consumer = dirname(fileURLToPath(import.meta.url))
const distribution = join(consumer, 'node_modules/spacewave/dist')

async function current(page: Page, id: string, key: string): Promise<void> {
  await page.waitForFunction(
    ({ id, key }) => {
      const state = window.fixture.latest(id)
      return (
        state?.status === 'current' &&
        state.data.some((entry) => entry.key === key)
      )
    },
    { id, key },
  )
}

async function eventually(check: () => Promise<boolean>): Promise<void> {
  const signal = AbortSignal.timeout(10_000)
  while (!(await check())) {
    signal.throwIfAborted()
    await new Promise((resolve) => setTimeout(resolve, 20))
  }
}

for (const browserType of [chromium, webkit]) {
  test(
    `packed task-board journey in ${browserType.name()}`,
    { timeout: 120_000 },
    async () => {
      const directory = await mkdtemp(
        join(tmpdir(), `sync-${browserType.name()}-`),
      )
      const revoked = new AbortController()
      const mutationEntered = Promise.withResolvers<void>()
      const continueMutation = Promise.withResolvers<void>()
      let increments = 0
      const server = await createServer({
        directory,
        schema,
        limits: { maxRecords: 3, maxSnapshotBytes: 1024, maxRecordBytes: 256 },
        authenticate: async (token) => {
          if (token === 'bad')
            throw new SyncError('AUTHENTICATION', 'Sign in again')
          return {
            subject: token,
            scope: token === 'outside' ? 'outside' : 'team',
            signal: token === 'revocable' ? revoked.signal : undefined,
            expiresAt: token === 'expiring' ? Date.now() + 1000 : undefined,
          }
        },
        authorize: ({ principal, collection, action }) =>
          collection !== 'secrets' &&
          !(principal.subject === 'reader' && action === 'write') &&
          !(principal.subject === 'writer' && action === 'read'),
        mutations: {
          increment: async (context) => {
            increments++
            const counts = context.collection('counts')
            const value = ((await counts.get('counter')) ?? 0) + 1
            await counts.put('counter', value)
            mutationEntered.resolve()
            await continueMutation.promise
            return value
          },
        },
      })
      const http = createHTTPServer((request, response) => {
        const route = new URL(request.url ?? '/', 'http://fixture').pathname
        if (route === '/') {
          response.setHeader('content-type', 'text/html')
          response.end(
            '<!doctype html><title>Sync acceptance</title><script type="importmap">{"imports":{"spacewave":"/pkg/index.mjs"}}</script><script type="module" src="/fixture.mjs"></script>',
          )
          return
        }
        if (route === '/health') {
          response.end('ready')
          return
        }
        const path =
          route === '/fixture.mjs'
            ? join(consumer, 'browser.mjs')
            : resolve(distribution, '.' + route.slice(4))
        if (
          route !== '/fixture.mjs' &&
          (!route.startsWith('/pkg/') || !path.startsWith(distribution + sep))
        ) {
          response.writeHead(404)
          response.end()
          return
        }
        void readFile(path).then(
          (data) => {
            response.setHeader('content-type', 'text/javascript')
            response.end(data)
          },
          () => {
            response.writeHead(404)
            response.end()
          },
        )
      })
      const attachment = server.attach(http)
      await new Promise<void>((resolve) => http.listen(0, '127.0.0.1', resolve))
      const address = http.address()
      assert.ok(address && typeof address !== 'string')
      const origin = `http://127.0.0.1:${address.port}`
      const url = origin.replace('http:', 'ws:') + '/sync'
      const browser = await browserType.launch({ headless: true })
      const context = await browser.newContext()
      const pages: Page[] = []
      const pageErrors: string[] = []
      const open = async (token: string) => {
        const page = await context.newPage()
        pages.push(page)
        page.on('pageerror', (error) => pageErrors.push(error.message))
        await page.goto(origin)
        await page.waitForFunction(() => Boolean(window.fixture))
        const result = await page.evaluate(
          ({ url, token }) => window.fixture.start(url, token),
          { url, token },
        )
        return { page, result }
      }
      try {
        const { page: alice, result: aliceResult } = await open('alice')
        const { page: bob, result: bobResult } = await open('bob')
        assert.equal(aliceResult.code, undefined)
        assert.equal(bobResult.code, undefined)
        await alice.evaluate(() => window.fixture.watch('all'))
        await bob.evaluate(() => window.fixture.watch('all'))
        assert.equal(
          (
            await alice.evaluate(() =>
              window.fixture.put('a', { title: 'Ship', done: false }),
            )
          ).code,
          undefined,
        )
        await current(bob, 'all', 'a')
        await server
          .as({ subject: 'server', scope: 'team' })
          .collection('todos')
          .put('b', { title: 'Server write', done: true })
        await current(alice, 'all', 'b')
        assert.equal(
          (await alice.evaluate(() => window.fixture.put('bad', { title: 42 })))
            .code,
          'VALIDATION',
        )
        assert.equal(
          (await alice.evaluate(() => window.fixture.scan('secrets'))).code,
          'DENIED',
        )
        const { page: reader } = await open('reader')
        assert.equal(
          (
            await reader.evaluate(() =>
              window.fixture.put('no', { title: 'Denied', done: false }),
            )
          ).code,
          'DENIED',
        )
        const { page: writer } = await open('writer')
        assert.equal(
          (
            await writer.evaluate(() =>
              window.fixture.put('a', { title: 'Write only', done: false }),
            )
          ).code,
          undefined,
        )
        assert.equal(
          (await writer.evaluate(() => window.fixture.get('a'))).code,
          'DENIED',
        )
        const { page: outside } = await open('outside')
        assert.equal(
          (await outside.evaluate(() => window.fixture.get('a'))).result,
          undefined,
        )
        const { result: bad } = await open('bad')
        assert.equal(bad.code, 'AUTHENTICATION')
        await Promise.all(
          [reader, writer, outside].map((page) =>
            page.evaluate(() => window.fixture.close()),
          ),
        )

        await bob.evaluate(() => window.fixture.disconnect())
        await bob.waitForFunction(
          () =>
            window.fixture.status().tokens > 1 &&
            window.fixture.status().state.status === 'ready',
        )
        await current(bob, 'all', 'b')
        assert.ok(
          (await bob.evaluate(() => window.fixture.states('all'))).includes(
            'stale',
          ),
        )

        await alice.evaluate(() =>
          window.fixture.beginLostReply('lost-increment'),
        )
        await mutationEntered.promise
        await alice.evaluate(() => window.fixture.discardReply())
        continueMutation.resolve()
        await eventually(
          async () =>
            (await server.admin('team').collection('counts').get('counter')) ===
            1,
        )
        await alice.evaluate(() => window.fixture.recoverReply())
        assert.deepEqual(await alice.evaluate(() => window.fixture.result()), {
          result: 1,
        })
        assert.equal(increments, 1)

        await bob.evaluate(() => window.fixture.watch('prefix', 'a'))
        await current(bob, 'prefix', 'a')
        await server
          .admin('team')
          .collection('todos')
          .put('c', { title: 'Third', done: false })
        await server
          .admin('team')
          .collection('todos')
          .put('d', { title: 'Fourth', done: false })
        await bob.waitForFunction(
          () => window.fixture.latest('all')?.error?.code === 'QUERY_LIMIT',
        )
        assert.equal(
          (await bob.evaluate(() => window.fixture.scan())).code,
          'QUERY_LIMIT',
        )
        await server
          .admin('team')
          .collection('todos')
          .put('a', { title: 'Prefix still alive', done: true })
        await bob.waitForFunction(
          () =>
            (
              window.fixture.latest('prefix')?.data[0]?.value as {
                title?: string
              }
            )?.title === 'Prefix still alive',
        )
        await bob.evaluate(() => window.fixture.stop('prefix'))
        await server.admin('team').collection('todos').delete('d')
        await bob.evaluate(() => window.fixture.watch('new'))
        await current(bob, 'new', 'c')
        assert.deepEqual(
          await bob.evaluate(() =>
            window.fixture.latest('new')?.data.map((entry) => entry.key),
          ),
          ['a', 'b', 'c'],
        )
        await server.admin('team').collection('todos').delete('b')
        await bob.waitForFunction(
          () => window.fixture.latest('new')?.data.length === 2,
        )

        const { page: revocable } = await open('revocable')
        await revocable.evaluate(() => window.fixture.watch('all'))
        await current(revocable, 'all', 'a')
        revoked.abort()
        await revocable.waitForFunction(
          () => window.fixture.status().state.status === 'closed',
        )
        assert.equal(
          await revocable.evaluate(
            () => window.fixture.status().state.error?.code,
          ),
          'AUTHENTICATION',
        )
        const { page: expiring } = await open('expiring')
        await expiring.waitForFunction(() => window.fixture.status().tokens > 1)
        await expiring.evaluate(() => window.fixture.close())

        await Promise.all(
          [alice, bob].map((page) =>
            page.evaluate(() => window.fixture.close()),
          ),
        )
        assert.equal(
          (await bob.evaluate(() => window.fixture.status())).openSockets,
          0,
        )
        assert.deepEqual(pageErrors, [])
        await attachment.close()
        assert.equal(await (await fetch(origin + '/health')).text(), 'ready')
      } finally {
        continueMutation.resolve()
        await Promise.allSettled(
          pages.map((page) => page.evaluate(() => window.fixture?.close())),
        )
        await context.close()
        await browser.close()
        await server.close()
        await new Promise<void>((resolve, reject) =>
          http.close((error) => (error ? reject(error) : resolve())),
        )
        await rm(directory, { recursive: true, force: true })
      }
    },
  )
}
