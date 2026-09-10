import { createServer as createHTTPServer } from 'node:http'
import { readFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { build } from 'esbuild'
import { SyncError } from 'spacewave'
import { createServer } from 'spacewave/server'
import { schema } from './schema.ts'

const directory = dirname(fileURLToPath(import.meta.url))
await build({
  absWorkingDir: directory,
  entryPoints: ['vanilla.ts', 'react.tsx'],
  outdir: 'public',
  bundle: true,
  format: 'esm',
  platform: 'browser',
  target: 'es2024',
  define: { 'process.env.NODE_ENV': '"production"' },
})

const server = await createServer({
  directory: process.env.DATA_DIRECTORY ?? join(directory, '.data'),
  schema,
  // Replace the local demo identity with your application's token verifier.
  authenticate: async (token) => {
    if (token !== 'local-demo')
      throw new SyncError('AUTHENTICATION', 'Sign in again')
    return { subject: 'demo-user', scope: 'demo' }
  },
  authorize: ({ scope }) => scope === 'demo',
  mutations: {
    completeTodo: async (context, { id }) => {
      const todos = context.collection('todos')
      const todo = await todos.get(id)
      if (!todo) throw new SyncError('VALIDATION', 'This task no longer exists')
      const updated = { ...todo, done: true }
      await todos.put(id, updated)
      return updated
    },
  },
})

const tasks = server
  .as({ subject: 'demo-user', scope: 'demo' })
  .collection('todos')
if (!(await tasks.get('welcome'))) {
  await tasks.put('welcome', {
    title: 'Open this board in two tabs',
    done: false,
  })
}

const files = new Map([
  ['/', ['vanilla.html', 'text/html']],
  ['/react', ['react.html', 'text/html']],
  ['/vanilla.js', ['public/vanilla.js', 'text/javascript']],
  ['/react.js', ['public/react.js', 'text/javascript']],
  ['/style.css', ['style.css', 'text/css']],
])
const http = createHTTPServer((request, response) => {
  const path = new URL(request.url ?? '/', 'http://localhost').pathname
  const file = files.get(path)
  if (!file) {
    response.writeHead(404)
    response.end()
    return
  }
  void readFile(join(directory, file[0])).then(
    (data) => {
      response.setHeader('Content-Type', file[1])
      response.end(data)
    },
    () => {
      response.writeHead(500)
      response.end('The page could not load')
    },
  )
})
server.attach(http)
const port = Number(process.env.PORT ?? 8787)
await new Promise<void>((resolve, reject) => {
  http.once('error', reject)
  http.listen(port, '127.0.0.1', resolve)
})
const address = http.address()
if (address && typeof address !== 'string')
  console.log(`Task board: http://127.0.0.1:${address.port}`)

let closing: Promise<void> | undefined
const close = () =>
  (closing ??= (async () => {
    await server.close()
    await new Promise<void>((resolve, reject) =>
      http.close((error) => (error ? reject(error) : resolve())),
    )
  })())
for (const signal of ['SIGINT', 'SIGTERM'] as const) {
  process.once(signal, () => {
    void close().catch((error: unknown) => {
      console.error(error)
      process.exitCode = 1
    })
  })
}
