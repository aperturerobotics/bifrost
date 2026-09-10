import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { LinkDeviceDoneStep } from './LinkDeviceDoneStep.js'
import type { Session } from '@s4wave/sdk/session/session.js'
import type { ConfirmPairingResponse } from '@s4wave/sdk/session/session.pb.js'

// Completion follows the persisted Session result, independently of peer presence.
describe('LinkDeviceDoneStep', () => {
  afterEach(cleanup)

  it('opens the returned Session only after account enrollment completes', async () => {
    const completion = Promise.withResolvers<ConfirmPairingResponse>()
    const confirmPairing = vi.fn(() => completion.promise)
    const watchPairedDevices = vi.fn()
    const session = { confirmPairing, watchPairedDevices } as unknown as Session
    const onDone = vi.fn()
    render(
      <LinkDeviceDoneStep
        session={session}
        remotePeerId="source-peer"
        onDone={onDone}
        onLinkMore={vi.fn()}
      />,
    )

    await waitFor(() => expect(confirmPairing).toHaveBeenCalled())
    expect(screen.getByText('Opening paired account…')).toBeDefined()
    expect(screen.queryByText('Account connected')).toBeNull()
    expect(screen.queryByRole('button', { name: 'Open account' })).toBeNull()
    expect(watchPairedDevices).not.toHaveBeenCalled()

    const entry = { sessionIndex: 7 }
    completion.resolve({ sessionListEntry: entry })
    expect(await screen.findByText('Account connected')).toBeDefined()
    fireEvent.click(screen.getByRole('button', { name: 'Open account' }))
    expect(onDone).toHaveBeenCalledWith(entry)
  })

  it('shows an enrollment failure without a success or file-copy claim', async () => {
    const session = {
      confirmPairing: vi.fn().mockRejectedValue(new Error('Storage is full')),
    } as unknown as Session
    render(
      <LinkDeviceDoneStep
        session={session}
        remotePeerId="source-peer"
        onDone={vi.fn()}
        onLinkMore={vi.fn()}
      />,
    )

    expect(await screen.findByText('Storage is full')).toBeDefined()
    expect(screen.getByRole('button', { name: 'Try again' })).toBeDefined()
    expect(screen.queryByText('Account connected')).toBeNull()
    expect(screen.queryByRole('button', { name: 'Open account' })).toBeNull()
  })
})
