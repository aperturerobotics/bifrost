import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, render, screen } from '@testing-library/react'

import { AppSession } from './AppSession.js'
import { SessionLockMode } from '@s4wave/core/session/session.pb.js'

const mockUseParams = vi.hoisted(() => vi.fn())
const mockUsePath = vi.hoisted(() => vi.fn(() => '/u/1'))
const mockUseNavigate = vi.hoisted(() => vi.fn())
const mockUseSessionMount = vi.hoisted(() => vi.fn())
const mockUseRootResource = vi.hoisted(() => vi.fn())
const mockUseSessionMetadata = vi.hoisted(() => vi.fn())

vi.mock('@s4wave/web/router/router.js', () => ({
  resolvePath: vi.fn(),
  useNavigate: () => mockUseNavigate,
  useParams: mockUseParams,
  usePath: mockUsePath,
}))

vi.mock('./hooks/useSharedMounts.js', () => ({
  useSessionMount: mockUseSessionMount,
}))

vi.mock('@s4wave/web/hooks/useRootResource.js', () => ({
  useRootResource: mockUseRootResource,
}))

vi.mock('@s4wave/web/state/interaction.js', () => ({
  markInteracted: vi.fn(),
}))

vi.mock('@s4wave/web/contexts/contexts.js', () => ({
  SessionIndexContext: {
    Provider: ({ children }: { children?: ReactNode }) => <>{children}</>,
  },
  SessionRouteContext: {
    Provider: ({ children }: { children?: ReactNode }) => <>{children}</>,
  },
}))

vi.mock('@s4wave/app/hooks/useSessionMetadata.js', () => ({
  useSessionMetadata: mockUseSessionMetadata,
}))

vi.mock('./session/SessionContainer.js', () => ({
  SessionContainer: () => <div data-testid="session-container" />,
}))

vi.mock('./session/PinUnlockOverlay.js', () => ({
  PinUnlockOverlay: () => <div data-testid="pin-unlock-overlay" />,
}))

describe('AppSession', () => {
  afterEach(() => {
    cleanup()
    vi.clearAllMocks()
    mockUsePath.mockReturnValue('/u/1')
  })

  it('describes a pending session attachment as connecting', () => {
    mockUseParams.mockReturnValue({ sessionIndex: '1' })
    mockUseRootResource.mockReturnValue({ value: null })
    mockUseSessionMetadata.mockReturnValue(null)
    mockUseSessionMount.mockReturnValue({
      value: null,
      loading: true,
      error: null,
      retry: vi.fn(),
    })

    render(<AppSession />)

    expect(screen.getByText('Connecting to the session.')).toBeTruthy()
  })

  it('renders the pre-mount PIN unlock overlay for locked sessions', () => {
    mockUseParams.mockReturnValue({ sessionIndex: '1' })
    mockUseRootResource.mockReturnValue({ value: null })
    mockUseSessionMetadata.mockReturnValue({
      lockMode: SessionLockMode.PIN_ENCRYPTED,
      displayName: 'Cloud Session',
    })
    mockUseSessionMount.mockReturnValue({
      value: null,
      loading: false,
      error: null,
      retry: vi.fn(),
    })

    render(<AppSession />)

    expect(screen.getByTestId('pin-unlock-overlay')).toBeTruthy()
    expect(screen.queryByTestId('session-container')).toBeNull()
  })

  it('renders the PIN unlock overlay while switching to a locked session', () => {
    mockUseParams.mockReturnValue({ sessionIndex: '2' })
    mockUsePath.mockReturnValue('/u/2')
    mockUseRootResource.mockReturnValue({ value: null })
    mockUseSessionMetadata.mockReturnValue({
      lockMode: SessionLockMode.PIN_ENCRYPTED,
      displayName: 'Cloud Session',
    })
    // A remount retains the previous session's value while loading.
    mockUseSessionMount.mockReturnValue({
      value: {} as never,
      loading: true,
      error: null,
      retry: vi.fn(),
    })

    render(<AppSession />)

    expect(screen.getByTestId('pin-unlock-overlay')).toBeTruthy()
    expect(screen.queryByTestId('session-container')).toBeNull()
  })

  it('renders the session container once the session is mounted', () => {
    mockUseParams.mockReturnValue({ sessionIndex: '1' })
    mockUseRootResource.mockReturnValue({ value: null })
    mockUseSessionMetadata.mockReturnValue({
      lockMode: SessionLockMode.PIN_ENCRYPTED,
      displayName: 'Cloud Session',
    })
    mockUseSessionMount.mockReturnValue({
      value: {} as never,
      loading: false,
      error: null,
      retry: vi.fn(),
    })

    render(<AppSession />)

    expect(screen.getByTestId('session-container')).toBeTruthy()
    expect(screen.queryByTestId('pin-unlock-overlay')).toBeNull()
  })

  it('leaves a Session index that does not exist', () => {
    mockUseParams.mockReturnValue({ sessionIndex: '9' })
    mockUsePath.mockReturnValue('/u/9')
    mockUseRootResource.mockReturnValue({ value: {} })
    mockUseSessionMetadata.mockReturnValue(null)
    mockUseSessionMount.mockReturnValue({
      value: null,
      loading: false,
      error: null,
      retry: vi.fn(),
    })

    render(<AppSession />)

    expect(mockUseNavigate).toHaveBeenCalledWith({ path: '/', replace: true })
  })

  it('refreshes a moved account after leaving the pairing flow', () => {
    mockUseParams.mockReturnValue({ sessionIndex: '1' })
    mockUseRootResource.mockReturnValue({ value: null })
    mockUseSessionMount.mockReturnValue({ value: null, loading: true })
    mockUseSessionMetadata.mockReturnValue({
      providerId: 'local',
      providerAccountId: 'source',
    })
    mockUsePath.mockReturnValue('/u/1/setup/link-device')
    const { rerender } = render(<AppSession />)
    const generation = () => mockUseSessionMount.mock.lastCall?.[2]?.generation
    const pairingGeneration = generation()

    mockUseSessionMetadata.mockReturnValue({
      providerId: 'local',
      providerAccountId: 'destination',
    })
    rerender(<AppSession />)
    expect(generation()).toEqual(pairingGeneration)

    mockUsePath.mockReturnValue('/u/1/so/drive')
    rerender(<AppSession />)
    expect(generation()).toEqual('local/destination')
    mockUsePath.mockReturnValue('/u/1/so/another-drive')
    rerender(<AppSession />)
    expect(generation()).toEqual('local/destination')
  })

  it('does not remount when initial Session metadata arrives', () => {
    mockUseParams.mockReturnValue({ sessionIndex: '1' })
    mockUseRootResource.mockReturnValue({ value: null })
    mockUseSessionMount.mockReturnValue({ value: null, loading: true })
    mockUseSessionMetadata.mockReturnValue(null)
    const { rerender } = render(<AppSession />)
    const generation = () => mockUseSessionMount.mock.lastCall?.[2]?.generation
    expect(generation()).toEqual('')

    mockUseSessionMetadata.mockReturnValue({
      providerId: 'local',
      providerAccountId: 'source',
    })
    rerender(<AppSession />)
    expect(generation()).toEqual('')
  })

  it('remounts when an established Session moves accounts', () => {
    mockUseParams.mockReturnValue({ sessionIndex: '1' })
    mockUseRootResource.mockReturnValue({ value: null })
    mockUseSessionMount.mockReturnValue({ value: null, loading: true })
    mockUseSessionMetadata.mockReturnValue({
      providerId: 'local',
      providerAccountId: 'source',
    })
    const { rerender } = render(<AppSession />)
    const generation = () => mockUseSessionMount.mock.lastCall?.[2]?.generation
    expect(generation()).toEqual('')

    mockUseSessionMetadata.mockReturnValue({
      providerId: 'local',
      providerAccountId: 'destination',
    })
    rerender(<AppSession />)
    expect(generation()).toEqual('local/destination')
  })
})
