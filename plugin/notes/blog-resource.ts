import type { ClientResourceRef } from '@aptre/bldr-sdk/resource/client.js'
import type { MessageStream } from 'starpc'

import { watchObjectQuery } from '@s4wave/sdk/sync/object-query.js'
import { Engine } from '@s4wave/sdk/world/engine.js'
import type { IObjectState } from '@s4wave/sdk/world/object-state.js'

import { readObjectBlockData } from './object-block.js'
import { Blog } from './proto/blog.pb.js'
import type { WatchBlogRequest, WatchBlogResponse } from './sdk/blog.pb.js'
import type { BlogResourceService } from './sdk/blog_srpc.pb.js'

// readBlog decodes the Blog block of a pinned object snapshot.
async function readBlog(
  objectState: IObjectState,
  signal?: AbortSignal,
): Promise<Blog | null> {
  const data = await readObjectBlockData(objectState, signal)
  return data ? Blog.fromBinary(data) : null
}

/** BlogResource serves BlogResourceService for one Blog World object. */
class BlogResource implements BlogResourceService {
  private objectKey: string
  private engineRef: ClientResourceRef | undefined

  constructor(objectKey: string, engineRef: ClientResourceRef | undefined) {
    this.objectKey = objectKey
    this.engineRef = engineRef
  }

  /** WatchBlog streams the Blog block again after each change to its object. */
  async *WatchBlog(
    _request: WatchBlogRequest,
    abortSignal?: AbortSignal,
  ): MessageStream<WatchBlogResponse> {
    if (!this.engineRef) {
      return
    }

    // The query owns its Engine reference independently of this resource.
    try {
      for await (const { value } of watchObjectQuery(
        new Engine(this.engineRef),
        this.objectKey,
        readBlog,
        abortSignal,
      )) {
        if (value) {
          yield { blog: value }
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

export { BlogResource }
