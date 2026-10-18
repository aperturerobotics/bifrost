import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import {
  FlexTabContextMenu,
  type FlexTabContextMenuEntry,
  type FlexTabContextMenuState,
} from './FlexTabContextMenu.js'

// OpenMenu holds the menu state the way a layout does: open at a tab, closed
// on dismissal.
function OpenMenu({ items }: { items: FlexTabContextMenuEntry[] }) {
  const [state, setState] = useState<FlexTabContextMenuState | null>({
    tabId: 'tab-7',
    x: 12,
    y: 34,
  })
  return (
    <FlexTabContextMenu
      state={state}
      items={items}
      onClose={() => setState(null)}
    />
  )
}

describe('FlexTabContextMenu', () => {
  afterEach(() => {
    cleanup()
  })

  it('runs the chosen action for the clicked tab after the menu closes', async () => {
    const user = userEvent.setup()
    let menuOpenAtRename: boolean | null = null
    const rename = vi.fn(() => {
      menuOpenAtRename = screen.queryByRole('menu') !== null
    })
    const close = vi.fn()

    render(
      <OpenMenu
        items={[
          { id: 'rename', label: 'Rename Tab', onSelect: rename },
          { id: 'sep', type: 'separator' },
          {
            id: 'close',
            label: 'Close Tab',
            variant: 'destructive',
            onSelect: close,
          },
        ]}
      />,
    )

    expect(screen.getAllByRole('separator')).toHaveLength(1)
    await user.click(screen.getByRole('menuitem', { name: 'Rename Tab' }))

    await waitFor(() => expect(rename).toHaveBeenCalledWith('tab-7'))
    expect(menuOpenAtRename).toBe(false)
    expect(close).not.toHaveBeenCalled()
  })

  it('runs nothing when dismissed or when a disabled item is clicked', async () => {
    const user = userEvent.setup()
    const disabled = vi.fn()

    render(
      <OpenMenu
        items={[
          {
            id: 'disabled',
            label: 'Close Tab',
            disabled: true,
            onSelect: disabled,
          },
        ]}
      />,
    )

    await user.click(screen.getByRole('menuitem', { name: 'Close Tab' }))
    expect(screen.getByRole('menu')).toBeTruthy()
    await user.keyboard('{Escape}')

    await waitFor(() => expect(screen.queryByRole('menu')).toBeNull())
    expect(disabled).not.toHaveBeenCalled()
  })
})
