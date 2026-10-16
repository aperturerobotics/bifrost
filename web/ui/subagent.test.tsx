import { describe, it, expect, afterEach } from 'vitest'
import { fireEvent, render, screen, cleanup } from '@testing-library/react'

import { Subagent } from './subagent.js'

describe('Subagent', () => {
  afterEach(() => {
    cleanup()
  })

  it('shows steps while the child session runs', () => {
    render(
      <Subagent
        name="the research agent"
        status="running"
        steps={[
          { id: 'grep', title: 'Ran grep', status: 'done' },
          { id: 'seed', title: 'Reading the seed', status: 'running' },
        ]}
      />,
    )

    const line = screen
      .getByText('Running the research agent')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('running')
    expect(screen.getByText('Running')).toBeTruthy()
    expect(screen.getByText('Done')).toBeTruthy()
    expect(screen.getByText('Reading the seed')).toBeTruthy()
  })

  it('closes a finished run and shows how long it took', () => {
    render(
      <Subagent
        name="the research agent"
        status="done"
        seconds={48}
        result="The cache is keyed on the commit SHA."
      />,
    )

    const line = screen
      .getByText('Ran the research agent')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('done')
    expect(screen.getByText('48s')).toBeTruthy()
    expect(
      screen.queryByText('The cache is keyed on the commit SHA.'),
    ).toBeNull()
  })

  it('closes when a running session finishes', () => {
    const view = render(
      <Subagent
        name="the research agent"
        status="running"
        task="Find the cache key."
      />,
    )
    expect(screen.getByText('Find the cache key.')).toBeTruthy()

    view.rerender(
      <Subagent
        name="the research agent"
        status="done"
        task="Find the cache key."
      />,
    )
    expect(screen.queryByText('Find the cache key.')).toBeNull()
    expect(screen.getByText('Ran the research agent')).toBeTruthy()
  })

  it('labels the task and the result', () => {
    render(
      <Subagent
        name="the research agent"
        status="running"
        task="Find the cache key."
        result="The cache is keyed on the commit SHA."
      />,
    )

    expect(screen.getByText('Task')).toBeTruthy()
    expect(screen.getByText('Find the cache key.')).toBeTruthy()
    expect(screen.getByText('Result')).toBeTruthy()
    expect(
      screen.getByText('The cache is keyed on the commit SHA.'),
    ).toBeTruthy()
  })

  it('renders a line when the run has no body', () => {
    render(
      <Subagent
        name="the research agent"
        status="done"
        href="/session/child"
      />,
    )

    expect(screen.getByText('Ran the research agent')).toBeTruthy()
    expect(screen.queryByRole('button')).toBeNull()
    expect(
      screen.getByRole('link', { name: 'Open the research agent session' }),
    ).toBeTruthy()
  })

  it('marks a failed run', () => {
    render(<Subagent name="the research agent" status="error" />)

    const line = screen
      .getByText('Failed to run the research agent')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('error')
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('keeps the session link outside the trigger', () => {
    render(
      <Subagent
        name="the research agent"
        status="done"
        href="/session/child"
        task="Find the cache key."
      />,
    )

    const link = screen.getByRole('link', {
      name: 'Open the research agent session',
    })
    expect(link.getAttribute('href')).toBe('/session/child')
    expect(link.closest('button')).toBeNull()
    fireEvent.click(
      screen.getByRole('button', { name: /Ran the research agent/ }),
    )
    expect(screen.getByText('Task')).toBeTruthy()
    expect(screen.getByText('Find the cache key.')).toBeTruthy()
  })

  it('renders the model and tools', () => {
    render(
      <Subagent
        name="the research agent"
        status="running"
        model="claude-sonnet"
        tools={['runCommand', 'search']}
      />,
    )

    expect(screen.getByText('claude-sonnet')).toBeTruthy()
    expect(screen.getByText('runCommand')).toBeTruthy()
    expect(screen.getByText('search')).toBeTruthy()
  })
})
