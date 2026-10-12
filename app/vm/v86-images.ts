import type { Cdn } from '@s4wave/sdk/cdn/cdn.js'
import { Root } from '@s4wave/sdk/root/root.js'
import type { Space } from '@s4wave/sdk/space/space.js'
import { V86Image } from '@s4wave/sdk/vm/v86.pb.js'
import { V86ImageTypeID } from '@s4wave/sdk/vm/v86image.js'
import { listObjectsWithType } from '@s4wave/sdk/world/types/types.js'
import type { IWorldState } from '@s4wave/sdk/world/world-state.js'

import {
  compareV86ImageNewestFirst,
  isDefaultV86Image,
} from './v86-wizard-config.js'

// V86ImageEntry is a V86Image object and its decoded metadata. metadataError
// is set when the metadata could not be decoded; the image then carries a
// placeholder.
export interface V86ImageEntry {
  objectKey: string
  image: V86Image
  metadataError?: string
}

// readV86Image reads the V86Image at objectKey, or returns undefined when no
// image metadata is stored there.
export async function readV86Image(
  world: IWorldState,
  objectKey: string,
  signal: AbortSignal,
): Promise<V86ImageEntry | undefined> {
  using obj = await world.getObject(objectKey, signal)
  if (!obj) return undefined
  using cursor = await obj.accessWorldState(undefined, signal)
  const resp = await cursor.unmarshal({ blockType: V86ImageTypeID }, signal)
  if (!resp.found || !resp.data?.length) return undefined
  try {
    return { objectKey, image: V86Image.fromBinary(resp.data) }
  } catch (err) {
    return {
      objectKey,
      image: {
        name: objectKey,
        platform: 'v86',
        description: 'Metadata could not be decoded.',
        tags: [],
      },
      metadataError:
        err instanceof Error ? err.message : 'metadata decode failed',
    }
  }
}

// listV86Images reads every V86Image in world, newest first.
export async function listV86Images(
  world: IWorldState,
  signal: AbortSignal,
): Promise<V86ImageEntry[]> {
  const keys = await listObjectsWithType(world, V86ImageTypeID, signal)
  const entries = await Promise.all(
    keys.map((key) => readV86Image(world, key, signal)),
  )
  return entries
    .filter((entry): entry is V86ImageEntry => !!entry)
    .sort(compareV86ImageNewestFirst)
}

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

// listCdnV86Images reads every V86Image published in a CDN catalog Space.
export async function listCdnV86Images(
  space: Space,
  signal: AbortSignal,
): Promise<V86ImageEntry[]> {
  using world = await space.accessWorldState(false, signal)
  return await listV86Images(world, signal)
}

// loadCdnV86Image reads the catalog image at objectKey from the CDN cdnId, or
// the catalog's default image when objectKey is empty.
export async function loadCdnV86Image(
  root: Root,
  cdnId: string,
  objectKey: string,
  signal: AbortSignal,
): Promise<V86ImageEntry | undefined> {
  using catalog = await mountCdnImageSpace(root, cdnId, signal)
  if (!objectKey) {
    const entries = await listCdnV86Images(catalog.space, signal)
    return entries.find((entry) => isDefaultV86Image(entry.image))
  }
  using world = await catalog.space.accessWorldState(false, signal)
  return await readV86Image(world, objectKey, signal)
}

// formatV86ImageLabel names an image for display, with its version.
export function formatV86ImageLabel(image: V86Image): string {
  const name = image.name || image.distro || 'VM image'
  return image.version ? `${name} (${image.version})` : name
}

// V86CatalogErrorCopy is the user-facing description of a catalog failure.
export interface V86CatalogErrorCopy {
  title: string
  detail: string
  // unpublished is set when the catalog is empty rather than unreachable, so
  // retrying cannot help.
  unpublished: boolean
}

// getV86CatalogErrorCopy describes a catalog load failure without exposing
// backend detail.
export function getV86CatalogErrorCopy(error: unknown): V86CatalogErrorCopy {
  const message =
    error instanceof Error
      ? error.message
      : typeof error === 'string'
        ? error
        : ''
  const unpublished =
    /published head|not published|no published|empty catalog/i.test(message)
  return unpublished
    ? {
        title: 'No VM images are published yet',
        detail: 'This image catalog has no published images yet.',
        unpublished: true,
      }
    : {
        title: 'Image catalog unavailable',
        detail: 'The VM image catalog could not be loaded. Try again.',
        unpublished: false,
      }
}
