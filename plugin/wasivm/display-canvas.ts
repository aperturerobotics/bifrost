import {
  takeDirty,
  type SceneRect,
  type SceneRenderer,
  type SceneTexture,
} from './display-scene.js'

// openCanvasRenderer draws scenes with Canvas 2D, keeping each texture on an
// offscreen canvas. It returns null when the canvas has no 2D context, as
// when a WebGPU context already took it.
export function openCanvasRenderer(
  canvas: HTMLCanvasElement,
): SceneRenderer | null {
  const ctx = canvas.getContext('2d', { alpha: false })
  if (!ctx) return null
  // copies maps each texture to its offscreen canvas.
  let copies = new Map<SceneTexture, OffscreenCanvas>()

  return {
    draw(scene) {
      // Size the canvas to the screen, which clears it.
      const { width, height, scale, background } = scene.screen
      if (canvas.width !== width * scale || canvas.height !== height * scale) {
        canvas.width = width * scale
        canvas.height = height * scale
      }

      // Bring each texture's copy up to date, dropping copies of textures
      // the scene no longer draws.
      const layers = scene.drawList()
      const kept = new Map<SceneTexture, OffscreenCanvas>()
      for (const l of layers) {
        const t = l.texture
        let copy = copies.get(t)
        if (!copy) {
          copy = new OffscreenCanvas(t.width, t.height)
          t.dirty = null
        }
        kept.set(t, copy)
        const dirty = takeDirty(t)
        const rects = dirty ?? [
          { x: 0, y: 0, width: t.width, height: t.height },
        ]
        const cctx = copy.getContext('2d')
        for (const r of rects) {
          cctx?.putImageData(unpremultiply(t, r), r.x, r.y)
        }
      }
      copies = kept

      // Fill the background, then draw the layers bottom to top.
      ctx.fillStyle = `rgb(${background >>> 24} ${(background >>> 16) & 255} ${(background >>> 8) & 255})`
      ctx.fillRect(0, 0, canvas.width, canvas.height)
      for (const l of layers) {
        const copy = copies.get(l.texture)
        if (!copy) continue
        ctx.drawImage(
          copy,
          l.x * scale,
          l.y * scale,
          l.width * scale,
          l.height * scale,
        )
      }
    },
    destroy() {
      copies.clear()
    },
  }
}

// unpremultiply returns the area r of the texture t as image data, with each
// pixel's color divided by its alpha, which putImageData expects.
function unpremultiply(t: SceneTexture, r: SceneRect): ImageData {
  // Copy the area's rows.
  const out = new ImageData(r.width, r.height)
  const dst = out.data
  for (let y = 0; y < r.height; y++) {
    const from = ((r.y + y) * t.width + r.x) * 4
    dst.set(t.pixels.subarray(from, from + r.width * 4), y * r.width * 4)
  }

  // Divide each partly transparent pixel's color by its alpha.
  for (let i = 0; i < dst.length; i += 4) {
    const a = dst[i + 3] ?? 0
    if (a === 0 || a === 255) continue
    dst[i] = ((dst[i] ?? 0) * 255) / a
    dst[i + 1] = ((dst[i + 1] ?? 0) * 255) / a
    dst[i + 2] = ((dst[i + 2] ?? 0) * 255) / a
  }
  return out
}
