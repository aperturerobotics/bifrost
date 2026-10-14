import { Engine } from './engine.js'
import type { IWorldState } from './world-state.js'

/** WorldQuery reads one bounded answer from an immutable World snapshot. */
export type WorldQuery<T> = (
  world: IWorldState,
  signal?: AbortSignal,
) => Promise<T>

/** WorldQuerySnapshot binds a query result to the World revision it read. */
export interface WorldQuerySnapshot<T> {
  readonly value: T
  readonly seqno: bigint
}

/**
 * watchWorldQuery runs a query on the World head, then again after each commit
 * that changes what the previous run read. Every run reads one read snapshot,
 * so its result is consistent, and the server records the snapshot's reads to
 * decide when to run again. The iterator retains its own Engine reference and
 * releases every snapshot before yielding.
 */
export async function* watchWorldQuery<T>(
  owner: Engine,
  query: WorldQuery<T>,
  signal?: AbortSignal,
): AsyncIterable<WorldQuerySnapshot<T>> {
  using engine = new Engine(owner.resourceRef.createRef(owner.id))
  for await (const world of engine.watchWorldState(signal)) {
    // Pin the result to the snapshot's revision, then release the snapshot.
    let snapshot: WorldQuerySnapshot<T>
    try {
      const { seqno } = await world.getSeqno(signal)
      snapshot = { seqno, value: await query(world, signal) }
    } finally {
      world[Symbol.dispose]()
    }

    signal?.throwIfAborted()
    yield snapshot
  }
}
