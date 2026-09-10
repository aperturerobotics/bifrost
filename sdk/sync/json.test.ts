import { describe, expect, test } from 'vitest'
import {
  canonicalJSON,
  decodeJSON,
  encodeJSON,
  type JsonValue,
} from './json.js'

describe('canonicalJSON', () => {
  test('null round-trips and stays distinct from absent', () => {
    expect(canonicalJSON(null)).toBe('null')
    expect(decodeJSON(encodeJSON(null))).toBe(null)
  })

  test('scalars round-trip', () => {
    for (const v of [true, false, 0, 1.5, -3, 1e21, '', 'a\u0000b']) {
      expect(decodeJSON(encodeJSON(v))).toEqual(v)
    }
    // Negative zero normalizes to zero in canonical form.
    expect(canonicalJSON(-0)).toBe('0')
  })

  test('object keys sort lexically', () => {
    expect(canonicalJSON({ b: 1, a: { d: 2, c: 3 }, '': 4 })).toBe(
      '{"":{"a":'.slice(0, 0) + '{"":4,"a":{"c":3,"d":2},"b":1}',
    )
  })

  test('nested arrays and objects round-trip', () => {
    const v = { list: [1, 'two', [true, null]], empty: {}, arr: [] }
    expect(decodeJSON(encodeJSON(v))).toEqual(v)
  })

  test('unicode strings round-trip and escape control characters', () => {
    const v = { key: 'héllo \u2028 \ud83d\ude00 \u0001' }
    const text = canonicalJSON(v)
    expect(decodeJSON(encodeJSON(v))).toEqual(v)
    expect(text).not.toContain('\u0001')
  })

  test('repeated references without a cycle are allowed', () => {
    const shared = { x: 1 }
    const v = { a: shared, b: shared }
    expect(canonicalJSON(v)).toBe('{"a":{"x":1},"b":{"x":1}}')
    expect(decodeJSON(encodeJSON(v))).toEqual(v)
  })

  test('rejects cycles', () => {
    const v: Record<string, unknown> = { name: 'root' }
    v.self = v
    expect(() => canonicalJSON(v)).toThrow(TypeError)
    const arr: unknown[] = [1]
    arr.push(arr)
    expect(() => canonicalJSON(arr)).toThrow(TypeError)
  })

  test('rejects lossy and non-JSON values', () => {
    class Cls {
      x = 1
    }
    const d = new Date(0)
    for (const v of [
      undefined,
      10n,
      Symbol('s'),
      () => {},
      NaN,
      Infinity,
      -Infinity,
      new Cls(),
      d,
      new Map(),
      { a: undefined },
      { a: 10n },
      { a: NaN },
      [undefined],
      // eslint-disable-next-line no-sparse-arrays -- Holes must fail JSON validation.
      [1, , 3],
      { nested: { deep: new Date(0) } },
    ]) {
      expect(() => canonicalJSON(v)).toThrow(TypeError)
    }
  })

  test('error messages carry a field path and reason, not the value', () => {
    try {
      canonicalJSON({ a: { b: [1, 10n] } })
      expect.unreachable()
    } catch (e) {
      expect(e).toBeInstanceOf(TypeError)
      expect((e as Error).message).toContain('$.a.b[1]')
      expect((e as Error).message).not.toContain('10')
    }
  })

  test('rejects accessor properties without calling getters', () => {
    let called = false
    const v = {
      get bad() {
        called = true
        return 1
      },
    }
    expect(() => canonicalJSON(v)).toThrow(TypeError)
    expect(called).toBe(false)
  })

  test('rejects enumerable symbol keys', () => {
    const v: Record<PropertyKey, unknown> = { a: 1 }
    v[Symbol('s')] = 2
    expect(() => canonicalJSON(v)).toThrow(TypeError)
  })

  test('rejects extra enumerable array properties', () => {
    const v = Object.assign([1, 2], { extra: 3 })
    expect(() => canonicalJSON(v)).toThrow(TypeError)
  })

  test('class instances are rejected', () => {
    class Vec {
      x = 1
    }
    expect(() => canonicalJSON(new Vec())).toThrow(/not a plain object/)
  })

  test('null-prototype objects are accepted and round-trip', () => {
    const v = Object.create(null) as Record<string, unknown>
    v.a = 1
    v.b = { c: [true] }
    expect(canonicalJSON(v)).toBe('{"a":1,"b":{"c":[true]}}')
    expect(decodeJSON(encodeJSON(v))).toEqual({ a: 1, b: { c: [true] } })
  })

  test('decodeJSON preserves own reserved keys as data properties', () => {
    const text = '{"__proto__":1,"constructor":2,"prototype":3}'
    const out = decodeJSON(new TextEncoder().encode(text)) as Record<
      string,
      JsonValue
    >
    expect(Object.keys(out).sort()).toEqual([
      '__proto__',
      'constructor',
      'prototype',
    ])
    expect(Object.getPrototypeOf(out)).toBe(Object.prototype)
    for (const [key, want] of [
      ['__proto__', 1],
      ['constructor', 2],
      ['prototype', 3],
    ] as const) {
      expect(Object.hasOwn(out, key)).toBe(true)
      expect(out[key]).toBe(want)
    }
    // Round-trip through the codec keeps the keys.
    const back = decodeJSON(encodeJSON(out)) as Record<string, JsonValue>
    for (const key of ['__proto__', 'constructor', 'prototype'] as const) {
      expect(Object.hasOwn(back, key)).toBe(true)
      expect(back[key]).toBe(out[key])
    }
  })

  test('decodeJSON rejects invalid UTF-8', () => {
    expect(() => decodeJSON(new Uint8Array([0xff, 0xfe]))).toThrow(TypeError)
  })

  test('decodeJSON rejects non-finite numbers and invalid JSON', () => {
    expect(() => decodeJSON(new TextEncoder().encode('1e999'))).toThrow(
      TypeError,
    )
    expect(() => decodeJSON(new TextEncoder().encode('{'))).toThrow(TypeError)
  })
})
