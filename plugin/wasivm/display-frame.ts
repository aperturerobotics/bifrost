import {
  WasiVmSceneFrameButtonKind,
  type WasiVmSceneFrame,
} from '@s4wave/sdk/vm/wasivm.pb.js'

import type { FramePainter, SceneTexture } from './display-scene.js'

// FrameTheme is the app's colors and font a title bar is painted in.
interface FrameTheme {
  // active and inactive are the title bar's background for the active
  // window and the others.
  active: string
  inactive: string
  // text and mutedText are the title and icons' color on each.
  text: string
  mutedText: string
  // border is the color of the edge around each frame.
  border: string
  // font is the font family.
  font: string
}

// framePainter returns a painter of frames in the app's tokens, as the
// element el resolves them when each texture is painted.
export function framePainter(el: Element): FramePainter {
  return {
    title: (f, scale) => paintFrame(readTheme(el), f, scale),
    edge: () => paintEdge(readTheme(el)),
  }
}

// readTheme reads the title bar's colors and font from the tokens el sees.
function readTheme(el: Element): FrameTheme {
  const style = getComputedStyle(el)
  const token = (name: string, fallback: string) =>
    style.getPropertyValue(name).trim() || fallback
  return {
    active: token('--color-background-panel', '#3a3a3a'),
    inactive: token('--color-background-card', '#2a2a2a'),
    text: token('--color-foreground', '#ffffff'),
    mutedText: token('--color-muted-foreground', '#aaaaaa'),
    border: token('--color-border', '#555555'),
    font: style.fontFamily || 'sans-serif',
  }
}

// paintFrame paints the title bar of the frame f at scale pixels per unit:
// the background, the title at the left cut short of the buttons, and each
// button's icon.
function paintFrame(
  theme: FrameTheme,
  f: WasiVmSceneFrame,
  scale: number,
): SceneTexture {
  // Size an opaque canvas to the title bar and draw in units.
  const width = (f.width ?? 0) * scale
  const height = (f.titleHeight ?? 0) * scale
  const canvas = new OffscreenCanvas(Math.max(width, 1), Math.max(height, 1))
  const ctx = canvas.getContext('2d', { alpha: false })
  if (!ctx)
    return { width: 0, height: 0, pixels: new Uint8Array(), dirty: null }
  ctx.scale(scale, scale)

  // Fill the background in the window's state.
  ctx.fillStyle = f.active ? theme.active : theme.inactive
  ctx.fillRect(0, 0, f.width ?? 0, f.titleHeight ?? 0)
  const ink = f.active ? theme.text : theme.mutedText

  // Write the title, ending with an ellipsis where it would reach the
  // leftmost button.
  const left = Math.min(
    ...(f.buttons ?? []).map((b) => (b.x ?? 0) - (f.x ?? 0)),
    f.width ?? 0,
  )
  ctx.font = `500 13px ${theme.font}`
  ctx.fillStyle = ink
  ctx.textBaseline = 'middle'
  ctx.fillText(
    fitText(ctx, f.title ?? '', left - 20),
    10,
    (f.titleHeight ?? 0) / 2,
  )

  // Draw each button's icon, centered in its square.
  ctx.strokeStyle = ink
  ctx.lineWidth = 1.25
  for (const b of f.buttons ?? []) {
    const cx = (b.x ?? 0) - (f.x ?? 0) + (b.width ?? 0) / 2
    const cy = (b.y ?? 0) - (f.y ?? 0) + (b.height ?? 0) / 2
    drawIcon(ctx, b.kind, f.maximized ?? false, cx, cy)
  }

  // Take the pixels, which are opaque, so already premultiplied.
  const image = ctx.getImageData(0, 0, canvas.width, canvas.height)
  return {
    width: canvas.width,
    height: canvas.height,
    pixels: new Uint8Array(image.data.buffer),
    dirty: null,
  }
}

// paintEdge paints the one opaque pixel of a frame's edge.
function paintEdge(theme: FrameTheme): SceneTexture {
  // Fill an opaque pixel in the border color and take it.
  const canvas = new OffscreenCanvas(1, 1)
  const ctx = canvas.getContext('2d', { alpha: false })
  if (!ctx)
    return { width: 0, height: 0, pixels: new Uint8Array(), dirty: null }
  ctx.fillStyle = theme.border
  ctx.fillRect(0, 0, 1, 1)
  const image = ctx.getImageData(0, 0, 1, 1)
  return {
    width: 1,
    height: 1,
    pixels: new Uint8Array(image.data.buffer),
    dirty: null,
  }
}

// fitText returns text, or its longest start followed by an ellipsis, that
// fits in width units.
function fitText(
  ctx: OffscreenCanvasRenderingContext2D,
  text: string,
  width: number,
): string {
  if (ctx.measureText(text).width <= width) return text
  let n = text.length
  while (n > 0 && ctx.measureText(text.slice(0, n) + '…').width > width) n--
  return n > 0 ? text.slice(0, n) + '…' : ''
}

// drawIcon strokes the icon of a button of kind centered at cx, cy: a cross
// for close, a square for maximize, or two offset squares to restore a
// maximized window.
function drawIcon(
  ctx: OffscreenCanvasRenderingContext2D,
  kind: WasiVmSceneFrameButtonKind | undefined,
  maximized: boolean,
  cx: number,
  cy: number,
) {
  const r = 4.5
  ctx.beginPath()
  switch (kind) {
    case WasiVmSceneFrameButtonKind.CLOSE:
      ctx.moveTo(cx - r, cy - r)
      ctx.lineTo(cx + r, cy + r)
      ctx.moveTo(cx + r, cy - r)
      ctx.lineTo(cx - r, cy + r)
      break
    case WasiVmSceneFrameButtonKind.MAXIMIZE:
      if (maximized) {
        ctx.rect(cx - r, cy - r + 2.5, 2 * r - 2.5, 2 * r - 2.5)
        ctx.moveTo(cx - r + 2.5, cy - r + 2.5)
        ctx.lineTo(cx - r + 2.5, cy - r)
        ctx.lineTo(cx + r, cy - r)
        ctx.lineTo(cx + r, cy + r - 2.5)
        ctx.lineTo(cx + r - 2.5, cy + r - 2.5)
      } else {
        ctx.rect(cx - r, cy - r, 2 * r, 2 * r)
      }
      break
  }
  ctx.stroke()
}
