import { describe, expect, it, vi } from 'vitest'
import { createRef } from 'react'
import { render, waitFor } from '@testing-library/react'

import {
  Turnstile,
  TURNSTILE_PROD_SITE_KEY,
  TURNSTILE_TEST_SITE_KEY,
  TURNSTILE_TEST_TOKEN,
  isTurnstileBypassed,
  type TurnstileInstance,
} from './turnstile.js'

describe('turnstile', () => {
  it('does not bypass the production key', () => {
    expect(isTurnstileBypassed(TURNSTILE_PROD_SITE_KEY)).toBe(false)
  })

  it('bypasses widget loading for the test key', async () => {
    const appendChild = vi.spyOn(document.head, 'appendChild')
    const ref = createRef<TurnstileInstance>()
    const { container } = render(
      <Turnstile ref={ref} siteKey={TURNSTILE_TEST_SITE_KEY} />,
    )

    expect(isTurnstileBypassed(TURNSTILE_TEST_SITE_KEY)).toBe(true)
    expect(container.firstChild).toBeNull()
    expect(appendChild).not.toHaveBeenCalled()
    await expect(ref.current?.takeToken()).resolves.toBe(TURNSTILE_TEST_TOKEN)

    appendChild.mockRestore()
  })

  it('hands out each token once and solves again for the next', async () => {
    let issue: ((token: string) => void) | undefined
    const reset = vi.fn<(widgetId: string) => void>()
    window.turnstile = {
      render: (_container, params) => {
        issue = params.callback
        return 'widget-1'
      },
      remove: vi.fn(),
      reset,
    }
    const ref = createRef<TurnstileInstance>()
    render(<Turnstile ref={ref} siteKey={TURNSTILE_PROD_SITE_KEY} />)

    // Load the script and let the widget issue its first token.
    const onload = (window as unknown as Record<string, () => void>)
      .onloadTurnstileCallback
    onload()
    await waitFor(() => expect(issue).toBeDefined())
    issue?.('token-1')
    await expect(ref.current?.takeToken()).resolves.toBe('token-1')
    expect(reset).toHaveBeenCalledWith('widget-1')

    // A retry waits for the next token instead of reusing the spent one.
    const next = ref.current?.takeToken()
    issue?.('token-2')
    await expect(next).resolves.toBe('token-2')
    expect(reset).toHaveBeenCalledTimes(2)
  })
})
