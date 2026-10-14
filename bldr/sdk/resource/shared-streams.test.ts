import { describe, expect, it } from 'vitest'

import { SharedStreams } from './shared-streams.js'

// source is a pushed snapshot stream that records its opens and aborts.
function source() {
  // Queue pushed snapshots and wake the reader waiting for one.
  const queue: number[] = []
  let wake: (() => void) | undefined
  let fail: ((err: Error) => void) | undefined
  const state = {
    opens: 0,
    aborts: 0,
    push(value: number) {
      queue.push(value)
      wake?.()
    },
    fail(err: Error) {
      fail?.(err)
    },
    async *open(signal: AbortSignal): AsyncIterable<number> {
      state.opens++
      signal.addEventListener('abort', () => state.aborts++, { once: true })
      for (;;) {
        const next = queue.shift()
        if (next !== undefined) {
          yield next
          continue
        }
        await new Promise<void>((resolve, reject) => {
          wake = resolve
          fail = reject
          signal.addEventListener('abort', () => reject(signal.reason), {
            once: true,
          })
        })
      }
    },
  }
  return state
}

describe('SharedStreams', () => {
  it('shares one source between the watchers of a key', async () => {
    // Two watchers of one key read the same snapshots.
    const streams = new SharedStreams()
    const src = source()
    const a = streams.watch('k', src.open)[Symbol.asyncIterator]()
    const b = streams.watch('k', src.open)[Symbol.asyncIterator]()
    const first = Promise.all([a.next(), b.next()])
    src.push(1)
    expect((await first).map((r) => r.value)).toEqual([1, 1])

    // A late watcher starts from the latest snapshot.
    src.push(2)
    expect((await a.next()).value).toBe(2)
    const c = streams.watch('k', src.open)[Symbol.asyncIterator]()
    expect((await c.next()).value).toBe(2)
    expect((await b.next()).value).toBe(2)
    expect(src.opens).toBe(1)

    // A watcher of another key opens its own source.
    const other = source()
    const d = streams.watch('other', other.open)[Symbol.asyncIterator]()
    const next = d.next()
    other.push(9)
    expect((await next).value).toBe(9)
    expect(src.opens).toBe(1)
    await Promise.all([a.return?.(), b.return?.(), c.return?.(), d.return?.()])
  })

  it('aborts the source with its last watcher and reopens on demand', async () => {
    // Two watchers read the first snapshot.
    const streams = new SharedStreams()
    const src = source()
    const first = new AbortController()
    const a = streams.watch('k', src.open, first.signal)[Symbol.asyncIterator]()
    const b = streams.watch('k', src.open)[Symbol.asyncIterator]()
    const pending = Promise.all([a.next(), b.next()])
    src.push(1)
    await pending

    // The source outlives the first watcher's exit.
    const aborted = a.next()
    first.abort()
    await expect(aborted).rejects.toThrow()
    expect(src.aborts).toBe(0)

    // The last watcher's exit aborts it, and the next watch reopens it.
    await b.return?.()
    expect(src.aborts).toBe(1)
    const c = streams.watch('k', src.open)[Symbol.asyncIterator]()
    const reopened = c.next()
    src.push(2)
    expect((await reopened).value).toBe(2)
    expect(src.opens).toBe(2)
    await c.return?.()
  })

  it('reports a source error to every watcher', async () => {
    // Two watchers read the first snapshot.
    const streams = new SharedStreams()
    const src = source()
    const a = streams.watch('k', src.open)[Symbol.asyncIterator]()
    const b = streams.watch('k', src.open)[Symbol.asyncIterator]()
    const pending = Promise.all([a.next(), b.next()])
    src.push(1)
    await pending

    // Both watchers see the source fail.
    const failed = [a.next(), b.next()]
    src.fail(new Error('boom'))
    for (const next of failed) {
      await expect(next).rejects.toThrow('boom')
    }

    // The next watch opens a fresh source.
    const c = streams.watch('k', src.open)[Symbol.asyncIterator]()
    const reopened = c.next()
    src.push(3)
    expect((await reopened).value).toBe(3)
    expect(src.opens).toBe(2)
    await c.return?.()
  })
})
