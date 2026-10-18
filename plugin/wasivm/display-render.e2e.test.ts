import { describe, expect, it } from 'vitest'

import { openCanvasRenderer } from './display-canvas.js'
import { openGpuRenderer } from './display-gpu.js'
import { Scene, type SceneRenderer } from './display-scene.js'

// testScene returns a 4 by 4 scene at scale 1 over a green background: an
// opaque red layer covering the top left 2 by 2 units, and a half transparent
// blue layer, premultiplied, covering the right column.
function testScene(): Scene {
  // Describe the screen and the two layers of one root.
  const scene = new Scene()
  scene.apply({
    screen: { width: 4, height: 4, scale: 1, background: 0x00ff00ff },
    roots: { roots: [{ id: 1 }] },
    layers: {
      layers: [
        {
          id: 1,
          root: 1,
          width: 2,
          height: 2,
          textureWidth: 2,
          textureHeight: 2,
          opaque: true,
        },
        {
          id: 2,
          root: 1,
          x: 3,
          width: 1,
          height: 4,
          textureWidth: 1,
          textureHeight: 4,
        },
      ],
    },
    patches: [
      {
        layer: 1,
        width: 2,
        height: 2,
        rgba: fill(4, [255, 0, 0, 255]),
      },
      {
        layer: 2,
        width: 1,
        height: 4,
        rgba: fill(4, [0, 0, 128, 128]),
      },
    ],
  })
  return scene
}

// fill returns n pixels of the RGBA color c.
function fill(n: number, c: number[]): Uint8Array {
  const out = new Uint8Array(n * 4)
  for (let i = 0; i < n; i++) out.set(c, i * 4)
  return out
}

// readPixels draws scene with renderer on canvas and returns the canvas's
// pixels, copied in the same task so a WebGPU canvas still holds the frame.
function readPixels(
  renderer: SceneRenderer,
  canvas: HTMLCanvasElement,
  scene: Scene,
): Uint8ClampedArray {
  // Draw, then copy the frame into a readable 2D canvas.
  renderer.draw(scene)
  const copy = document.createElement('canvas')
  copy.width = canvas.width
  copy.height = canvas.height
  const ctx = copy.getContext('2d', { willReadFrequently: true })!
  ctx.drawImage(canvas, 0, 0)
  return ctx.getImageData(0, 0, copy.width, copy.height).data
}

// expectScene checks the pixels of the 4 by 4 test scene: red under the
// opaque layer, half blue over green under the other, green elsewhere.
function expectScene(pix: Uint8ClampedArray) {
  // pixel returns the RGB of the pixel at x, y.
  const pixel = (x: number, y: number) =>
    Array.from(pix.subarray((y * 4 + x) * 4, (y * 4 + x) * 4 + 3))
  const near = (got: number[], want: number[]) =>
    got.every((v, i) => Math.abs(v - want[i]!) <= 2)

  // Check each layer's pixels and the background between them.
  expect(pixel(0, 0)).toEqual([255, 0, 0])
  expect(pixel(1, 1)).toEqual([255, 0, 0])
  expect(pixel(2, 2)).toEqual([0, 255, 0])
  expect(near(pixel(3, 0), [0, 127, 128]), `${pixel(3, 0)}`).toBe(true)
  expect(near(pixel(3, 3), [0, 127, 128]), `${pixel(3, 3)}`).toBe(true)
}

describe('display renderers', () => {
  it('draws a scene with Canvas 2D', () => {
    // Canvas 2D is the fallback every browser supports.
    const canvas = document.createElement('canvas')
    const renderer = openCanvasRenderer(canvas)
    expect(renderer).not.toBeNull()
    expectScene(readPixels(renderer!, canvas, testScene()))
    renderer!.destroy()
  })

  it('draws a scene with WebGPU', async (ctx) => {
    // Skip where the browser offers no WebGPU adapter, as headless Chromium
    // does by default; the viewer then falls back to Canvas 2D, checked
    // above. Run it on a hardware adapter: SwiftShader draws, but its canvas
    // reads back as zeros.
    const canvas = document.createElement('canvas')
    const renderer =
      'gpu' in navigator ? await openGpuRenderer(canvas, () => {}) : null
    if (!renderer) {
      ctx.skip()
      return
    }
    expectScene(readPixels(renderer, canvas, testScene()))
    renderer.destroy()
  })
})
