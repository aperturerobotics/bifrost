import { afterEach, describe, expect, it } from 'vitest'
import { page } from 'vitest/browser'

import spacewaveIcon from '@s4wave/web/images/spacewave-icon.png'

import { buildStartupShell } from './startup-shell.js'
import {
  writeBrowserBootStatus,
  resetBrowserStartupMarksForTest,
} from './boot-status.js'

let host: HTMLDivElement | undefined

afterEach(() => {
  host?.remove()
  host = undefined
  globalThis.__swBootStatus = undefined
  resetBrowserStartupMarksForTest()
})

describe('initial startup HTML', () => {
  it('updates actual startup phases and preserves usable error recovery', async () => {
    await page.viewport(1440, 900)
    host = document.createElement('div')
    host.style.cssText = 'position:fixed;inset:0;display:flex'
    host.innerHTML = buildStartupShell(spacewaveIcon)
    document.body.append(host)
    const shell = host.querySelector<HTMLElement>('#sw-loading')!
    shell.style.display = 'block'

    writeBrowserBootStatus({
      phase: 'runtime',
      state: 'loading',
      detail: 'Runtime channel opened.',
      progress: 0.6,
    })
    await expect
      .element(page.getByRole('heading', { name: 'Preparing Spacewave' }))
      .toBeVisible()
    expect(host.querySelector('details')).toBeNull()
    expect(host.querySelector('[role="progressbar"]')).toBeNull()
    expect(host.querySelector('[data-sw-boot-status]')?.textContent).toContain(
      'Connecting the Spacewave runtime.',
    )
    await page.screenshot({
      path: '__screenshots__/browser-startup/initial-html-desktop.png',
    })

    writeBrowserBootStatus({
      phase: 'runtime-error',
      state: 'error',
      detail: 'Network unavailable.',
    })
    await expect
      .element(page.getByRole('heading', { name: 'Unable to open Spacewave' }))
      .toBeVisible()
    await expect.element(page.getByRole('alert')).toBeVisible()
    await expect
      .element(page.getByRole('button', { name: 'Retry' }))
      .toBeVisible()
    await expect
      .element(page.getByRole('button', { name: 'Back', exact: true }))
      .toBeVisible()
    expect(
      host
        .querySelector('[data-sw-boot-phase=runtime]')
        ?.getAttribute('data-sw-boot-phase-state'),
    ).toBe('error')
  })
})
