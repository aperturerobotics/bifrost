import { expect, it } from 'vitest'
import { page } from 'vitest/browser'
import { render } from 'vitest-browser-react'

import { Toaster } from '@s4wave/web/ui/toaster.js'

import { SharedObjectSyncNotice } from './SharedObjectSyncNotice.js'

it('keeps recovery readable without covering the viewport and dismisses on convergence', async () => {
  const screen = await render(
    <>
      <Toaster />
      <SharedObjectSyncNotice health={{ syncRecoveryPeerIds: ['source'] }} />
    </>,
  )
  await expect
    .element(page.getByText('Direct sync needs attention'))
    .toBeVisible()
  for (const width of [1280, 390, 280]) {
    await page.viewport(width, 800)
    const notice = document.querySelector('[data-sonner-toast]')
    expect(notice).not.toBeNull()
    await expect
      .poll(() => notice!.getBoundingClientRect().bottom)
      .toBeLessThanOrEqual(window.innerHeight - 40)
    const bounds = notice!.getBoundingClientRect()
    expect(bounds.left).toBeGreaterThanOrEqual(0)
    expect(bounds.right).toBeLessThanOrEqual(window.innerWidth)
    await expect
      .element(page.getByText(/You can keep using local content\./))
      .toBeVisible()
    await page.screenshot({
      path: `__screenshots__/sync-recovery/${width}.png`,
    })
  }
  await screen.rerender(
    <>
      <Toaster />
      <SharedObjectSyncNotice health={{ syncRecoveryPeerIds: [] }} />
    </>,
  )
  await expect
    .element(page.getByText('Direct sync needs attention'))
    .not.toBeInTheDocument()
})
