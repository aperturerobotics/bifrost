import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import {
  appendFile,
  cp,
  mkdir,
  readFile,
  rm,
  writeFile,
} from 'node:fs/promises'
import { dirname, join, relative, resolve } from 'node:path'
import { gzipSync } from 'node:zlib'
import { parseSync } from 'oxc-parser'

// Qualification installs one tarball and exercises only its public consumer APIs.
const root = process.cwd()
const working = join(root, '.tmp/sync-qualification')
const consumer = join(working, 'consumer')
const artifacts = resolve(
  process.env.SYNC_ARTIFACT_DIRECTORY ?? '.bldr-dist/sync-library',
)
const node = resolve(process.env.SYNC_NODE_BINARY ?? Bun.which('node') ?? '')
const bun = process.execPath
const checks: { name: string; elapsedMs: number }[] = []
await mkdir(artifacts, { recursive: true })
await rm(working, { recursive: true, force: true })
await mkdir(consumer, { recursive: true })

async function run(
  name: string,
  cmd: string[],
  cwd = root,
  env: Record<string, string> = {},
): Promise<string> {
  const started = performance.now()
  const log = join(working, name + '.log')
  await writeFile(log, '')
  const child = Bun.spawn(cmd, {
    cwd,
    env: { ...process.env, ...env },
    stdout: 'pipe',
    stderr: 'pipe',
  })
  async function consume(stream: ReadableStream<Uint8Array>): Promise<string> {
    const parts: Uint8Array[] = []
    for await (const chunk of stream) {
      parts.push(chunk)
      await appendFile(log, chunk)
    }
    return Buffer.concat(parts).toString('utf8')
  }
  const [stdout, stderr, code] = await Promise.all([
    consume(child.stdout),
    consume(child.stderr),
    child.exited,
  ])
  if (code) throw new Error(`${name} failed (${code}):\n${stdout}\n${stderr}`)
  checks.push({ name, elapsedMs: performance.now() - started })
  console.log(`Passed ${name}`)
  return stdout
}

const version = (await run('node-version', [node, '--version'])).trim()
assert.match(
  version,
  /^v24\./,
  'Set SYNC_NODE_BINARY to supported Node 24.15 or later within 24.x',
)
assert.ok(Number(version.split('.')[1]) >= 15)
if (!process.argv.includes('--skip-build'))
  await run('build', ['go', 'run', './scripts/sync-library'])
const metadata = JSON.parse(
  await readFile('packages/spacewave/dist/build.json', 'utf8'),
)
if (!process.argv.includes('--allow-dirty'))
  assert.equal(
    metadata.dirty,
    false,
    'Release qualification requires a clean source revision',
  )
const packed = JSON.parse(
  await run(
    'pack',
    [
      'npm',
      'pack',
      '--ignore-scripts',
      '--json',
      '--pack-destination',
      artifacts,
    ],
    join(root, 'packages/spacewave'),
  ),
)[0] as { filename: string; size: number; unpackedSize: number }
const tarball = join(artifacts, packed.filename)
const sha256 = createHash('sha256')
  .update(await readFile(tarball))
  .digest('hex')
await writeFile(tarball + '.sha256', `${sha256}  ${packed.filename}\n`)
await writeFile(
  join(consumer, 'package.json'),
  JSON.stringify(
    {
      private: true,
      type: 'module',
      dependencies: {
        spacewave: `file:${tarball}`,
        '@standard-schema/spec': '1.1.0',
        react: '19.2.8',
        'react-dom': '19.2.8',
        '@types/react': '19.2.18',
        '@types/react-dom': '19.2.4',
        '@types/node': '24.13.3',
        typescript: '7.0.2',
        esbuild: '0.27.0',
        playwright: '1.62.1',
        zod: '4.5.4',
      },
    },
    null,
    2,
  ),
)
await run('install', [bun, 'install', '--ignore-scripts'], consumer)
const distribution = join(consumer, 'node_modules/spacewave/dist')
assert.deepEqual(
  JSON.parse(await readFile(join(distribution, 'build.json'), 'utf8')),
  metadata,
)
const noGo = { PATH: `${dirname(node)}:/usr/bin:/bin` }
await run(
  'ssr-import',
  [
    node,
    '--input-type=module',
    '-e',
    "await import('spacewave'); await import('spacewave/server'); await import('spacewave/react')",
  ],
  consumer,
  noGo,
)

async function bundle(
  source: string,
  target: string,
  browser = false,
  external: string[] = ['spacewave', 'playwright'],
): Promise<void> {
  await run('bundle-' + target.replaceAll('.', '-'), [
    bun,
    'build',
    source,
    '--target=' + (browser ? 'browser' : 'node'),
    '--outfile=' + join(consumer, target),
    ...external.flatMap((name) => ['--external', name]),
  ])
}
await bundle('scripts/sync-library/consumer.test.ts', 'consumer.test.mjs')
await bundle('scripts/sync-library/live.test.ts', 'live.test.mjs')
await run(
  'public-process',
  [node, '--test', 'consumer.test.mjs', 'live.test.mjs'],
  consumer,
  noGo,
)

const internal = join(working, 'internal')
await cp(distribution, internal, { recursive: true })
await run('bundle-internal', [
  bun,
  'build',
  'scripts/sync-library/application.test.ts',
  '--target=node',
  '--outfile=' + join(internal, 'application.test.mjs'),
])
await run(
  'internal-transactions',
  [node, '--test', 'application.test.mjs'],
  internal,
  noGo,
)

await cp(
  'scripts/sync-library/types.test.tsx',
  join(consumer, 'types.test.tsx'),
)
await writeFile(
  join(consumer, 'tsconfig.json'),
  JSON.stringify(
    {
      compilerOptions: {
        strict: true,
        noEmit: true,
        target: 'ES2024',
        module: 'NodeNext',
        moduleResolution: 'NodeNext',
        jsx: 'react-jsx',
        lib: ['ES2024', 'DOM', 'ESNext.Disposable'],
        types: ['node', 'react'],
        skipLibCheck: false,
      },
      include: ['types.test.tsx'],
    },
    null,
    2,
  ),
)
await run('consumer-types', [bun, 'x', 'tsgo', '-p', 'tsconfig.json'], consumer)
await bundle('e2e/sync-library/browser.ts', 'browser.mjs', true, ['spacewave'])
await cp('e2e/sync-library/react.tsx', join(consumer, 'react-fixture.tsx'))
await cp('e2e/sync-library/schema.ts', join(consumer, 'schema.ts'))
await run(
  'bundle-react',
  [
    bun,
    'build',
    'react-fixture.tsx',
    '--target=browser',
    '--outfile=react-fixture.mjs',
  ],
  consumer,
)
await bundle('e2e/sync-library/journey.test.ts', 'journey.test.mjs')
await run(
  'browser-journey',
  [node, '--test', 'journey.test.mjs'],
  consumer,
  noGo,
)

const example = join(working, 'example')
await cp(
  join(consumer, 'node_modules/spacewave/examples/task-board'),
  example,
  { recursive: true },
)
const exampleManifest = JSON.parse(
  await readFile(join(example, 'package.json'), 'utf8'),
)
exampleManifest.dependencies.spacewave = `file:${tarball}`
await writeFile(
  join(example, 'package.json'),
  JSON.stringify(exampleManifest, null, 2),
)
await run('example-install', [bun, 'install', '--ignore-scripts'], example)
await run('example-types', [bun, 'x', 'tsgo', '-p', 'tsconfig.json'], example)
await bundle('e2e/sync-library/example.test.ts', 'example.test.mjs')
await run('example-browser', [node, '--test', 'example.test.mjs'], consumer, {
  ...noGo,
  SYNC_EXAMPLE_DIRECTORY: example,
})

// Entry closures expose accidental Node, React, or workspace imports in the client.
async function closure(entry: string): Promise<{
  files: string[]
  bytes: number
  gzipBytes: number
  external: string[]
}> {
  const files = new Set<string>()
  const external = new Set<string>()
  let bytes = 0
  let gzipBytes = 0
  async function visit(path: string): Promise<void> {
    if (files.has(path)) return
    files.add(path)
    const text = await readFile(path, 'utf8')
    bytes += Buffer.byteLength(text)
    gzipBytes += gzipSync(text).length
    const parsed = parseSync(path, text)
    assert.equal(parsed.errors.length, 0)
    for (const imported of parsed.module.staticImports) {
      const name = imported.moduleRequest.value
      if (name.startsWith('.')) await visit(resolve(dirname(path), name))
      else external.add(name)
    }
  }
  await visit(join(distribution, entry + '.mjs'))
  return {
    files: [...files].map((file) => relative(distribution, file)),
    bytes,
    gzipBytes,
    external: [...external].sort(),
  }
}
const bundles = Object.fromEntries(
  await Promise.all(
    ['index', 'react', 'server', 'engine-worker'].map(async (name) => [
      name,
      await closure(name),
    ]),
  ),
)
const buildReport = JSON.parse(
  await readFile('.tmp/sync-library/build-report.json', 'utf8'),
) as { inputs: string[] }
const compiledPackages = [
  ...new Set(
    buildReport.inputs.flatMap((path) => {
      const source = path.split('/@goscript/github.com/s4wave/spacewave/')[1]
      return source ? [dirname(source)] : []
    }),
  ),
].sort()
const forbidden = [
  'app/',
  'core/app/',
  'core/session/',
  'net/transport/',
  'net/webrtc/',
  'net/circuit/',
]
assert.ok(
  compiledPackages.every((path) =>
    forbidden.every((prefix) => !(path + '/').startsWith(prefix)),
  ),
)
assert.deepEqual(bundles.index.external, [])
assert.ok(
  bundles.react.external.every(
    (name: string) => name === 'react' || name === 'react/jsx-runtime',
  ),
)
assert.ok(
  bundles.server.external.every((name: string) => name.startsWith('node:')),
)
assert.ok(
  bundles['engine-worker'].external.every((name: string) =>
    name.startsWith('node:'),
  ),
)
await bundle('scripts/sync-library/workload.ts', 'workload.mjs')
const workloadLog = await run(
  'workload',
  [node, 'workload.mjs'],
  consumer,
  noGo,
)
const workload = JSON.parse(
  workloadLog
    .trim()
    .split('\n')
    .findLast((line) => line.startsWith('{'))!,
)
const report = {
  artifact: packed.filename,
  sha256,
  ...metadata,
  node: version,
  platform: process.platform,
  arch: process.arch,
  packedBytes: packed.size,
  unpackedBytes: packed.unpackedSize,
  bundles,
  compiledPackages,
  workload,
  checks,
}
await writeFile(
  join(artifacts, 'qualification.json'),
  JSON.stringify(report, null, 2) + '\n',
)
console.log(JSON.stringify(report, null, 2))
