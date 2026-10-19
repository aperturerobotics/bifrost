import { describe, expect, it } from 'vitest'

import { selectStartupView, shouldShowStartupLoading } from './startup-view.js'

describe('startup loading route selection', () => {
  it('shows landing for a new visitor on bare root', () => {
    expect(shouldShowStartupLoading('/', '', false)).toBe(false)
  })

  it('shows loading for hash routes on root', () => {
    expect(shouldShowStartupLoading('/', '#/login', false)).toBe(true)
  })

  it('shows loading for direct app pathnames', () => {
    expect(shouldShowStartupLoading('/login', '', false)).toBe(true)
  })

  it('shows loading after prior interaction', () => {
    expect(shouldShowStartupLoading('/', '', true)).toBe(true)
  })

  it('keeps static pathnames eligible for prerendered startup content', () => {
    expect(shouldShowStartupLoading('/pricing', '', false)).toBe(false)
  })
})

describe('startup view selection', () => {
  it('shows a static page on its pathname without a hash', () => {
    expect(selectStartupView('/pricing', '', true)).toBe('static')
  })

  it('shows landing for a new visitor on bare root', () => {
    expect(selectStartupView('/', '', false)).toBe('landing')
  })

  it('shows loading for app hash routes, even on a static pathname', () => {
    expect(selectStartupView('/', '#/u/1', false)).toBe('loading')
    expect(selectStartupView('/pricing', '#/u/1', false)).toBe('loading')
  })
})
