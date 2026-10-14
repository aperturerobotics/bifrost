import { describe, expect, it, vi } from 'vitest'

import { SharedStreams } from '@aptre/bldr-sdk/resource/shared-streams.js'

import type { Engine } from './engine.js'
import { watchWorldQuery } from './world-query.js'
import type { IWorldState } from './world-state.js'

// fakeWorld is a live World that yields a new tracked snapshot when commit is
// called, as the server does after a commit touching the previous reads.
function fakeWorld() {
  let seqno = 1n
  let wake: (() => void) | undefined
  const snapshots: Array<{ [Symbol.dispose]: ReturnType<typeof vi.fn> }> = []
  const engine = {
    watchWorldState: vi.fn(async function* (signal?: AbortSignal) {
      for (let next = 1n; ; next++) {
        // Wait for the commit that produces the next snapshot.
        while (seqno < next) {
          // eslint-disable-next-line react-doctor/async-await-in-loop -- The fake waits for each commit in turn.
          await new Promise<void>((resolve, reject) => {
            wake = resolve
            signal?.addEventListener('abort', () => reject(signal.reason), {
              once: true,
            })
          })
        }
        const pinned = next
        const snapshot = {
          getSeqno: vi.fn(async () => ({ seqno: pinned })),
          [Symbol.dispose]: vi.fn(),
        }
        snapshots.push(snapshot)
        yield snapshot
      }
    }),
    [Symbol.dispose]: vi.fn(),
  }
  const streams = new SharedStreams()
  const owner = {
    id: 7,
    resourceRef: { createRef: vi.fn(() => engine) },
    watchShared: streams.watch.bind(streams),
  }
  return {
    engine,
    owner: owner as unknown as Engine,
    snapshots,
    commit() {
      seqno++
      wake?.()
    },
  }
}

vi.mock('./engine.js', () => ({
  // The watch wraps its own reference to the owner's Engine.
  Engine: function (ref: unknown) {
    return ref
  },
}))

describe('watchWorldQuery', () => {
  it('re-runs the query on each tracked snapshot', async () => {
    const world = fakeWorld()
    const query = vi.fn(
      async (_ws: IWorldState) => `run-${query.mock.calls.length}`,
    )
    const controller = new AbortController()
    const it = watchWorldQuery(world.owner, query, controller.signal)[
      Symbol.asyncIterator
    ]()

    await expect(it.next()).resolves.toEqual({
      done: false,
      value: { seqno: 1n, value: 'run-1' },
    })

    // The next run waits for the server's next snapshot.
    const next = it.next()
    world.commit()
    await expect(next).resolves.toEqual({
      done: false,
      value: { seqno: 2n, value: 'run-2' },
    })

    // Every run's snapshot is released before its result is yielded.
    expect(world.snapshots).toHaveLength(2)
    for (const snapshot of world.snapshots) {
      expect(snapshot[Symbol.dispose]).toHaveBeenCalledOnce()
    }

    // Aborting ends the wait and releases the Engine reference.
    const pending = it.next()
    controller.abort()
    await expect(pending).rejects.toBeDefined()
    expect(world.engine[Symbol.dispose]).toHaveBeenCalledOnce()
  })

  it('releases the snapshot when the query fails', async () => {
    const world = fakeWorld()
    const failure = new Error('query failed')
    const it = watchWorldQuery(world.owner, async () => {
      throw failure
    })[Symbol.asyncIterator]()

    await expect(it.next()).rejects.toBe(failure)
    expect(world.snapshots[0][Symbol.dispose]).toHaveBeenCalledOnce()
    expect(world.engine[Symbol.dispose]).toHaveBeenCalledOnce()
  })

  it('shares one run among watches with the same key', async () => {
    const world = fakeWorld()
    const query = vi.fn(async () => query.mock.calls.length)
    const first = new AbortController()
    const second = new AbortController()
    const a = watchWorldQuery(world.owner, query, first.signal, 'k')[
      Symbol.asyncIterator
    ]()
    const b = watchWorldQuery(world.owner, query, second.signal, 'k')[
      Symbol.asyncIterator
    ]()

    await expect(a.next()).resolves.toMatchObject({ value: { value: 1 } })
    await expect(b.next()).resolves.toMatchObject({ value: { value: 1 } })

    // A commit re-runs the query once for both watches.
    const nextA = a.next()
    const nextB = b.next()
    world.commit()
    await expect(nextA).resolves.toMatchObject({ value: { value: 2 } })
    await expect(nextB).resolves.toMatchObject({ value: { value: 2 } })
    expect(world.engine.watchWorldState).toHaveBeenCalledOnce()
    expect(query).toHaveBeenCalledTimes(2)

    // The shared run ends once the last watch leaves.
    first.abort()
    await expect(a.next()).rejects.toBeDefined()
    expect(world.engine[Symbol.dispose]).not.toHaveBeenCalled()
    const pending = b.next()
    second.abort()
    await expect(pending).rejects.toBeDefined()
    await vi.waitFor(() =>
      expect(world.engine[Symbol.dispose]).toHaveBeenCalledOnce(),
    )
  })
})
