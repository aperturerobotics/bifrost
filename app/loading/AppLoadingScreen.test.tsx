import { afterEach, describe, expect, it, vi } from 'vitest'
import { act, cleanup, render, screen } from '@testing-library/react'
import {
  advanceBootDownload,
  beginBootDownload,
  completeBootDownload,
  failBootDownload,
  resetBootDownloadsForTest,
} from '@aptre/bldr'

import { AppLoadingScreen } from './AppLoadingScreen.js'
import type { BrowserStartupProjection } from './status/browser-startup-model.js'

const mockProjection = vi.hoisted<{
  initial: BrowserStartupProjection
  current: BrowserStartupProjection
}>(() => {
  const initial: BrowserStartupProjection = {
    view: {
      state: 'loading',
      title: 'Starting the Spacewave runtime',
      detail: 'Runtime initialization: Connecting the Spacewave runtime.',
      progress: 0.58,
    },
    phase: {
      id: 'runtime',
      label: 'Runtime',
    },
    phases: [
      { id: 'prepare', label: 'Prepare', state: 'complete' },
      { id: 'connect', label: 'Connect', state: 'complete' },
      { id: 'runtime', label: 'Runtime', state: 'current' },
      { id: 'frame', label: 'App', state: 'pending' },
      { id: 'done', label: 'Done', state: 'pending' },
    ],
    evidence: {
      status: {
        phase: 'runtime',
        detail: 'Connecting runtime...',
        state: 'loading',
        progress: 0.58,
      },
      marks: [],
      runtime: {
        startup: { phase: 'runtime' },
        document: { state: 'unknown' },
        runtimeClient: { state: 'opening' },
        serviceWorker: { state: 'unknown' },
        pluginGeneration: { state: 'idle' },
        frame: { state: 'idle' },
        warmProjection: {
          state: 'cold',
          connection: false,
          neutralFrame: false,
          finalReveal: false,
        },
      },
    },
  }
  return {
    initial,
    current: structuredClone(initial),
  }
})

vi.mock('@s4wave/app/loading/status/browser-startup.js', () => ({
  useBrowserStartupProjection: () => mockProjection.current,
}))

afterEach(() => {
  cleanup()
  resetBootDownloadsForTest()
  mockProjection.current = structuredClone(mockProjection.initial)
  vi.useRealTimers()
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

describe('AppLoadingScreen', () => {
  it('shows projected phases without synthetic percentages', () => {
    render(<AppLoadingScreen />)
    expect(
      screen.getByRole('heading', {
        name: 'Starting the Spacewave runtime',
      }),
    ).toBeDefined()
    expect(screen.getByLabelText('Startup phases')).toBeDefined()
    expect(
      screen.getByText('Runtime').closest('li')?.getAttribute('aria-current'),
    ).toBe('step')
    expect(screen.queryByRole('progressbar')).toBeNull()
    expect(screen.queryByText('58%')).toBeNull()
    expect(screen.getByRole('button', { name: 'Back to home' })).toBeDefined()
  })

  it('advances measured download progress and exposes failed downloads', () => {
    beginBootDownload('app', 'Application', 100)
    advanceBootDownload('app', 10, 100)
    beginBootDownload('plugin', 'Plugin', 200)
    advanceBootDownload('plugin', 50, 200)
    beginBootDownload('styles', 'Styles', 10)

    render(<AppLoadingScreen />)
    expect(screen.getByRole('progressbar').getAttribute('aria-valuenow')).toBe(
      '10',
    )

    act(() => {
      completeBootDownload('app')
    })

    expect(screen.getByRole('progressbar').getAttribute('aria-valuenow')).toBe(
      '25',
    )

    act(() => {
      failBootDownload('styles', 'network error')
    })

    expect(screen.getByRole('alert').textContent).toBe('network error')
    expect(screen.getByRole('button', { name: 'Retry' })).toBeDefined()
    expect(screen.queryByRole('progressbar')).toBeNull()
  })

  it('renders retry and back affordances for startup errors', () => {
    mockProjection.current = {
      ...mockProjection.current,
      view: {
        state: 'error',
        title: 'Connecting to your Space',
        detail: 'Session connection: Downloading the application.',
        progress: 0.3,
        error:
          'Startup did not finish. Check the browser console or startup marks for details.',
      },
      phase: {
        id: 'connect',
        label: 'Connect',
      },
      phases: mockProjection.current.phases.map((phase) => ({
        ...phase,
        state:
          phase.id === 'prepare'
            ? 'complete'
            : phase.id === 'connect'
              ? 'error'
              : 'pending',
      })),
    }

    render(<AppLoadingScreen />)

    expect(screen.getByText('Retry')).toBeDefined()
    expect(screen.getByText('Back')).toBeDefined()
    expect(
      screen.getByText(
        'Startup did not finish. Check the browser console or startup marks for details.',
      ),
    ).toBeDefined()
    expect(screen.getByRole('alert')).toBeDefined()
    expect(screen.queryByRole('progressbar')).toBeNull()
  })

  it('retains startup phases with reduced motion', () => {
    vi.stubGlobal('matchMedia', (query: string) => ({
      matches: query === '(prefers-reduced-motion: reduce)',
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }))

    const { container } = render(<AppLoadingScreen />)

    expect(
      container
        .querySelector('[data-sw-reduced-motion]')
        ?.getAttribute('data-sw-reduced-motion'),
    ).toBe('true')
    expect(screen.getByLabelText('Startup phases')).toBeDefined()
    expect(screen.queryByRole('progressbar')).toBeNull()
  })
})
