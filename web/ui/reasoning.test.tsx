import { describe, it, expect, afterEach } from 'vitest'
import { fireEvent, render, screen, cleanup } from '@testing-library/react'

import { Reasoning } from './reasoning.js'

describe('Reasoning', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders a running line before any text arrives', () => {
    render(<Reasoning text="" status="running" />)

    const line = screen
      .getByText('Thinking')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('running')
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('opens the text while thinking', () => {
    render(<Reasoning text="The lockfile is stale." status="running" />)

    expect(screen.getByText('Thinking')).toBeTruthy()
    expect(screen.getByText('The lockfile is stale.')).toBeTruthy()
  })

  it('closes the text once thinking finishes', () => {
    render(<Reasoning text="The lockfile is stale." seconds={5} />)

    const line = screen.getByText('Thought').closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('done')
    expect(screen.getByText('5s')).toBeTruthy()
    expect(screen.queryByText('The lockfile is stale.')).toBeNull()
  })

  it("keeps a person's open choice after thinking finishes", () => {
    const view = render(<Reasoning text="The lockfile is stale." />)

    fireEvent.click(screen.getByRole('button', { name: 'Thought' }))
    view.rerender(<Reasoning text="The lockfile is stale." status="running" />)
    view.rerender(<Reasoning text="The lockfile is stale." />)

    expect(screen.getByText('The lockfile is stale.')).toBeTruthy()
  })
})
