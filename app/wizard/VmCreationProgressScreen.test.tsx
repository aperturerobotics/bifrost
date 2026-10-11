import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  type VmCreationProgress,
  VmCreationProgressScreen,
} from './VmCreationProgressScreen.js'

afterEach(cleanup)

function renderProgress(
  progress: VmCreationProgress,
  options?: { error?: string; onRetry?: () => void },
) {
  return render(
    <VmCreationProgressScreen
      progress={progress}
      vmName="Debian lab"
      includesCdnLink
      error={options?.error}
      onRetry={options?.onRetry}
    />,
  )
}

describe('VmCreationProgressScreen', () => {
  it.each([
    ['linking', 'Adding the catalog image to this Space.'],
    ['creating', 'Image ready. Writing the VM object.'],
    ['ready', 'Debian lab is ready. Opening it now.'],
  ] as const)('renders the %s progress event', (stage, detail) => {
    renderProgress({ stage })

    expect(screen.getByText(detail)).toBeTruthy()
    expect(screen.getByText('Linking catalog image')).toBeTruthy()
    expect(screen.getByText('Creating VM')).toBeTruthy()
    expect(screen.getByText('Ready')).toBeTruthy()
  })

  it('renders a recoverable error and retries the stream', async () => {
    const user = userEvent.setup()
    const onRetry = vi.fn()
    renderProgress(
      { stage: 'linking' },
      {
        error:
          'The image could not be linked. Check your connection and try again.',
        onRetry,
      },
    )

    expect(screen.getByText('VM could not be created')).toBeTruthy()
    expect(
      screen.getByText(
        'The image could not be linked. Check your connection and try again.',
      ),
    ).toBeTruthy()
    await user.click(screen.getByRole('button', { name: 'Retry' }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })
})
