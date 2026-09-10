import type { JsonValue } from './json.js'

// Operation is the validated JSON request retained for stable acceptance replay.
export type Operation =
  | { kind: 'get'; collection: string; key: string }
  | { kind: 'scan'; collection: string; prefix: string }
  | { kind: 'put'; collection: string; key: string; value: JsonValue }
  | { kind: 'delete'; collection: string; key: string }
  | { kind: 'mutate'; name: string; input: JsonValue }
