import {
  Client as ResourceClient,
  type ClientResourceRef,
} from '@aptre/bldr-sdk/resource/client.js'
import { Resource } from '@aptre/bldr-sdk/resource/resource.js'
import { ResourceServiceClient } from '@aptre/bldr-sdk/resource/resource_srpc.pb.js'
import type { Engine } from '../world/engine.js'

export type AppStorage =
  | { storageId: string }
  | { world: Engine; objectPrefix: string }
  | { ephemeral: true }

// AppRuntime owns an attached runtime connection and its storage capability.
// Release frontend resources before releasing this attachment.
export class AppRuntime extends Resource {
  readonly resourceClient: ResourceClient
  private readonly controller = new AbortController()

  constructor(
    ref: ClientResourceRef,
    private readonly worldRef?: ClientResourceRef,
    readonly httpPathPrefix = '',
  ) {
    super(ref)
    this.resourceClient = new ResourceClient(
      new ResourceServiceClient(ref.client),
      this.controller.signal,
    )
  }

  override release() {
    if (this.released) return
    this.resourceClient.dispose()
    this.controller.abort()
    super.release()
    this.worldRef?.release()
  }
}
