import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'

import { Context } from './context.js'

describe('Context', () => {
  afterEach(() => {
    cleanup()
  })

  it('shows the percent and the used limit', () => {
    render(<Context used={32000} limit={128000} />)

    const meter = screen.getByRole('progressbar', { name: 'Context used' })
    expect(meter.getAttribute('aria-valuenow')).toBe('25')
    expect(meter.getAttribute('aria-valuetext')).toBe('32000 of 128000')
    expect(screen.getByText('25%')).toBeTruthy()
    expect(screen.getByText('32000 / 128000')).toBeTruthy()
  })

  it('caps the meter when used is over the limit', () => {
    render(<Context used={150} limit={100} />)

    const meter = screen.getByRole('progressbar', { name: 'Context used' })
    expect(meter.getAttribute('aria-valuenow')).toBe('100')
    expect(meter.getAttribute('aria-valuetext')).toBe('150 of 100')
    expect(screen.getByText('100%')).toBeTruthy()
    expect(screen.getByText('150 / 100')).toBeTruthy()
  })

  it('clamps a negative used count to zero', () => {
    render(<Context used={-20} limit={100} />)

    const meter = screen.getByRole('progressbar', { name: 'Context used' })
    expect(meter.getAttribute('aria-valuenow')).toBe('0')
    expect(meter.getAttribute('aria-valuetext')).toBe('0 of 100')
    expect(screen.getByText('0 / 100')).toBeTruthy()
  })

  it('stays empty when the limit is zero', () => {
    render(<Context used={10} limit={0} />)

    const meter = screen.getByRole('progressbar', { name: 'Context used' })
    expect(meter.getAttribute('aria-valuenow')).toBe('0')
    expect(meter.getAttribute('aria-valuetext')).toBe('10 of 0')
    expect(screen.getByText('0%')).toBeTruthy()
    expect(screen.getByText('10 / 0')).toBeTruthy()
  })
})
