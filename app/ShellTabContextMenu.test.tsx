import { useState } from 'react'
import { afterEach, describe, expect, it, vi, type Mock } from 'vitest'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import {
  ShellTabContextMenu,
  type ShellTabContextMenuProps,
  type ShellTabContextMenuState,
} from './ShellTabContextMenu.js'

type ActionProp = Exclude<
  keyof ShellTabContextMenuProps,
  'state' | 'canCloseTabs' | 'onClose'
>

// OpenMenu holds the menu state the way a layout does: open at tab-7, closed
// on dismissal or selection.
function OpenMenu(props: Omit<ShellTabContextMenuProps, 'state' | 'onClose'>) {
  const [state, setState] = useState<ShellTabContextMenuState | null>({
    tabId: 'tab-7',
    x: 10,
    y: 20,
  })
  return (
    <ShellTabContextMenu
      {...props}
      state={state}
      onClose={() => setState(null)}
    />
  )
}

// renderMenu opens the shell tab menu with a spy for each action.
function renderMenu(canCloseTabs: boolean) {
  const actions: Record<ActionProp, Mock<(tabId: string) => void>> = {
    onNewTab: vi.fn(),
    onRenameTab: vi.fn(),
    onDuplicateTab: vi.fn(),
    onPopoutTab: vi.fn(),
    onCloseOtherTabs: vi.fn(),
    onCloseTab: vi.fn(),
  }
  render(<OpenMenu canCloseTabs={canCloseTabs} {...actions} />)
  return actions
}

describe('ShellTabContextMenu', () => {
  afterEach(() => {
    cleanup()
  })

  it.each<[string, ActionProp]>([
    ['New Tab', 'onNewTab'],
    ['Rename Tab', 'onRenameTab'],
    ['Duplicate Tab', 'onDuplicateTab'],
    ['Open in New Tab', 'onPopoutTab'],
    ['Close Other Tabs', 'onCloseOtherTabs'],
    ['Close Tab', 'onCloseTab'],
  ])('routes %s to %s for the clicked tab', async (label, prop) => {
    const actions = renderMenu(true)

    await userEvent.click(screen.getByRole('menuitem', { name: label }))

    await waitFor(() => expect(actions[prop]).toHaveBeenCalledWith('tab-7'))
    for (const [other, spy] of Object.entries(actions)) {
      if (other !== prop) expect(spy).not.toHaveBeenCalled()
    }
  })

  it('disables close actions when only one tab remains', () => {
    renderMenu(false)

    for (const label of ['Close Other Tabs', 'Close Tab']) {
      expect(
        screen
          .getByRole('menuitem', { name: label })
          .getAttribute('aria-disabled'),
      ).toBe('true')
    }
  })
})
