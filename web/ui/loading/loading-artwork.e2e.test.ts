import { afterEach, expect, it, vi } from 'vitest'
import { ShaderPass } from 'three/addons/postprocessing/ShaderPass.js'

import { createLoadingArtwork } from './loading-artwork.js'
import { LOADING_SCREEN_CSS } from './loading-screen-style.js'

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

it('bakes the sky once per geometry and limits foreground rendering to 30 fps', async () => {
  const host = document.createElement('div')
  host.style.cssText = 'position:fixed;inset:0'
  host.innerHTML = `<style>${LOADING_SCREEN_CSS}</style>
    <div class="swl-canvas">
      <canvas class="swl-artwork"></canvas>
      <div class="swl-emblem"></div>
      <div class="swl-console"></div>
    </div>`
  document.body.append(host)
  const shell = host.querySelector<HTMLElement>('.swl-canvas')!
  const canvas = host.querySelector('canvas')!
  const frames = new Map<number, FrameRequestCallback>()
  let nextId = 0
  vi.stubGlobal('requestAnimationFrame', (callback: FrameRequestCallback) => {
    frames.set(++nextId, callback)
    return nextId
  })
  vi.stubGlobal('cancelAnimationFrame', (id: number) => frames.delete(id))
  const passes = vi.spyOn(ShaderPass.prototype, 'render')
  const renderedPasses = passes.mock.contexts as ShaderPass[]
  const release = createLoadingArtwork(canvas, shell)
  function frame(now: number) {
    const callbacks = [...frames.values()]
    frames.clear()
    for (const callback of callbacks) callback(now)
  }
  try {
    // Feed a 120 Hz display clock through the real WebGL pipeline.
    for (let tick = 0; tick < 120; tick++) frame((tick * 1000) / 120)
    const skyPasses = renderedPasses.filter((pass) => !pass.uniforms.tDiffuse)
    const foregroundPasses = renderedPasses.filter(
      (pass) => pass.uniforms.tDiffuse,
    )
    expect(skyPasses).toHaveLength(1)
    expect(foregroundPasses).toHaveLength(30)
    expect(canvas.width).toBe(Math.round(canvas.clientWidth * devicePixelRatio))
    expect(canvas.height).toBe(
      Math.round(canvas.clientHeight * devicePixelRatio),
    )

    shell.dataset.state = 'loading'
    await new Promise<void>((resolve) => setTimeout(resolve, 0))
    frame(1100)
    expect(
      renderedPasses.filter((pass) => !pass.uniforms.tDiffuse),
    ).toHaveLength(1)

    // Moving the emblem invalidates the dust's hollow around its center.
    shell.style.setProperty('--focus-height', '35%')
    shell.dataset.state = 'ready'
    await new Promise<void>((resolve) => setTimeout(resolve, 0))
    frame(1200)
    expect(
      renderedPasses.filter((pass) => !pass.uniforms.tDiffuse),
    ).toHaveLength(2)
  } finally {
    release()
    host.remove()
  }
  expect(frames.size).toBe(0)
  expect(canvas.hasAttribute('data-renderer')).toBe(false)
})
