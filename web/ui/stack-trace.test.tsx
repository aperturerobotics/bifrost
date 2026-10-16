import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'

import { StackTrace } from './stack-trace.js'

describe('StackTrace', () => {
  afterEach(() => {
    cleanup()
  })

  it('marks the failure and repeats the error', () => {
    render(
      <StackTrace
        error="Cannot read email"
        frames={[{ text: 'at seedUsers (prisma/seed.ts:27:34)' }]}
      />,
    )

    const line = screen
      .getByText('Failed: Cannot read email')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('error')
    expect(screen.getByText('Cannot read email')).toBeTruthy()
  })

  it('links an own frame without putting it in a button', () => {
    render(
      <StackTrace
        error="Cannot read email"
        frames={[
          {
            text: 'at seedUsers (prisma/seed.ts:27:34)',
            href: 'prisma/seed.ts',
          },
        ]}
      />,
    )

    const frame = screen.getByRole('link', {
      name: 'at seedUsers (prisma/seed.ts:27:34)',
    })
    expect(frame.getAttribute('href')).toBe('prisma/seed.ts')
    expect(frame.closest('button')).toBeNull()
  })

  it('renders repeated frames', () => {
    render(
      <StackTrace
        error="Cannot read email"
        frames={[
          { text: 'at recurse (app.ts:1:1)' },
          { text: 'at recurse (app.ts:1:1)' },
        ]}
      />,
    )

    expect(screen.getAllByText('at recurse (app.ts:1:1)')).toHaveLength(2)
  })
})
