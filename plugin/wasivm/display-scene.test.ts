import { describe, expect, it } from 'vitest'

import { Scene, takeDirty } from './display-scene.js'

// layer returns a layer of root 1 at the origin with a texture of w by h
// pixels covering w by h units.
function layer(id: number, w: number, h: number) {
  return {
    id,
    root: 1,
    width: w,
    height: h,
    textureWidth: w,
    textureHeight: h,
  }
}

describe('Scene', () => {
  it('places layers at their roots and keeps textures across moves', () => {
    // Show one 2x2 layer, filled by a full patch.
    const scene = new Scene()
    const full = new Uint8Array(16).fill(9)
    scene.apply({
      screen: { width: 10, height: 10, scale: 2, background: 0x112233ff },
      layers: { layers: [layer(5, 2, 2)] },
      roots: { roots: [{ id: 1, x: 3, y: 4 }] },
      patches: [{ layer: 5, x: 0, y: 0, width: 2, height: 2, rgba: full }],
    })
    const [first] = scene.drawList()
    expect(first).toMatchObject({ id: 5, x: 3, y: 4, width: 2, height: 2 })
    expect(scene.screen.scale).toBe(2)
    expect(takeDirty(first!.texture)).toBeNull()

    // Move the root: the texture and its pixels stay, nothing is dirty.
    scene.apply({ roots: { roots: [{ id: 1, x: 7, y: 1 }] } })
    const [moved] = scene.drawList()
    expect(moved).toMatchObject({ x: 7, y: 1 })
    expect(moved!.texture).toBe(first!.texture)
    expect(moved!.texture.pixels).toEqual(full)
    expect(takeDirty(moved!.texture)).toEqual([])

    // Patch one pixel: only its rectangle is dirty.
    scene.apply({
      patches: [
        {
          layer: 5,
          x: 1,
          y: 1,
          width: 1,
          height: 1,
          rgba: Uint8Array.of(1, 2, 3, 4),
        },
      ],
    })
    const t = scene.drawList()[0]!.texture
    expect(Array.from(t.pixels.subarray(12))).toEqual([1, 2, 3, 4])
    expect(takeDirty(t)).toEqual([{ x: 1, y: 1, width: 1, height: 1 }])
  })

  it('moves a grabbed root with the local pointer until the grab ends', () => {
    // Show two roots at scale 2, and a move of root 1 pressed at 10,10 when
    // it sat at 4,6, whose top stays at 5 or below.
    const scene = new Scene()
    const one = { ...layer(1, 2, 2), root: 1 }
    const two = { ...layer(2, 2, 2), root: 2 }
    scene.apply({
      screen: { width: 40, height: 30, scale: 2 },
      layers: { layers: [one, two] },
      roots: {
        roots: [
          { id: 1, x: 4, y: 6 },
          { id: 2, x: 20, y: 20 },
        ],
      },
    })
    const at = (id: number) => scene.drawList().find((l) => l.id === id)
    expect(scene.point(0, 0)).toBe(false)
    scene.apply({
      grab: { move: { root: 1, pressX: 10, pressY: 10, x: 4, y: 6, minY: 5 } },
    })

    // Moving the pointer, in screen pixels, moves root 1 by its travel in
    // units and leaves root 2.
    expect(scene.point(31, 25)).toBe(true)
    expect(at(1)).toMatchObject({ x: 9, y: 8 })
    expect(at(2)).toMatchObject({ x: 20, y: 20 })

    // Its top stops at the least, and the pointer stays on the screen.
    scene.point(-50, 0)
    expect(at(1)).toMatchObject({ x: -6, y: 5 })

    // Roots that trail the pointer do not pull it back.
    scene.apply({ roots: { roots: [{ id: 1, x: 7, y: 7 }] } })
    expect(at(1)).toMatchObject({ x: -6, y: 5 })

    // When the grab ends, the roots place it again.
    scene.apply({ grab: {}, roots: { roots: [{ id: 1, x: -6, y: 5 }] } })
    expect(scene.point(100, 100)).toBe(false)
    expect(at(1)).toMatchObject({ x: -6, y: 5 })
  })

  it('drops gone layers and replaces resized textures', () => {
    // Show two layers, then list only the second, resized.
    const scene = new Scene()
    scene.apply({
      layers: { layers: [layer(1, 2, 2), layer(2, 2, 2)] },
      roots: { roots: [{ id: 1 }] },
    })
    const before = scene.drawList()[1]!.texture
    takeDirty(before)
    scene.apply({ layers: { layers: [layer(2, 3, 1)] } })

    // The second layer remains with a new, wholly dirty texture.
    const after = scene.drawList()
    expect(after.map((l) => l.id)).toEqual([2])
    expect(after[0]!.texture).not.toBe(before)
    expect(after[0]!.texture).toMatchObject({
      width: 3,
      height: 1,
      dirty: null,
    })
  })

  it('ignores a patch outside its texture', () => {
    const scene = new Scene()
    scene.apply({
      layers: { layers: [layer(1, 2, 2)] },
      roots: { roots: [{ id: 1 }] },
      patches: [
        {
          layer: 1,
          x: 1,
          y: 1,
          width: 2,
          height: 1,
          rgba: new Uint8Array(8).fill(7),
        },
      ],
    })
    expect(scene.drawList()[0]!.texture.pixels.every((b) => b === 0)).toBe(true)
  })
})
