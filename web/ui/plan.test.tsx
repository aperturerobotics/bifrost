import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'

import { Plan } from './plan.js'

describe('Plan', () => {
  afterEach(() => {
    cleanup()
  })

  it('marks the header running while a step is running', () => {
    render(
      <Plan
        title="Plan: migrate the dialog"
        steps={[
          { id: 'install', title: 'Install the package', status: 'done' },
          { id: 'rename', title: 'Rename the props', status: 'running' },
          { id: 'check', title: 'Run the type checker', status: 'pending' },
        ]}
      />,
    )

    const line = screen
      .getByText('Plan: migrate the dialog')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('running')
    expect(screen.getByText('1 of 3 done')).toBeTruthy()
    expect(screen.getByText('Running')).toBeTruthy()
    expect(screen.getByText('Done')).toBeTruthy()
    expect(screen.getByText('Pending')).toBeTruthy()
  })

  it('marks the header error when a step failed and none is running', () => {
    render(
      <Plan
        title="Plan: migrate the dialog"
        steps={[
          { id: 'install', title: 'Install the package', status: 'done' },
          { id: 'check', title: 'Run the type checker', status: 'error' },
        ]}
      />,
    )

    const line = screen
      .getByText('Plan: migrate the dialog')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('error')
    expect(screen.getByText('Failed')).toBeTruthy()
  })

  it('marks the header pending when work remains and nothing failed', () => {
    render(
      <Plan
        title="Plan: migrate the dialog"
        steps={[
          { id: 'install', title: 'Install the package', status: 'done' },
          { id: 'check', title: 'Run the type checker', status: 'pending' },
        ]}
      />,
    )

    const line = screen
      .getByText('Plan: migrate the dialog')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('pending')
  })

  it('settles the header once every step is done', () => {
    render(
      <Plan
        title="Plan: migrate the dialog"
        description="Replace the Radix imports."
        steps={[
          { id: 'install', title: 'Install the package', status: 'done' },
        ]}
      />,
    )

    const line = screen
      .getByText('Plan: migrate the dialog')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('done')
    expect(screen.getByText('1 of 1 done')).toBeTruthy()
    expect(screen.getByText('Replace the Radix imports.')).toBeTruthy()
    expect(screen.getByText('Done')).toBeTruthy()
  })
})
