import { describe, expect, it, vi } from 'vitest'
import type { V86fsMessage } from '@go/github.com/s4wave/spacewave/db/unixfs/v86fs/v86fs.pb.js'
import { createV86fsSrpcAdapter, type V86fsNotifier } from './v86fs-bridge.js'

const h = vi.hoisted(() => ({
  responses: [] as V86fsMessage[],
}))

vi.mock(
  '@go/github.com/s4wave/spacewave/db/unixfs/v86fs/v86fs_srpc.pb.js',
  () => ({
    V86fsServiceClient: class {
      async *RelayV86fs() {
        for (const msg of h.responses) {
          yield msg
        }
      }
    },
  }),
)

function fakeNotifier(): V86fsNotifier & Record<string, unknown> {
  return {
    invalidate_inode_range: vi.fn(),
    invalidate_dir: vi.fn(),
    mount_notify: vi.fn(),
    umount_notify: vi.fn(),
  }
}

describe('v86fs bridge notifications', () => {
  it('holds server notifications until a notifier attaches', async () => {
    h.responses = [
      {
        tag: 0,
        body: {
          case: 'mountNotify',
          value: { name: 'home', mountPath: '/home' },
        },
      },
      {
        tag: 0,
        body: { case: 'invalidateDir', value: { dirId: 9n } },
      },
    ]
    const bridge = createV86fsSrpcAdapter({} as never)
    await new Promise((resolve) => setTimeout(resolve, 0))

    const notifier = fakeNotifier()
    bridge.attachNotifier(notifier)
    expect(notifier.mount_notify).toHaveBeenCalledWith('home', '/home')
    expect(notifier.invalidate_dir).toHaveBeenCalledWith(9)
    bridge.close()
  })
})
