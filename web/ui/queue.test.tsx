import { describe, it, expect, afterEach } from 'vitest'
import { fireEvent, render, screen, cleanup } from '@testing-library/react'

import { Queue } from './queue.js'

const items = [
  { id: 'toggle', title: 'Add the toggle', status: 'done' as const },
  {
    id: 'contrast',
    title: 'Check contrast',
    description: 'Charts fail the contrast check.',
    status: 'running' as const,
  },
  { id: 'note', title: 'Write the note', status: 'pending' as const },
]

describe('Queue', () => {
  afterEach(() => {
    cleanup()
  })

  it('splits what is left from what is done', () => {
    render(<Queue items={items} />)

    expect(screen.getByText('2 to do')).toBeTruthy()
    expect(screen.getByText('1 completed')).toBeTruthy()
    expect(screen.getByText('Add the toggle')).toBeTruthy()
    expect(screen.getByText('Write the note')).toBeTruthy()
  })

  it('names each item status for assistive tech', () => {
    render(<Queue items={items} />)

    expect(screen.getByText('Running')).toBeTruthy()
    expect(screen.getByText('Pending')).toBeTruthy()
    expect(screen.getByText('Done')).toBeTruthy()
  })

  it('opens a description under its item', () => {
    render(<Queue items={items} />)

    expect(screen.queryByText('Charts fail the contrast check.')).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: /Check contrast/ }))
    expect(screen.getByText('Charts fail the contrast check.')).toBeTruthy()
  })

  it('renders no groups when the list is empty', () => {
    render(<Queue items={[]} />)

    expect(screen.queryByText(/to do/)).toBeNull()
    expect(screen.queryByText(/completed/)).toBeNull()
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('renders only the completed group when every item is done', () => {
    render(
      <Queue
        items={[{ id: 'toggle', title: 'Add the toggle', status: 'done' }]}
      />,
    )

    expect(screen.getByText('1 completed')).toBeTruthy()
    expect(screen.queryByText(/to do/)).toBeNull()
    expect(screen.getByText('Add the toggle')).toBeTruthy()
    expect(screen.getByText('Done')).toBeTruthy()
  })
})
