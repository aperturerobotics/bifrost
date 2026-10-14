import { describe, expect, it } from 'vitest'

import { keysymOf, xButtons } from './display-input.js'

describe('keysymOf', () => {
  it('maps characters, named keys and modifiers to X keysyms', () => {
    expect(keysymOf('a', 'KeyA')).toBe(0x61)
    expect(keysymOf('A', 'KeyA')).toBe(0x41)
    expect(keysymOf('é', 'KeyE')).toBe(0xe9)
    expect(keysymOf('€', 'KeyE')).toBe(0x010020ac)
    expect(keysymOf('Enter', 'Enter')).toBe(0xff0d)
    expect(keysymOf('F12', 'F12')).toBe(0xffc9)
    expect(keysymOf('Shift', 'ShiftRight')).toBe(0xffe2)
    expect(keysymOf('Dead', 'Quote')).toBe(0)
  })
})

describe('xButtons', () => {
  it('swaps the middle and right buttons', () => {
    expect(xButtons(1)).toBe(1)
    expect(xButtons(2)).toBe(4)
    expect(xButtons(4)).toBe(2)
    expect(xButtons(7)).toBe(7)
  })
})
