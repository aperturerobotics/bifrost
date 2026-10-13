import { describe, expect, it, vi } from 'vitest'

import { Actions, Model, TabNode } from '@aptre/flex-layout'
import type { LayoutModel } from '@s4wave/sdk/layout/layout.pb.js'
import type { LayoutHost } from '@s4wave/sdk/layout/layout_srpc.pb.js'

import { BaseLayout } from './BaseLayout.js'

function emptyLayoutStream(): AsyncIterable<LayoutModel> {
  return {
    [Symbol.asyncIterator]() {
      return {
        next: () => Promise.resolve({ value: undefined as never, done: true }),
      }
    },
  }
}

function buildLayoutHost(replaceTab: LayoutHost['ReplaceTab']): LayoutHost {
  return {
    WatchLayoutModel: () => emptyLayoutStream() as never,
    NavigateTab: () => Promise.resolve({}),
    ReplaceTab: replaceTab,
    AddTab: (request) => Promise.resolve({ tabId: request.tab?.id ?? '' }),
  }
}

describe('BaseLayout', () => {
  it('binds replacement requests to the current tab id', async () => {
    const replaceTab = vi.fn<LayoutHost['ReplaceTab']>(() =>
      Promise.resolve({}),
    )
    const layoutHost = buildLayoutHost(replaceTab)
    const layout = new BaseLayout({
      layoutHost,
      renderTab: () => null,
    })
    const replacementData = new TextEncoder().encode('replacement-payload')

    await expect(
      layout.replaceTab('current-tab', {
        tab: {
          id: 'replacement-tab',
          name: 'Replacement',
          helpText: 'replacement-help',
          enableClose: true,
          data: replacementData,
        },
      }),
    ).resolves.toEqual({})

    expect(replaceTab).toHaveBeenCalledTimes(1)
    const request = replaceTab.mock.calls[0]?.[0]
    if (!request) throw new Error('expected ReplaceTab request')
    expect(request).toMatchObject({
      tabId: 'current-tab',
      tab: {
        id: 'replacement-tab',
        name: 'Replacement',
        helpText: 'replacement-help',
        enableClose: true,
      },
    })
    expect(
      new TextDecoder().decode(request.tab?.data ?? new Uint8Array()),
    ).toBe('replacement-payload')
  })

  it('ignores replacement requests without a tab payload', async () => {
    const replaceTab = vi.fn<LayoutHost['ReplaceTab']>(() =>
      Promise.resolve({}),
    )
    const layoutHost = buildLayoutHost(replaceTab)
    const layout = new BaseLayout({
      layoutHost,
      renderTab: () => null,
    })

    await expect(layout.replaceTab('current-tab', {})).resolves.toEqual({})
    expect(replaceTab).not.toHaveBeenCalled()
  })

  it('names a tab after its object until the user renames it', () => {
    const layout = new BaseLayout({
      layoutHost: buildLayoutHost(() => Promise.resolve({})),
      renderTab: () => null,
    })
    const model = Model.fromJson({
      global: {},
      layout: {
        type: 'row',
        children: [
          {
            type: 'tabset',
            children: [{ type: 'tab', id: 'vm-tab', name: 'vm-1' }],
          },
        ],
      },
    })
    layout.state = { ...layout.state, model }
    const tabName = () => (model.getNodeById('vm-tab') as TabNode).getName()

    layout.nameTab('vm-tab', 'vm/vm-1', 'Build box')
    expect(tabName()).toBe('Build box')
    layout.nameTab('vm-tab', 'vm/vm-1', 'Release box')
    expect(tabName()).toBe('Release box')

    model.doAction(Actions.renameTab('vm-tab', 'Mine'))
    layout.nameTab('vm-tab', 'vm/vm-1', 'Build box')
    expect(tabName()).toBe('Mine')
  })
})
