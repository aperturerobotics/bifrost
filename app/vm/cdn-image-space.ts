import type { Cdn } from '@s4wave/sdk/cdn/cdn.js'
import { Root } from '@s4wave/sdk/root/root.js'
import type { Space } from '@s4wave/sdk/space/space.js'

// CdnImageSpace is the mounted Space of a CDN's image catalog.
export interface CdnImageSpace {
  cdn: Cdn
  space: Space
  [Symbol.dispose](): void
}

// mountCdnImageSpace mounts the catalog Space of the CDN cdnId.
export async function mountCdnImageSpace(
  root: Root,
  cdnId: string,
  signal: AbortSignal,
): Promise<CdnImageSpace> {
  // Keep a separate parent reference alive while child resources are created.
  const ownedRoot = root.resourceRef.createResource(
    root.resourceRef.resourceId,
    Root,
  )
  let cdn: Cdn | undefined
  let space: Space | undefined
  const release = () => {
    space?.[Symbol.dispose]()
    cdn?.[Symbol.dispose]()
    ownedRoot[Symbol.dispose]()
  }
  try {
    cdn = (await ownedRoot.getCdn(cdnId, signal)).cdn
    space = await cdn.mountCdnSpace(signal)
    return { cdn, space, [Symbol.dispose]: release }
  } catch (err) {
    release()
    throw err
  }
}
