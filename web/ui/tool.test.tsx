import { describe, it, expect, afterEach } from 'vitest'
import { fireEvent, render, screen, cleanup } from '@testing-library/react'

import { Tool } from './tool.js'

describe('Tool', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders a running call as a line', () => {
    render(<Tool name="pnpm test" status="running" />)

    const line = screen
      .getByText('Running pnpm test')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('running')
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('opens the output of a finished call', () => {
    render(
      <Tool
        name="git status"
        status="done"
        input="git status --short"
        output="modified app.ts"
      />,
    )

    const line = screen
      .getByText('Ran git status')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('done')
    expect(screen.queryByText('modified app.ts')).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Ran git status' }))

    const output = screen
      .getByText('modified app.ts')
      .closest('[data-slot="code-block"]')
    expect(output?.getAttribute('data-variant')).toBe('default')
    expect(screen.getByText('git status --short')).toBeTruthy()
  })

  it('opens a finished call that has only input', () => {
    render(<Tool name="git status" status="done" input="git status --short" />)

    expect(screen.queryByText('git status --short')).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Ran git status' }))
    expect(screen.getByText('git status --short')).toBeTruthy()
  })

  it('opens a failed call in a destructive code block', () => {
    render(
      <Tool name="pnpm dev" status="error" output="command not found: pnpm" />,
    )

    const line = screen
      .getByText('Failed to run pnpm dev')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('error')
    fireEvent.click(
      screen.getByRole('button', { name: 'Failed to run pnpm dev' }),
    )

    const output = screen
      .getByText('command not found: pnpm')
      .closest('[data-slot="code-block"]')
    expect(output?.getAttribute('data-variant')).toBe('destructive')
  })

  it('keeps a failed call with no output as a line', () => {
    render(<Tool name="pnpm dev" status="error" />)

    expect(screen.getByText('Failed to run pnpm dev')).toBeTruthy()
    expect(screen.queryByRole('button')).toBeNull()
  })
})
