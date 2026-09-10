import type { StandardSchemaV1 } from '@standard-schema/spec'

import { SyncError } from './errors.js'
import { canonicalJSON, type JsonValue } from './json.js'

export type Validator = StandardSchemaV1
export type Input<S extends Validator> = StandardSchemaV1.InferInput<S>
export type Output<S extends Validator> = StandardSchemaV1.InferOutput<S>

export interface MutationDefinition {
  readonly input: Validator
  readonly output: Validator
}

export interface Schema {
  readonly id: string
  readonly version: number
  readonly collections: Readonly<Record<string, Validator>>
  readonly mutations: Readonly<Record<string, MutationDefinition>>
}

// defineSchema preserves validators' inferred types in a browser-safe contract.
export function defineSchema<const S extends Schema>(schema: S): S {
  if (
    !schema.id ||
    !Number.isSafeInteger(schema.version) ||
    schema.version < 1 ||
    schema.version > 0xffffffff
  ) {
    throw new SyncError(
      'VALIDATION',
      'Application ID and positive version are required',
    )
  }
  return schema
}

// validate persists the validator's output, and never transforms a stored read.
export async function validate<V extends Validator>(
  validator: V,
  value: unknown,
  field: string,
): Promise<Output<V> & JsonValue> {
  try {
    canonicalJSON(value)
    const result = await validator['~standard'].validate(value)
    if (result.issues)
      throw new SyncError('VALIDATION', `${field} does not match its schema`)
    canonicalJSON(result.value)
    return result.value as Output<V> & JsonValue
  } catch (error) {
    if (error instanceof SyncError) throw error
    throw new SyncError(
      'VALIDATION',
      `${field} must be a valid JSON value matching its schema`,
    )
  }
}

export interface CallOptions {
  readonly signal?: AbortSignal
  readonly requestId?: string
}

export interface RecordEntry<T> {
  readonly key: string
  readonly value: T
}

export interface Query {
  readonly prefix?: string
}

export interface TransactionCollection<V extends Validator> {
  get(key: string): Promise<Output<V> | undefined>
  scan(query?: Query): Promise<readonly RecordEntry<Output<V>>[]>
  put(key: string, value: Input<V>): Promise<void>
  delete(key: string): Promise<void>
}

export interface Transaction<S extends Schema> {
  collection<K extends keyof S['collections'] & string>(
    name: K,
  ): TransactionCollection<S['collections'][K]>
}

// Principal is supplied by the application's authentication owner.
export interface Principal {
  readonly subject: string
  readonly scope: string
  readonly expiresAt?: number
  readonly signal?: AbortSignal
}

export interface MutationContext<
  S extends Schema,
  P extends Principal,
> extends Transaction<S> {
  readonly principal: P
  readonly scope: string
  readonly signal: AbortSignal
}

export type MutationHandlers<S extends Schema, P extends Principal> = {
  [K in keyof S['mutations']]: (
    context: MutationContext<S, P>,
    input: Output<S['mutations'][K]['input']>,
  ) => Promise<Input<S['mutations'][K]['output']>>
}
