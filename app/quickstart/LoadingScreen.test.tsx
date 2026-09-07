import { afterEach, describe, expect, it } from 'vitest'
import { cleanup, render, screen } from '@testing-library/react'

import { LoadingScreen } from './LoadingScreen.js'

afterEach(() => {
  cleanup()
})

describe('quickstart LoadingScreen', () => {
  it('shows default quickstart progress before the first setup event', () => {
    render(<LoadingScreen quickstartId="drive" />)

    expect(
      screen.getByRole('heading', { name: 'Creating your space' }),
    ).toBeDefined()
    expect(screen.getByText('Local session')).toBeDefined()
    expect(screen.getByText('Open workspace')).toBeDefined()
    expect(screen.getByText('Add starter content')).toBeDefined()
    expect(screen.queryByText('Frame-Ready')).toBeNull()
    expect(screen.queryByText('Content-Ready')).toBeNull()
  })

  it('shows specific setup progress for the active phase', () => {
    render(
      <LoadingScreen
        quickstartId="drive"
        progress={{
          step: 'content',
          stepIndex: 4,
          stepCount: 4,
          detail: 'Seeding My Drive content',
        }}
      />,
    )

    expect(
      screen
        .getByText('Add starter content')
        .parentElement?.getAttribute('aria-current'),
    ).toBe('step')
    expect(screen.queryByText('88%')).toBeNull()
    expect(screen.queryByRole('progressbar')).toBeNull()
  })

  it('keeps local-only quickstart progress scoped to session setup', () => {
    render(<LoadingScreen quickstartId="local" />)

    expect(screen.getByText('Local session')).toBeDefined()
    expect(screen.queryByText('New Space')).toBeNull()
  })
})
