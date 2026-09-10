import { describe, expect, it, vi } from 'vitest'
import superjson from 'superjson'
import type { Root } from '@s4wave/sdk/root'
import { createAppEnvironment } from './environment.js'
import { persistAppEnvironment } from './persistence.js'

describe('persistent app environment', () => {
  it('flushes edits and restores a named view without sharing another view', async () => {
    const saved = new Map<string, string>()
    const release = vi.fn()
    let finishWrite: (() => void) | undefined
    let firstWrite = true
    const root = {
      accessStateAtom: async ({ storeId }: { storeId: string }) => ({
        getState: async () => ({ stateJson: saved.get(storeId) ?? '{}' }),
        setState: async (value: string) => {
          if (firstWrite) {
            firstWrite = false
            await new Promise<void>((resolve) => {
              finishWrite = resolve
            })
          }
          saved.set(storeId, value)
        },
        release,
      }),
    } as unknown as Root
    const failure = vi.fn()
    const signal = new AbortController().signal
    const left = createAppEnvironment('left')
    const close = await persistAppEnvironment(root, left, signal, failure)
    left.storage.setItem('tabs', 'first')
    left.storage.setItem('tabs', 'latest')
    left.rootAtom.set({ selected: 'canvas' })
    left.navigation.setAppPath('/u/1/space', { panel: 'files' })
    await Promise.resolve()
    finishWrite!()
    await close()
    expect(failure).not.toHaveBeenCalled()
    expect(release).toHaveBeenCalledOnce()

    const reopened = createAppEnvironment('left')
    const closeReopened = await persistAppEnvironment(
      root,
      reopened,
      signal,
      failure,
    )
    expect(reopened.storage.getItem('tabs')).toBe('latest')
    expect(reopened.rootAtom.get()).toEqual({ selected: 'canvas' })
    expect(reopened.navigation.getAppNavigation()).toEqual({
      path: '/u/1/space',
      params: { panel: 'files' },
    })
    const right = createAppEnvironment('right')
    const closeRight = await persistAppEnvironment(root, right, signal, failure)
    expect(right.storage.length).toBe(0)
    expect(right.navigation.getAppPath()).toBe('/')
    await closeReopened()
    await closeRight()
    expect(superjson.parse(saved.get('app-environment/left')!)).toMatchObject({
      storage: { tabs: 'latest' },
    })
  })
})
