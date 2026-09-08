import { render, cleanup } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { SharedObjectHealthStatus } from '@s4wave/core/sobject/sobject.pb.js'
import { toast } from '@s4wave/web/ui/toaster.js'

import { SharedObjectSyncNotice } from './SharedObjectSyncNotice.js'

vi.mock('@s4wave/web/ui/toaster.js', () => ({
  toast: { warning: vi.fn(), dismiss: vi.fn() },
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

describe('SharedObjectSyncNotice', () => {
  it('keeps recovery visible after admission and clears only on convergence', () => {
    const { rerender } = render(
      <SharedObjectSyncNotice
        health={{
          status: SharedObjectHealthStatus.READY,
          syncRecoveryPeerIds: ['private-participant'],
          syncDeniedPeerIds: ['private-participant'],
        }}
      />,
    )
    expect(toast.warning).toHaveBeenLastCalledWith(
      'Direct sync needs attention',
      expect.objectContaining({
        description: expect.stringContaining('ask the owner for a new invite'),
        duration: Infinity,
      }),
    )
    rerender(
      <SharedObjectSyncNotice
        health={{
          status: SharedObjectHealthStatus.READY,
          syncRecoveryPeerIds: ['private-participant'],
          syncDeniedPeerIds: [],
        }}
      />,
    )
    expect(toast.warning).toHaveBeenLastCalledWith(
      'Direct sync needs attention',
      expect.objectContaining({
        description: expect.stringContaining(
          'You can keep using local content.',
        ),
      }),
    )
    vi.mocked(toast.dismiss).mockClear()
    rerender(<SharedObjectSyncNotice health={{ syncRecoveryPeerIds: [] }} />)
    expect(toast.dismiss).toHaveBeenCalledTimes(1)
  })

  it('distinguishes a source refusal from loss of local access', () => {
    render(
      <SharedObjectSyncNotice health={{ syncDeniedPeerIds: ['source'] }} />,
    )
    expect(toast.warning).toHaveBeenCalledWith(
      'Direct sync needs attention',
      expect.objectContaining({
        description:
          'A connected device declined to sync this Space. Ask the owner to confirm your access. You can keep using local content.',
      }),
    )
  })
})
