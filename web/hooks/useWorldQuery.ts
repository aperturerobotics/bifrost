import type { DependencyList } from 'react'

import type { Resource } from '@aptre/bldr-sdk/hooks/useResource.js'
import {
  useStreamingResource,
  type UseStreamingResourceOptions,
} from '@aptre/bldr-sdk/hooks/useStreamingResource.js'
import {
  watchWorldQuery,
  type WorldQuery,
} from '@s4wave/sdk/world/world-query.js'
import type { IWorldState } from '@s4wave/sdk/world/world-state.js'

/**
 * useWorldQuery runs a bounded World query and re-runs it on each World
 * revision. A live World is watched through its Engine; a World without one,
 * such as a historical snapshot, is read once. deps lists the values the
 * query closes over; changing one restarts the watch. A watch that fails
 * after its first result reconnects; options can override that or add an
 * equals check that keeps an unchanged result.
 */
export function useWorldQuery<T>(
  world: Resource<IWorldState>,
  query: WorldQuery<T>,
  deps: DependencyList,
  options?: UseStreamingResourceOptions<T>,
): Resource<T> {
  return useStreamingResource(
    world,
    async function* (state, signal) {
      const engine = state.getEngine?.()
      if (!engine) {
        yield await query(state, signal)
        return
      }
      for await (const snapshot of watchWorldQuery(engine, query, signal)) {
        yield snapshot.value
      }
    },
    deps,
    { reconnect: true, ...options },
  )
}
