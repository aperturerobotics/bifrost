import { useCallback, useState } from 'react'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  resolvePath,
  Route,
  Router,
  Routes,
  type To,
} from '@s4wave/web/router/router.js'
import { worldObjectNavigation } from '@s4wave/web/router/HistoryRouter.js'
import { parseObjectUri } from '@s4wave/sdk/space/object-uri.js'

import { SpaceObjectContainer } from './SpaceObjectContainer.js'

const wizardPath = '/u/1/so/space-1/-/wizard/device-1'

afterEach(cleanup)

const h = vi.hoisted(() => ({
  viewer: vi.fn(),
  navigateToRoot: vi.fn(),
  navigateToSubPath: vi.fn(),
  spaceContentsResource: {
    value: {},
    loading: false,
    error: null,
    retry: vi.fn(),
  },
  spaceContext: {
    spaceId: 'space-1',
    objectKey: 'wizard',
    objectPath: 'device-1',
    spaceWorldResource: {
      value: null as null | {
        openNestedWorld: (key: string, signal: AbortSignal) => Promise<unknown>
      },
      loading: false,
      error: null,
      retry: vi.fn(),
    },
  },
}))

vi.mock('@s4wave/web/object/ObjectViewer.js', () => ({
  ObjectViewer: ({
    onNavigate,
    ...props
  }: {
    onNavigate?: (to: { path: string }) => void
    objectInfo?: { info?: { value?: { objectKey?: string } } }
    path?: string
    worldState?: { value: unknown }
  }) => {
    h.viewer({ onNavigate, ...props })
    return (
      <button type="button" onClick={() => onNavigate?.({ path: '/login' })}>
        Sign in or create account
      </button>
    )
  },
}))

vi.mock('@s4wave/web/contexts/contexts.js', () => ({
  SpaceContentsContext: {
    useContext: () => h.spaceContentsResource,
  },
  useSessionIndex: () => 1,
}))

vi.mock('@s4wave/web/contexts/SpaceContainerContext.js', () => ({
  SpaceContainerContext: {
    useContext: () => ({
      ...h.spaceContext,
      navigateToRoot: h.navigateToRoot,
      navigateToSubPath: h.navigateToSubPath,
    }),
  },
}))

vi.mock('@s4wave/web/hooks/useObjectMetadata.js', () => ({
  useObjectMetadata: () => ({
    value: 'spacewave/wizard',
    loading: false,
    error: null,
    retry: vi.fn(),
  }),
}))

vi.mock('@s4wave/app/quickstart/session-handoff.js', () => ({
  getQuickstartInitialObjectHandoff: vi.fn(),
}))

/** AppRouteHarness routes the container beside a login page with a history stack. */
function AppRouteHarness() {
  // Keep a history stack that the login page's back button can pop.
  const [entries, setEntries] = useState([wizardPath])
  const [entryIndex, setEntryIndex] = useState(0)
  const path = entries[entryIndex] ?? wizardPath
  const navigate = useCallback(
    (to: To) => {
      const next = resolvePath(path, to)
      if (to.replace) {
        setEntries((current) => current.with(entryIndex, next))
        return
      }
      setEntries((current) => [...current.slice(0, entryIndex + 1), next])
      setEntryIndex((current) => current + 1)
    },
    [entryIndex, path],
  )
  const goBack = useCallback(() => {
    setEntryIndex((current) => Math.max(0, current - 1))
  }, [])

  return (
    <>
      <output data-testid="current-route">{path}</output>
      <Router path={path} onNavigate={navigate}>
        <Routes fullPath>
          <Route path="/u/:sessionIndex/so/:spaceId/-/*">
            <SpaceObjectContainer />
          </Route>
          <Route path="/login">
            <div>
              <p>Spacewave Cloud login</p>
              <button type="button" onClick={goBack}>
                Back to device wizard
              </button>
            </div>
          </Route>
        </Routes>
      </Router>
    </>
  )
}

describe('SpaceObjectContainer authentication navigation', () => {
  it('opens the app login route and keeps the device wizard as the back target', () => {
    // Open the login route from the viewer.
    render(<AppRouteHarness />)
    fireEvent.click(
      screen.getByRole('button', { name: 'Sign in or create account' }),
    )

    // The app route changes without a Space subpath navigation.
    expect(screen.getByText('Spacewave Cloud login')).toBeTruthy()
    expect(screen.getByTestId('current-route').textContent).toBe('/login')
    expect(h.navigateToSubPath).not.toHaveBeenCalled()

    // Going back returns to the device wizard.
    fireEvent.click(
      screen.getByRole('button', { name: 'Back to device wizard' }),
    )
    expect(
      screen.getByRole('button', { name: 'Sign in or create account' }),
    ).toBeTruthy()
    expect(screen.getByTestId('current-route').textContent).toBe(wizardPath)
  })
})

describe('nested World route', () => {
  it('decodes the outer key and nested viewer subpath from a deep link', () => {
    // Parse the outer route, then the nested route inside its path.
    const outer = parseObjectUri(
      'projection/logbook/-/world/-/orient/plan/first/-/details',
    )
    const inner = parseObjectUri(outer.path.slice('world/-/'.length))

    // Each level carries its own key and subpath.
    expect(outer.objectKey).toBe('projection/logbook')
    expect(inner).toEqual({
      objectKey: 'orient/plan/first',
      path: 'details',
    })
  })

  it('opens the outer object, navigates within the inner viewer, and releases on route change', async () => {
    // Route to an object inside the outer object's nested World.
    const release = vi.fn()
    const nestedWorld = { [Symbol.dispose]: release, getReadOnly: () => true }
    const openNestedWorld = vi.fn().mockResolvedValue(nestedWorld)
    h.spaceContext.spaceWorldResource.value = { openNestedWorld }
    h.spaceContext.objectKey = 'projection/logbook'
    h.spaceContext.objectPath = 'world/-/orient/plan/first/-/details'
    h.viewer.mockClear()

    // The viewer shows the inner object from the opened nested World.
    const view = render(
      <Router path={wizardPath} onNavigate={vi.fn()}>
        <SpaceObjectContainer />
      </Router>,
    )
    await waitFor(() => expect(h.viewer).toHaveBeenCalled())
    const props = h.viewer.mock.lastCall?.[0]
    expect(openNestedWorld).toHaveBeenCalledWith(
      'projection/logbook',
      expect.any(AbortSignal),
    )
    expect(props.objectInfo.info.value.objectKey).toBe('orient/plan/first')
    expect(props.worldState.value).toBe(nestedWorld)
    expect(props.path).toBe('/details')

    // Local and sibling navigation stays inside the nested route.
    props.onNavigate({ path: 'history' })
    expect(h.navigateToSubPath).toHaveBeenCalledWith(
      'projection/logbook/-/world/-/orient/plan/first/-/details/history',
    )
    props.onNavigate(worldObjectNavigation('plans/second.org'))
    expect(h.navigateToSubPath).toHaveBeenLastCalledWith(
      'projection/logbook/-/world/-/plans/second.org',
    )
    props.onNavigate(worldObjectNavigation('orient/repo', 'plans'))
    expect(h.navigateToSubPath).toHaveBeenLastCalledWith(
      'projection/logbook/-/world/-/orient/repo/-/plans',
    )

    // Leaving the nested route releases the nested World.
    h.spaceContext.objectPath = ''
    view.rerender(
      <Router path={wizardPath} onNavigate={vi.fn()}>
        <SpaceObjectContainer />
      </Router>,
    )
    await waitFor(() => expect(release).toHaveBeenCalledTimes(1))

    // Returning opens it again, and unmounting releases it.
    h.spaceContext.objectPath = 'world/-/orient/issue/second'
    view.rerender(
      <Router path={wizardPath} onNavigate={vi.fn()}>
        <SpaceObjectContainer />
      </Router>,
    )
    await waitFor(() => expect(openNestedWorld).toHaveBeenCalledTimes(2))
    view.unmount()
    expect(release).toHaveBeenCalledTimes(2)
  })

  it('shows an empty route without opening a World', () => {
    // Route to a nested World without an inner object.
    const openNestedWorld = vi.fn()
    h.spaceContext.spaceWorldResource.value = { openNestedWorld }
    h.spaceContext.objectKey = 'projection/logbook'
    h.spaceContext.objectPath = 'world/-/'

    // The empty state shows and no World opens.
    render(
      <Router path={wizardPath} onNavigate={vi.fn()}>
        <SpaceObjectContainer />
      </Router>,
    )
    expect(screen.getByRole('status').textContent).toContain('No nested object')
    expect(openNestedWorld).not.toHaveBeenCalled()
  })

  it('shows an open error with retry instead of an indefinite loading state', async () => {
    // Route to a nested World that fails to open.
    const openNestedWorld = vi.fn().mockRejectedValue(new Error('Unavailable'))
    h.spaceContext.spaceWorldResource.value = { openNestedWorld }
    h.spaceContext.objectKey = 'projection/logbook'
    h.spaceContext.objectPath = 'world/-/orient/plan/first'

    // The error and a retry show in place of the loading state.
    render(
      <Router path={wizardPath} onNavigate={vi.fn()}>
        <SpaceObjectContainer />
      </Router>,
    )
    expect((await screen.findByRole('alert')).textContent).toContain(
      'Unavailable',
    )
    expect(screen.getByRole('button', { name: 'Try again' })).toBeTruthy()
  })
})
