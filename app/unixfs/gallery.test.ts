import { describe, expect, it } from 'vitest'

import type { HandleWatchFilesResponse } from '@s4wave/sdk/unixfs/handle.pb.js'

import {
  type UnixFSGalleryDiscoveryState,
  streamUnixFSGalleryCandidates,
} from './gallery.js'

// FakeHandle serves one path of a fake tree and replays the WatchFiles
// responses of its directory.
class FakeHandle {
  watchedExtensions: string[] = []

  constructor(
    private readonly files: Set<string>,
    private readonly responses: Map<string, HandleWatchFilesResponse[]>,
    readonly path = '',
  ) {}

  [Symbol.dispose](): void {}

  clone(): Promise<FakeHandle> {
    return Promise.resolve(this)
  }

  getFileInfo(): Promise<{ isDir: boolean }> {
    return Promise.resolve({ isDir: !this.files.has(this.path) })
  }

  lookupPath(path: string): Promise<{ handle: FakeHandle }> {
    const handle = new FakeHandle(this.files, this.responses, path)
    lastHandle = handle
    return Promise.resolve({ handle })
  }

  async *watchFiles(
    extensions: string[],
  ): AsyncIterable<HandleWatchFilesResponse> {
    this.watchedExtensions = extensions
    for (const resp of this.responses.get(this.path) ?? []) {
      yield await Promise.resolve(resp)
    }
  }
}

let lastHandle: FakeHandle | undefined

// collectStates runs the gallery stream over a fake tree until its watch ends.
async function collectStates(
  scopePath: string,
  responses: Record<string, HandleWatchFilesResponse[]>,
  files: string[] = [],
): Promise<UnixFSGalleryDiscoveryState[]> {
  const root = new FakeHandle(
    new Set(files),
    new Map(Object.entries(responses)),
  )
  const states: UnixFSGalleryDiscoveryState[] = []
  for await (const state of streamUnixFSGalleryCandidates(
    root as unknown as Parameters<typeof streamUnixFSGalleryCandidates>[0],
    scopePath,
    new AbortController().signal,
  )) {
    states.push(state)
  }
  return states
}

describe('streamUnixFSGalleryCandidates', () => {
  it('maps the first response to the complete index of the scope', async () => {
    const states = await collectStates('/docs', {
      docs: [
        {
          added: ['nested/poster.jpg', 'cover.png'],
          errors: ['locked: permission denied'],
          truncated: true,
        },
      ],
    })

    expect(lastHandle?.watchedExtensions).toContain('.png')
    expect(states).toEqual([
      {
        scopePath: '/docs',
        items: [
          {
            path: '/docs/cover.png',
            name: 'cover.png',
            label: 'cover.png',
            mimeType: 'image/png',
          },
          {
            path: '/docs/nested/poster.jpg',
            name: 'poster.jpg',
            label: 'nested/poster.jpg',
            mimeType: 'image/jpeg',
          },
        ],
        errors: ['locked: permission denied'],
        truncated: true,
      },
    ])
  })

  it('applies the paths each change adds and removes', async () => {
    const states = await collectStates('/docs', {
      docs: [
        { added: ['cover.png', 'b.webp'] },
        { added: ['a.svg'], removed: ['cover.png'] },
      ],
    })

    expect(states.map((state) => state.items.map((item) => item.path))).toEqual(
      [
        ['/docs/b.webp', '/docs/cover.png'],
        ['/docs/a.svg', '/docs/b.webp'],
      ],
    )
  })

  it('indexes the parent directory when the scope path names a file', async () => {
    const states = await collectStates(
      '/docs/cover.png',
      { docs: [{ added: ['cover.png'] }] },
      ['docs/cover.png'],
    )

    expect(states.map((state) => state.scopePath)).toEqual(['/docs'])
    expect(states[0]?.items.map((item) => item.path)).toEqual([
      '/docs/cover.png',
    ])
  })
})
