import { useCallback, useState } from 'react'

import type { IObjectState } from '@s4wave/sdk/world/object-state.js'
import type { IWorldState } from '@s4wave/sdk/world/world-state.js'

// readVmBlock decodes the block of a VM object, or returns null when it has
// none.
async function readVmBlock<T>(
  objectState: IObjectState,
  decode: (data: Uint8Array) => T,
  signal: AbortSignal,
): Promise<T | null> {
  using cursor = await objectState.accessWorldState(undefined, signal)
  const blockResp = await cursor.getBlock({}, signal)
  if (!blockResp.found || !blockResp.data) return null
  return decode(blockResp.data)
}

// watchVmObject yields the decoded block of the VM at objectKey and again at
// each new revision of the object.
export async function* watchVmObject<T>(
  world: IWorldState,
  objectKey: string,
  decode: (data: Uint8Array) => T,
  signal: AbortSignal,
): AsyncIterable<T> {
  if (!objectKey) return
  const objectState = await world.getObject(objectKey, signal)
  if (!objectState) return
  using object = objectState
  let revision = (await object.getRootRef(signal)).rev ?? 0n
  for (;;) {
    const vm = await readVmBlock(object, decode, signal)
    if (!vm) return
    yield vm
    revision =
      (await object.waitRev(revision + 1n, false, signal)).rev ?? revision
  }
}

// useVmRequest runs one action at a time from a view and keeps the failure
// of the last one until the next starts.
export function useVmRequest(): [
  string,
  (action: () => Promise<void>) => void,
] {
  const [error, setError] = useState('')
  const request = useCallback((action: () => Promise<void>) => {
    setError('')
    action().catch((err: unknown) => {
      setError(err instanceof Error ? err.message : String(err))
    })
  }, [])
  return [error, request]
}
