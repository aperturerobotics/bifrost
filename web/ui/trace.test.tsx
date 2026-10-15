import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'

import {
  Trace,
  TraceCollapsible,
  TraceCollapsibleContent,
  TraceCollapsibleTrigger,
  TraceLine,
  TraceTitle,
} from './trace.js'

describe('Trace', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders a running line', () => {
    render(
      <Trace>
        <TraceLine status="running">
          <TraceTitle>Reading the kit</TraceTitle>
        </TraceLine>
      </Trace>,
    )

    const line = screen
      .getByText('Reading the kit')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('running')
  })

  it('renders a done line', () => {
    render(
      <Trace>
        <TraceLine status="done">
          <TraceTitle>Landed the change</TraceTitle>
        </TraceLine>
      </Trace>,
    )

    const line = screen
      .getByText('Landed the change')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('done')
  })

  it('renders an error line', () => {
    render(
      <Trace>
        <TraceLine status="error">
          <TraceTitle>Check failed</TraceTitle>
        </TraceLine>
      </Trace>,
    )

    const line = screen
      .getByText('Check failed')
      .closest('[data-slot="trace-line"]')
    expect(line?.getAttribute('data-status')).toBe('error')
  })

  it('renders nested lines in each status', () => {
    render(
      <Trace>
        <TraceCollapsible open onOpenChange={() => undefined}>
          <TraceCollapsibleTrigger>
            <TraceLine status="done">
              <TraceTitle>Parent step</TraceTitle>
            </TraceLine>
          </TraceCollapsibleTrigger>
          <TraceCollapsibleContent>
            <Trace>
              <TraceLine status="running">
                <TraceTitle>Nested running</TraceTitle>
              </TraceLine>
              <TraceLine status="done">
                <TraceTitle>Nested done</TraceTitle>
              </TraceLine>
              <TraceLine status="error">
                <TraceTitle>Nested error</TraceTitle>
              </TraceLine>
            </Trace>
          </TraceCollapsibleContent>
        </TraceCollapsible>
      </Trace>,
    )

    const nested = screen
      .getByText('Nested running')
      .closest('[data-slot="trace"]')
    expect(nested).toBeTruthy()
    expect(
      screen
        .getByText('Nested running')
        .closest('[data-slot="trace-line"]')
        ?.getAttribute('data-status'),
    ).toBe('running')
    expect(
      screen
        .getByText('Nested done')
        .closest('[data-slot="trace-line"]')
        ?.getAttribute('data-status'),
    ).toBe('done')
    expect(
      screen
        .getByText('Nested error')
        .closest('[data-slot="trace-line"]')
        ?.getAttribute('data-status'),
    ).toBe('error')
    expect(
      nested?.closest('[data-slot="trace-collapsible-content"]'),
    ).toBeTruthy()
  })
})
