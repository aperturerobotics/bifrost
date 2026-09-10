import { SyncError } from './errors.js'
import { decodeJSON, encodeJSON, type JsonValue } from './json.js'
import type { Operation } from './operation.js'
import type {
  CallOptions,
  Input,
  Output,
  Query,
  RecordEntry,
  Schema,
  Validator,
} from './schema.js'

export interface CollectionAccess<V extends Validator> {
  get(key: string, options?: CallOptions): Promise<Output<V> | undefined>
  scan(
    query?: Query,
    options?: CallOptions,
  ): Promise<readonly RecordEntry<Output<V>>[]>
  put(key: string, value: Input<V>, options?: CallOptions): Promise<void>
  delete(key: string, options?: CallOptions): Promise<void>
}

export interface DatabaseAccess<S extends Schema> {
  collection<K extends keyof S['collections'] & string>(
    name: K,
  ): CollectionAccess<S['collections'][K]>
  mutate<K extends keyof S['mutations'] & string>(
    name: K,
    input: Input<S['mutations'][K]['input']>,
    options?: CallOptions,
  ): Promise<Output<S['mutations'][K]['output']>>
}

export type Dispatch = (
  operation: Operation,
  options?: CallOptions,
) => Promise<JsonValue>

// createAccess shares typed operation construction between local and remote callers.
export function createAccess<S extends Schema>(
  dispatch: Dispatch,
): DatabaseAccess<S> {
  return {
    collection: (name) =>
      ({
        get: async (key, options) => {
          const result = (await dispatch(
            { kind: 'get', collection: name, key },
            options,
          )) as { found: boolean; value?: JsonValue }
          return result.found ? result.value : undefined
        },
        scan: async (query, options) =>
          (await dispatch(
            { kind: 'scan', collection: name, prefix: query?.prefix ?? '' },
            options,
          )) as unknown as readonly RecordEntry<never>[],
        put: async (key, value, options) => {
          await dispatch(
            { kind: 'put', collection: name, key, value: jsonInput(value) },
            options,
          )
        },
        delete: async (key, options) => {
          await dispatch({ kind: 'delete', collection: name, key }, options)
        },
      }) as CollectionAccess<S['collections'][typeof name]>,
    mutate: async (name, input, options) =>
      (await dispatch(
        { kind: 'mutate', name, input: jsonInput(input) },
        options,
      )) as Output<S['mutations'][typeof name]['output']>,
  }
}

// jsonInput freezes the call's meaning before asynchronous work or retry begins.
function jsonInput(value: unknown): JsonValue {
  try {
    return decodeJSON(encodeJSON(value))
  } catch {
    throw new SyncError(
      'VALIDATION',
      'Input must contain only JSON-compatible values',
    )
  }
}
