import type { ClientResourceRef } from '@aptre/bldr-sdk/resource/client.js'
import type { MessageStream } from 'starpc'

import { watchObjectQuery } from '@s4wave/sdk/sync/object-query.js'
import { Engine } from '@s4wave/sdk/world/engine.js'
import type { IObjectState } from '@s4wave/sdk/world/object-state.js'

import { readObjectBlockData } from './object-block.js'
import { Documentation } from './proto/docs.pb.js'
import type { WatchDocsRequest, WatchDocsResponse } from './sdk/docs.pb.js'
import type { DocsResourceService } from './sdk/docs_srpc.pb.js'

// readDocs decodes the Documentation block of a pinned object snapshot.
async function readDocs(
  objectState: IObjectState,
  signal?: AbortSignal,
): Promise<Documentation | null> {
  const data = await readObjectBlockData(objectState, signal)
  return data ? Documentation.fromBinary(data) : null
}

/** DocsResource serves DocsResourceService for one Documentation World object. */
class DocsResource implements DocsResourceService {
  private objectKey: string
  private engineRef: ClientResourceRef | undefined

  constructor(objectKey: string, engineRef: ClientResourceRef | undefined) {
    this.objectKey = objectKey
    this.engineRef = engineRef
  }

  /** WatchDocs streams the Documentation block again after each change to its object. */
  async *WatchDocs(
    _request: WatchDocsRequest,
    abortSignal?: AbortSignal,
  ): MessageStream<WatchDocsResponse> {
    if (!this.engineRef) {
      return
    }

    // The query owns its Engine reference independently of this resource.
    try {
      for await (const { value } of watchObjectQuery(
        new Engine(this.engineRef),
        this.objectKey,
        readDocs,
        abortSignal,
      )) {
        if (value) {
          yield { documentation: value }
        }
      }
    } catch (err) {
      if (abortSignal?.aborted) {
        return
      }
      throw err
    }
  }

  /** dispose releases the engine ref if still held. */
  dispose(): void {
    this.engineRef?.release()
    this.engineRef = undefined
  }
}

export { DocsResource }
