import { afterEach, expect, it, vi } from 'vitest'
import { flushSync } from 'react-dom'
import { createRoot, type Root } from 'react-dom/client'

import { SpaceMountingScreen } from '@s4wave/app/space/SpaceMountingScreen.js'

let root: Root | undefined
let host: HTMLDivElement | undefined

afterEach(() => {
  if (root) flushSync(() => root!.unmount())
  host?.remove()
  root = undefined
  host = undefined
  vi.restoreAllMocks()
})

it('paints continuous artwork across Space loading owners before presenting the new canvas', () => {
  host = document.createElement('div')
  host.style.cssText = 'position:fixed;inset:0;display:flex'
  document.body.append(host)
  root = createRoot(host)
  const clock = vi.spyOn(performance, 'now').mockReturnValue(1000)

  flushSync(() => {
    root!.render(
      <SpaceMountingScreen
        key="lookup"
        stage="resolve"
        detail="Looking up the shared object."
      />,
    )
  })
  const first = host.querySelector('canvas')!
  expect(first.dataset.renderer).toBe('three')
  const gl = first.getContext('webgl2')!
  const pixels = new Uint8Array(
    gl.drawingBufferWidth * gl.drawingBufferHeight * 4,
  )
  gl.readPixels(
    0,
    0,
    gl.drawingBufferWidth,
    gl.drawingBufferHeight,
    gl.RGBA,
    gl.UNSIGNED_BYTE,
    pixels,
  )
  expect(pixels.some((value, index) => index % 4 !== 3 && value > 0)).toBe(true)
  const initialFrame = first.toDataURL()

  flushSync(() => {
    root!.render(
      <SpaceMountingScreen
        key="lookup"
        stage="mount"
        detail="Mounting the space."
      />,
    )
  })
  expect(host.querySelector('canvas')).toBe(first)

  clock.mockReturnValue(6000)
  flushSync(() => {
    root!.render(
      <SpaceMountingScreen
        key="health"
        stage="mount"
        detail="Mounting the space."
      />,
    )
  })
  const second = host.querySelector('canvas')!
  expect(second).not.toBe(first)
  expect(first.hasAttribute('data-renderer')).toBe(false)
  const advancedFrame = second.toDataURL()
  expect(advancedFrame === initialFrame).toBe(false)

  flushSync(() => {
    root!.render(
      <SpaceMountingScreen
        key="world"
        stage="sync"
        detail="Preparing the space contents."
      />,
    )
  })
  const third = host.querySelector('canvas')!
  expect(third).not.toBe(second)
  expect(second.hasAttribute('data-renderer')).toBe(false)
  expect(third.toDataURL() === advancedFrame).toBe(true)
})
