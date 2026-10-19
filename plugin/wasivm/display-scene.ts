import type {
  WasiVmScene,
  WasiVmSceneLayer,
  WasiVmSceneMove,
} from '@s4wave/sdk/vm/wasivm.pb.js'

// SceneScreen is the screen a scene covers: its size in units, the screen
// pixels per unit, and the background as RGBA in the high to low bytes.
export interface SceneScreen {
  width: number
  height: number
  scale: number
  background: number
}

// SceneRect is a rectangle of texture pixels.
export interface SceneRect {
  x: number
  y: number
  width: number
  height: number
}

// SceneTexture is the viewer's copy of one layer's texture: premultiplied
// RGBA rows, and the areas changed since a renderer last took them.
export interface SceneTexture {
  width: number
  height: number
  pixels: Uint8Array
  // dirty lists the areas the patches changed since takeDirty, or null when
  // the whole texture changed.
  dirty: SceneRect[] | null
}

// SceneLayer is a layer as the renderer draws it: its texture and where it
// covers the screen, in units.
export interface SceneLayer {
  id: number
  x: number
  y: number
  width: number
  height: number
  opaque: boolean
  texture: SceneTexture
}

// ScenePoint is a position on the screen in units.
export interface ScenePoint {
  x: number
  y: number
}

// Scene is the state a display stream builds: the screen, the layers bottom to
// top, the roots' positions, each layer's texture and the window move in
// progress. apply folds each update into it and point takes the viewer's own
// pointer; the renderer draws it.
export class Scene {
  // screen is the screen's size in units, its scale and its background.
  screen: SceneScreen = {
    width: 0,
    height: 0,
    scale: 1,
    background: 0x000000ff,
  }
  // layers is the layers, bottom to top, as the last update listing them
  // gave them.
  private layers: WasiVmSceneLayer[] = []
  // roots maps each root's id to its position in units.
  private roots = new Map<number, ScenePoint>()
  // move is the window move in progress, or null.
  private move: WasiVmSceneMove | null = null
  // pointer is the viewer's pointer in units, or null before it first moves.
  private pointer: ScenePoint | null = null
  // textures maps each layer's id to its texture.
  private textures = new Map<number, SceneTexture>()

  // apply folds the update u into the scene: the screen, the layers, the
  // roots and the grab it sets, then its patches. A layer missing from a new
  // list loses its texture, and a layer whose texture size changed gets an
  // empty one.
  apply(u: WasiVmScene) {
    // Take the screen and the roots.
    if (u.screen) {
      this.screen = {
        width: u.screen.width ?? 0,
        height: u.screen.height ?? 0,
        scale: u.screen.scale || 1,
        background: u.screen.background ?? 0,
      }
    }
    if (u.roots) {
      this.roots = new Map(
        (u.roots.roots ?? []).map((r) => [
          r.id ?? 0,
          { x: r.x ?? 0, y: r.y ?? 0 },
        ]),
      )
    }
    if (u.grab) {
      this.move = u.grab.move ?? null
    }

    // Take the layers, keeping each texture whose size holds.
    if (u.layers) {
      this.layers = u.layers.layers ?? []
      const textures = new Map<number, SceneTexture>()
      for (const l of this.layers) {
        const width = l.textureWidth ?? 0
        const height = l.textureHeight ?? 0
        const kept = this.textures.get(l.id ?? 0)
        textures.set(
          l.id ?? 0,
          kept && kept.width === width && kept.height === height
            ? kept
            : {
                width,
                height,
                pixels: new Uint8Array(width * height * 4),
                dirty: null,
              },
        )
      }
      this.textures = textures
    }

    // Copy each patch's rows into its texture.
    for (const p of u.patches ?? []) {
      const t = this.textures.get(p.layer ?? 0)
      const r = {
        x: p.x ?? 0,
        y: p.y ?? 0,
        width: p.width ?? 0,
        height: p.height ?? 0,
      }
      const rgba = p.rgba
      if (
        !t ||
        !rgba ||
        r.x + r.width > t.width ||
        r.y + r.height > t.height ||
        rgba.length !== r.width * r.height * 4
      ) {
        continue
      }
      const row = r.width * 4
      for (let y = 0; y < r.height; y++) {
        t.pixels.set(
          rgba.subarray(y * row, (y + 1) * row),
          ((r.y + y) * t.width + r.x) * 4,
        )
      }
      t.dirty?.push(r)
    }
  }

  // point takes the viewer's pointer at the screen pixel x, y, the position
  // it sends the compositor, and reports whether the scene changed: whether
  // a window moves with it.
  point(x: number, y: number): boolean {
    // Convert to units, keeping the pointer on the screen as the compositor
    // does.
    const { width, height, scale } = this.screen
    this.pointer = {
      x: Math.max(Math.min(Math.floor(x / scale), width - 1), 0),
      y: Math.max(Math.min(Math.floor(y / scale), height - 1), 0),
    }
    return this.move !== null
  }

  // rootAt returns the position of the root id: where the move puts it
  // under the viewer's pointer while it moves, else where the last roots
  // put it. The compositor places a moving root by the same rule, so the
  // roots catch up with the prediction rather than correct it.
  private rootAt(id: number): ScenePoint {
    const m = this.move
    const p = this.pointer
    if (m && p && m.root === id) {
      return {
        x: (m.x ?? 0) + p.x - (m.pressX ?? 0),
        y: Math.max((m.y ?? 0) + p.y - (m.pressY ?? 0), m.minY ?? 0),
      }
    }
    return this.roots.get(id) ?? { x: 0, y: 0 }
  }

  // drawList returns the layers bottom to top with their textures, placed at
  // their roots' positions. The array is fresh; the textures are the scene's.
  drawList(): SceneLayer[] {
    const out: SceneLayer[] = []
    for (const l of this.layers) {
      const texture = this.textures.get(l.id ?? 0)
      const root = this.rootAt(l.root ?? 0)
      if (!texture || !texture.width || !texture.height) continue
      out.push({
        id: l.id ?? 0,
        x: root.x + (l.x ?? 0),
        y: root.y + (l.y ?? 0),
        width: l.width ?? 0,
        height: l.height ?? 0,
        opaque: l.opaque ?? false,
        texture,
      })
    }
    return out
  }

  // allDirty marks every texture changed, so a renderer that lost its copies
  // uploads them all again.
  allDirty() {
    for (const t of this.textures.values()) {
      t.dirty = null
    }
  }
}

// SceneRenderer draws scenes onto one canvas.
export interface SceneRenderer {
  // draw sizes the canvas to the screen in pixels, uploads the areas of each
  // texture changed since the last draw, and draws the layers bottom to top
  // over the background.
  draw(scene: Scene): void
  // destroy releases the renderer's textures and its device.
  destroy(): void
}

// takeDirty returns the areas of t changed since the last call, or null when
// all of it changed, and clears them.
export function takeDirty(t: SceneTexture): SceneRect[] | null {
  const dirty = t.dirty
  t.dirty = []
  return dirty
}
