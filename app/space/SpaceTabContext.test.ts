import { describe, expect, it } from 'vitest'

import { createObjectLayoutAddTabRequest } from '@s4wave/sdk/layout/world/object-layout.js'
import { ObjectLayoutTab } from '@s4wave/sdk/layout/world/world.pb.js'

import { spaceAddTabRequest } from './SpaceTabContext.js'

// tabPath reads the path of the object layout tab a request adds.
function tabPath(data: Uint8Array | undefined): string {
  return ObjectLayoutTab.fromBinary(data ?? new Uint8Array()).path ?? ''
}

describe('spaceAddTabRequest', () => {
  const parentPaths = ['/u/1/so/space-1']

  it('opens a world object at its route with the viewer path under it', () => {
    const request = createObjectLayoutAddTabRequest({
      id: 'w',
      name: 'Window',
      objectKey: 'vm/demo',
      objectType: 'vm/wasivm',
      path: 'window/3/7',
    })
    const out = spaceAddTabRequest(request, parentPaths)
    expect(tabPath(out.tab?.data)).toBe(
      '/u/1/so/space-1/-/vm/demo/-/window/3/7',
    )
    expect(out.tab?.name).toBe('Window')
  })

  it('opens a world object without a viewer path at its route', () => {
    const request = createObjectLayoutAddTabRequest({
      id: 'o',
      name: 'Object',
      objectKey: 'vm/demo',
    })
    expect(tabPath(spaceAddTabRequest(request, parentPaths).tab?.data)).toBe(
      '/u/1/so/space-1/-/vm/demo',
    )
  })

  it('opens a UnixFS file at its UnixFS route with the file path', () => {
    const request = createObjectLayoutAddTabRequest({
      id: 'f',
      name: 'a.txt',
      objectInfo: {
        info: {
          case: 'unixfsObjectInfo',
          value: { unixfsId: 'files/drive', path: '/docs/a.txt' },
        },
      },
    })
    const out = spaceAddTabRequest(request, parentPaths)
    expect(tabPath(out.tab?.data)).toBe(
      '/u/1/so/space-1/-/files/drive/-/docs/a.txt',
    )
    expect(out.tab?.name).toBe('a.txt')
  })

  it('opens the root of a UnixFS at its route', () => {
    const request = createObjectLayoutAddTabRequest({
      id: 'r',
      name: 'Drive',
      objectInfo: {
        info: {
          case: 'unixfsObjectInfo',
          value: { unixfsId: 'files/drive', path: '/' },
        },
      },
    })
    expect(tabPath(spaceAddTabRequest(request, parentPaths).tab?.data)).toBe(
      '/u/1/so/space-1/-/files/drive',
    )
  })

  it('passes a request without an object through', () => {
    const request = {
      tab: {
        id: 'p',
        name: 'Path',
        data: ObjectLayoutTab.toBinary({ path: '/u/1/settings' }),
      },
    }
    expect(spaceAddTabRequest(request, parentPaths)).toBe(request)
  })
})
