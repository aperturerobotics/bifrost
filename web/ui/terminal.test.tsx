import { describe, it, expect, afterEach } from 'vitest'
import { fireEvent, render, screen, cleanup } from '@testing-library/react'

import { Terminal } from './terminal.js'

describe('Terminal', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders a running command before output arrives', () => {
    render(<Terminal command="pnpm test" />)

    const line = screen
      .getByText('Running pnpm test')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('running')
    expect(screen.queryByRole('button')).toBeNull()
    expect(screen.queryByText('15 passed')).toBeNull()
  })

  it('opens once output arrives and stays open after the command exits', () => {
    const view = render(<Terminal command="pnpm test" />)

    view.rerender(<Terminal command="pnpm test" output="15 passed" />)
    expect(screen.getByText('15 passed')).toBeTruthy()

    view.rerender(
      <Terminal command="pnpm test" output="15 passed" exitCode={0} />,
    )
    const line = screen
      .getByText('Ran pnpm test')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('done')
    expect(screen.getByText('15 passed')).toBeTruthy()
  })

  it('keeps a close choice after output arrives', () => {
    const view = render(<Terminal command="pnpm test" output="15 passed" />)

    fireEvent.click(screen.getByRole('button', { name: 'Running pnpm test' }))
    view.rerender(
      <Terminal command="pnpm test" output="15 passed" exitCode={0} />,
    )

    expect(screen.queryByText('15 passed')).toBeNull()
  })

  it('names a non-zero exit and marks the output destructive', () => {
    render(<Terminal command="pnpm lint" output="2 errors" exitCode={1} />)

    const line = screen
      .getByText('pnpm lint exited with code 1')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('error')
    const output = screen
      .getByText('2 errors')
      .closest('[data-slot="code-block"]')
    expect(output?.getAttribute('data-variant')).toBe('destructive')
  })

  it('renders a finished command with no output as a line', () => {
    render(<Terminal command="true" exitCode={0} />)

    expect(screen.getByText('Ran true')).toBeTruthy()
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('renders a failed command with no output as a line', () => {
    render(<Terminal command="false" exitCode={1} />)

    expect(screen.getByText('false exited with code 1')).toBeTruthy()
    expect(screen.queryByRole('button')).toBeNull()
  })
})
