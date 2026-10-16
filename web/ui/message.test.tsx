import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'

import { Message } from './message.js'

describe('Message', () => {
  afterEach(() => {
    cleanup()
  })

  it("renders the person's turn", () => {
    render(<Message from="user" text="How do I add a settings dialog?" />)

    expect(screen.getByText('How do I add a settings dialog?')).toBeTruthy()
  })

  it('renders the agent reply', () => {
    render(
      <Message
        from="assistant"
        text="Install Dialog, then open it from the gear button."
      />,
    )

    expect(
      screen.getByText('Install Dialog, then open it from the gear button.'),
    ).toBeTruthy()
  })

  it('renders a reply with no actions', () => {
    render(<Message from="assistant" text="Install Dialog." />)

    expect(screen.getByText('Install Dialog.')).toBeTruthy()
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('renders footer actions', () => {
    render(
      <Message
        from="assistant"
        text="Install Dialog."
        actions={<button type="button">Retry</button>}
      />,
    )

    expect(screen.getByRole('button', { name: 'Retry' })).toBeTruthy()
    expect(screen.getByText('Install Dialog.')).toBeTruthy()
  })
})
