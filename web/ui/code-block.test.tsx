import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'

import { CodeBlock, CodeBlockContent, CodeBlockLine } from './code-block.js'

describe('CodeBlock', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders a diff', () => {
    render(
      <CodeBlock>
        <CodeBlockContent>
          <CodeBlockLine variant="removed">old line</CodeBlockLine>
          <CodeBlockLine variant="added">new line</CodeBlockLine>
        </CodeBlockContent>
      </CodeBlock>,
    )

    const removed = screen
      .getByText('old line')
      .closest('[data-slot="code-block-line"]')
    const added = screen
      .getByText('new line')
      .closest('[data-slot="code-block-line"]')
    expect(removed?.getAttribute('data-variant')).toBe('removed')
    expect(added?.getAttribute('data-variant')).toBe('added')
    expect(
      screen.getByText('old line').closest('[data-slot="code-block"]'),
    ).toBeTruthy()
  })
})
