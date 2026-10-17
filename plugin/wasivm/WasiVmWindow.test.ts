import { describe, expect, it } from 'vitest'

import { formatWindowPath, parseWindowPath } from './WasiVmWindow.js'

describe('window paths', () => {
  it('round trips a window of a run', () => {
    const route = { run: 3n, id: 7 }
    expect(parseWindowPath(formatWindowPath(route))).toEqual(route)
    expect(parseWindowPath('/window/3/7')).toEqual(route)
  })

  it('reads other paths as the VM view', () => {
    expect(parseWindowPath('/')).toBeNull()
    expect(parseWindowPath('/window/3')).toBeNull()
    expect(parseWindowPath('/window/x/7')).toBeNull()
  })
})
