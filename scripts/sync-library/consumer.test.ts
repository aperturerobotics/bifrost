import assert from 'node:assert/strict'
import { mkdtemp, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import { Worker } from 'node:worker_threads'
import { spawnSync } from 'node:child_process'
import { openNodeEngine } from 'spacewave/server'

import { KeyValueStore, KVImplType } from '../../db/kvtx/block/kvtx.pb.js'
import { KvStore, KvStoreTypeID } from '../../sdk/kv/kv.js'
import type { Engine } from '../../sdk/world/engine.js'
import type { Client } from '../../bldr/sdk/resource/client.js'

// collection uses only the existing World and KV Resource contracts.
async function collection(
  engine: Engine,
  resources: Client,
  create: boolean,
): Promise<KvStore> {
  if (create) {
    const tx = await engine.newTransaction(true)
    try {
      const cursor = await tx.buildStorageCursor()
      try {
        const root = await cursor.putBlock({
          data: KeyValueStore.toBinary({
            implType: KVImplType.KV_IMPL_TYPE_IAVL,
          }),
        })
        const storage = await cursor.getRef()
        assert.ok(root.ref)
        const object = await tx.createObject('todos', {
          ...storage.ref,
          rootRef: root.ref,
        })
        object.release()
        const typeObject = await tx.createObject(`types/${KvStoreTypeID}`, {})
        typeObject.release()
        await tx.setGraphQuad('<todos>', '<type>', `<types/${KvStoreTypeID}>`)
        await tx.commit()
      } finally {
        cursor.release()
      }
    } finally {
      await tx.discard()
      tx.release()
    }
  }
  const typed = await engine.accessTypedObject('todos')
  assert.equal(typed.typeId, KvStoreTypeID)
  return new KvStore(resources.createResourceReference(typed.resourceId))
}

const encoder = new TextEncoder()

async function put(todos: KvStore, value: string): Promise<void> {
  await todos.withTransaction(true, async (tx) => {
    await tx.set(encoder.encode('first'), encoder.encode(value))
  })
}

async function get(todos: KvStore): Promise<string> {
  const result = await todos.get(encoder.encode('first'))
  assert.equal(result.found, true)
  return new TextDecoder().decode(result.data)
}

test(
  'installed Worker engine commits and reopens an ordinary World collection',
  { timeout: 60_000 },
  async () => {
    const directory = await mkdtemp(join(tmpdir(), 'spacewave-packed-'))
    assert.equal(spawnSync('go', ['version']).error?.code, 'ENOENT')
    try {
      const first = await openNodeEngine(directory)
      try {
        const todos = await collection(first.engine, first.resources, true)
        try {
          await put(todos, 'ship the RC')
        } finally {
          todos.release()
        }
        await assert.rejects(openNodeEngine(directory), /locked|writer|busy/i)
      } finally {
        await first.close()
      }
      const reopened = await openNodeEngine(directory)
      try {
        const todos = await collection(
          reopened.engine,
          reopened.resources,
          false,
        )
        try {
          assert.equal(await get(todos), 'ship the RC')
        } finally {
          todos.release()
        }
      } finally {
        await reopened.close()
        await reopened.close()
      }
    } finally {
      await rm(directory, { recursive: true, force: true })
    }
  },
)

test('installed engines isolate independent directories', { timeout: 60_000 }, async () => {
  const directory = await mkdtemp(join(tmpdir(), 'spacewave-isolated-'))
  const engines: Awaited<ReturnType<typeof openNodeEngine>>[] = []
  const stores: KvStore[] = []
  try {
    engines.push(await openNodeEngine(join(directory, 'a')))
    engines.push(await openNodeEngine(join(directory, 'b')))
    for (const engine of engines) {
      stores.push(await collection(engine.engine, engine.resources, true))
    }
    await Promise.all([put(stores[0], 'alpha'), put(stores[1], 'beta')])
    assert.deepEqual(await Promise.all(stores.map(get)), ['alpha', 'beta'])
  } finally {
    for (const store of stores) store.release()
    await Promise.all(engines.map((engine) => engine.close()))
    await rm(directory, { recursive: true, force: true })
  }
})

test('Worker loss rejects a pending request and preserves an acknowledged write', { timeout: 60_000 }, async () => {
  const directory = await mkdtemp(join(tmpdir(), 'spacewave-worker-loss-'))
  let worker: Worker | undefined
  const originalPost = Worker.prototype.postMessage
  // Capture the actual packaged Worker so the test can terminate it externally.
  Worker.prototype.postMessage = function (...args) {
    worker = this
    return originalPost.apply(this, args)
  }
  let engine: Awaited<ReturnType<typeof openNodeEngine>>
  try {
    engine = await openNodeEngine(directory)
  } finally {
    Worker.prototype.postMessage = originalPost
  }
  try {
    assert.ok(worker)
    const todos = await collection(engine.engine, engine.resources, true)
    try {
      await put(todos, 'accepted before failure')
    } finally {
      todos.release()
    }
    const pending = assert.rejects(engine.engine.waitSeqno(1_000_000n))
    await worker.terminate()
    await pending
    assert.equal(engine.signal.aborted, true)
    await assert.rejects(engine.close(), /Worker exited unexpectedly/)

    const reopened = await openNodeEngine(directory)
    try {
      const restored = await collection(reopened.engine, reopened.resources, false)
      try {
        assert.equal(await get(restored), 'accepted before failure')
      } finally {
        restored.release()
      }
    } finally {
      await reopened.close()
    }
  } finally {
    await engine.close().catch(() => {})
    await rm(directory, { recursive: true, force: true })
  }
})

test('failed startup releases the Worker and reports the cause', { timeout: 15_000 }, async () => {
  const directory = await mkdtemp(join(tmpdir(), 'spacewave-startup-'))
  try {
    const file = join(directory, 'file')
    await writeFile(file, 'not a directory')
    await assert.rejects(openNodeEngine(file), /directory|EEXIST|ENOTDIR/i)
  } finally {
    await rm(directory, { recursive: true, force: true })
  }
})
