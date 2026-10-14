import { ObjectRef } from '../../db/bucket/bucket.pb.js'
import { Engine } from '../world/engine.js'
import type { IObjectState } from '../world/object-state.js'

/** ObjectQuery reads a typed projection from one immutable object snapshot. */
export type ObjectQuery<T> = (
  object: IObjectState,
  signal?: AbortSignal,
) => Promise<T>

/** ObjectQuerySnapshot binds a projection, or an absent object, to its World revision. */
export interface ObjectQuerySnapshot<T> {
  readonly value: T | null
  readonly seqno: bigint
}

/**
 * watchObjectQuery projects an existing World object without copying its storage.
 * Each run reads the object from one tracked read snapshot, which the server
 * replaces only after a commit that changes the object. Unchanged roots skip
 * the callback. The iterator independently retains the Engine and releases
 * every snapshot before yielding.
 */
export async function* watchObjectQuery<T>(
  owner: Engine,
  objectKey: string,
  query: ObjectQuery<T>,
  signal?: AbortSignal,
): AsyncIterable<ObjectQuerySnapshot<T>> {
  using engine = new Engine(owner.resourceRef.createRef(owner.id))
  let previous: string | undefined
  for await (const world of engine.watchWorldState(signal)) {
    // Pin the value to the snapshot's revision, then release the snapshot.
    let next: ObjectQuerySnapshot<T> | undefined
    try {
      const { seqno } = await world.getSeqno(signal)
      using object = await world.getObject(objectKey, signal)
      const root = object ? (await object.getRootRef(signal)).rootRef : null
      const signature = !object
        ? 'missing'
        : root
          ? Array.from(ObjectRef.toBinary(root)).join(',')
          : 'empty'
      if (signature !== previous) {
        next = { seqno, value: object ? await query(object, signal) : null }
        previous = signature
      }
    } finally {
      world[Symbol.dispose]()
    }

    signal?.throwIfAborted()
    if (next) yield next
  }
}
