import { createHash, randomUUID } from 'node:crypto'

import {
  KvScanLimitError,
  KvStore,
  type KvTransaction,
} from '../../sdk/kv/kv.js'
import {
  KvObjectTypeError,
  openWorldKvStore,
} from '../../sdk/kv/world/store.js'
import { Engine } from '../../sdk/world/engine.js'
import type { Tx } from '../../sdk/world/tx.js'
import { getObjectType } from '../../sdk/world/types/types.js'
import { SyncError, publicError } from '../../sdk/sync/errors.js'
import type { Operation } from '../../sdk/sync/operation.js'
import {
  canonicalJSON,
  decodeJSON,
  encodeJSON,
  type JsonValue,
} from '../../sdk/sync/json.js'
import {
  collectionKey,
  encodeRecordKey,
  metadataKey,
  receiptKey,
} from '../../sdk/sync/keys.js'
import {
  validate,
  defineSchema,
  type CallOptions,
  type MutationHandlers,
  type Principal,
  type Schema,
  type Transaction,
  type TransactionCollection,
  type RecordEntry,
} from '../../sdk/sync/schema.js'

export interface Access<P extends Principal> {
  readonly principal: P
  readonly scope: string
  readonly collection: string
  readonly action: 'read' | 'write'
}

export interface ApplicationOptions<S extends Schema, P extends Principal> {
  schema: S
  engine: Engine
  authorize(access: Access<P>): boolean | Promise<boolean>
  mutations: MutationHandlers<S, P>
  limits?: {
    maxRecords?: number
    maxSnapshotBytes?: number
    maxRecordBytes?: number
  }
}

interface Receipt {
  fingerprint: string
  result: JsonValue
  accesses: { collection: string; action: 'read' | 'write' }[]
}

interface StoredVersion {
  application: string
  version: number
}

export interface Migration<S extends Schema> {
  from: number
  run(context: {
    scopes(): Promise<readonly string[]>
    scope(scope: string): Transaction<S>
    signal: AbortSignal
  }): Promise<void>
}

const encoder = new TextEncoder()
const decoder = new TextDecoder('utf-8', { fatal: true, ignoreBOM: true })
const versionKey = encoder.encode('version')

// Application owns admission, policy, transactions, receipts, and its engine ref.
// All application writes, including metadata and receipts, commit through one Tx.
export class Application<S extends Schema, P extends Principal> {
  readonly schema: S
  readonly signal: AbortSignal
  readonly limits: {
    maxRecords: number
    maxSnapshotBytes: number
    maxRecordBytes: number
  }
  private readonly engine: Engine
  private readonly controller = new AbortController()
  private readonly active = new Set<Promise<unknown>>()
  private closing?: Promise<void>

  constructor(private readonly options: ApplicationOptions<S, P>) {
    this.schema = defineSchema(options.schema)
    try {
      receiptKey(this.schema.id, 'schema')
      for (const [name, validator] of Object.entries(this.schema.collections)) {
        collectionKey(this.schema.id, 'schema', name)
        if (validator['~standard']?.version !== 1)
          throw new Error('invalid validator')
      }
      for (const [name, mutation] of Object.entries(this.schema.mutations)) {
        if (
          !Object.hasOwn(options.mutations, name) ||
          mutation.input['~standard']?.version !== 1 ||
          mutation.output['~standard']?.version !== 1
        )
          throw new Error('invalid mutation')
      }
    } catch {
      throw new SyncError(
        'VALIDATION',
        'Schema requires valid names, Standard Schema v1 validators, and every mutation handler',
      )
    }
    this.signal = this.controller.signal
    this.limits = {
      maxRecords: options.limits?.maxRecords ?? 10_000,
      maxSnapshotBytes: options.limits?.maxSnapshotBytes ?? 8 * 1024 * 1024,
      maxRecordBytes: options.limits?.maxRecordBytes ?? 256 * 1024,
    }
    for (const value of Object.values(this.limits)) {
      if (!Number.isSafeInteger(value) || value < 1)
        throw new SyncError('VALIDATION', 'Limits must be positive integers')
    }
    this.engine = new Engine(
      options.engine.resourceRef.createRef(options.engine.id),
    )
  }

  // initialize runs explicit maintenance before any application caller is admitted.
  async initialize(migration?: Migration<S>): Promise<void> {
    const tx = await this.engine.newTransaction(true, this.signal)
    const unit = new Unit(tx, true, this.signal)
    try {
      const metadata = await unit.store(metadataKey)
      if (!metadata) throw new Error('missing metadata store')
      const stored = await metadata.get(versionKey, this.signal)
      const previous = stored.found
        ? (decodeJSON(stored.data) as unknown as StoredVersion)
        : undefined
      if (previous && previous.application !== this.schema.id) {
        throw new SyncError(
          'SCHEMA_MISMATCH',
          'This dataset belongs to a different application',
        )
      }
      if (previous && previous.version !== this.schema.version) {
        if (!migration || migration.from !== previous.version) {
          throw new SyncError(
            'SCHEMA_MISMATCH',
            'Stored version differs; supply an explicit maintenance migration',
          )
        }
        await migration.run({
          signal: this.signal,
          scopes: async () => {
            const entries = await metadata.scanRecords(
              encoder.encode('scope/'),
              {
                maxRecords: this.limits.maxRecords,
                maxBytes: this.limits.maxSnapshotBytes,
              },
              this.signal,
            )
            return entries.map((entry) => decodeJSON(entry.value) as string)
          },
          scope: (scope) =>
            this.collections(unit, { subject: 'admin', scope } as P, true, []),
        })
      }
      await metadata.set(
        versionKey,
        encodeJSON({
          application: this.schema.id,
          version: this.schema.version,
        }),
        this.signal,
      )
      await unit.commit()
      if (!(await this.engine.sync()))
        throw new SyncError('STORAGE', 'Durable storage is unavailable')
    } finally {
      await unit.release()
    }
  }

  async authorize(
    principal: P,
    collection: string,
    action: 'read' | 'write',
    admin = false,
  ): Promise<void> {
    this.checkPrincipal(principal)
    if (!Object.hasOwn(this.schema.collections, collection))
      throw new SyncError('MISSING_COLLECTION', 'Collection is not declared')
    if (
      !admin &&
      !(await this.options.authorize({
        principal,
        scope: principal.scope,
        collection,
        action,
      }))
    ) {
      throw new SyncError('DENIED', `Collection ${action} is not authorized`)
    }
  }

  // Opening a capability grants no operation; every call checks its own policy.
  async openCollection(principal: P, collection: string): Promise<void> {
    try {
      await this.authorize(principal, collection, 'read')
    } catch (error) {
      if (!(error instanceof SyncError) || error.code !== 'DENIED') throw error
      await this.authorize(principal, collection, 'write')
    }
  }

  execute(
    principal: P,
    operation: Operation,
    options: CallOptions = {},
    admin = false,
  ): Promise<JsonValue> {
    if (this.closing)
      return Promise.reject(new SyncError('CLOSED', 'Server is closed'))
    const work = this.run(principal, operation, options, admin)
    this.active.add(work)
    void work.finally(() => this.active.delete(work)).catch(() => {})
    return work
  }

  close(): Promise<void> {
    this.closing ??= (async () => {
      this.controller.abort(new SyncError('CLOSED', 'Server is closing'))
      await Promise.allSettled(this.active)
      this.engine.release()
    })()
    return this.closing
  }

  // watch follows the owning KV producer and rechecks policy before every delivery.
  async *watch(
    principal: P,
    collection: string,
    prefix: string,
    signal: AbortSignal,
  ): AsyncIterable<readonly RecordEntry<JsonValue>[]> {
    if (this.closing) throw new SyncError('CLOSED', 'Server is closed')
    const lifetime = AbortSignal.any([
      this.signal,
      signal,
      ...(principal.signal ? [principal.signal] : []),
    ])
    const done = Promise.withResolvers<void>()
    this.active.add(done.promise)
    let store: KvStore | undefined
    try {
      const key = collectionKey(this.schema.id, principal.scope, collection)
      const bytes = prefix ? this.recordKey(prefix) : new Uint8Array()
      let emptyDelivered = false
      for (;;) {
        lifetime.throwIfAborted()
        // Each iteration follows a later World revision and rechecks its authority.
        // eslint-disable-next-line react-doctor/async-await-in-loop
        await this.authorize(principal, collection, 'read')
        const sequence = await this.engine.getSeqno(lifetime)
        const read = await this.engine.newTransaction(false, lifetime)
        let exists = false
        try {
          const object = await read.getObject(key, lifetime)
          exists = object !== null
          object?.release()
          if (
            exists &&
            (await getObjectType(read, key, lifetime)) !== 'kv/store'
          )
            throw new KvObjectTypeError()
        } finally {
          try {
            await read.discard()
          } finally {
            read.release()
          }
        }
        if (exists) break
        if (!emptyDelivered) {
          yield []
          emptyDelivered = true
        }
        await this.engine.waitSeqno((sequence.seqno ?? 0n) + 1n, lifetime)
      }
      const access = await this.engine.accessTypedObject(key, lifetime)
      store = this.engine.resourceRef.createResource(access.resourceId, KvStore)
      for await (const entries of store.watchRecords(
        bytes,
        {
          maxRecords: this.limits.maxRecords,
          maxBytes: this.limits.maxSnapshotBytes,
        },
        lifetime,
      )) {
        await this.authorize(principal, collection, 'read')
        yield entries.map((entry) => ({
          key: decoder.decode(entry.key),
          value: decodeJSON(entry.value),
        }))
      }
    } catch (error) {
      if (error instanceof KvScanLimitError)
        throw new SyncError('QUERY_LIMIT', error.message)
      if (!lifetime.aborted) throw publicError(error)
    } finally {
      store?.release()
      done.resolve()
      this.active.delete(done.promise)
    }
  }

  checkPrincipal(principal: P): void {
    if (
      !principal.subject ||
      !principal.scope ||
      principal.signal?.aborted ||
      (principal.expiresAt !== undefined &&
        (!Number.isFinite(principal.expiresAt) ||
          principal.expiresAt <= Date.now()))
    ) {
      throw new SyncError('AUTHENTICATION', 'Authentication is no longer valid')
    }
    try {
      receiptKey(this.schema.id, principal.scope)
      if (encodeRecordKey(principal.subject).length > 256)
        throw new Error('invalid subject')
    } catch {
      throw new SyncError(
        'AUTHENTICATION',
        'Principal subject and scope must be valid identities within 256 UTF-8 bytes',
      )
    }
  }

  private async run(
    principal: P,
    operation: Operation,
    options: CallOptions,
    admin: boolean,
  ): Promise<JsonValue> {
    let unit: Unit | undefined
    let committing = false
    const write = operation.kind !== 'get' && operation.kind !== 'scan'
    const requestId = options.requestId ?? randomUUID()
    const signal = AbortSignal.any([
      this.signal,
      ...[principal.signal, options.signal].filter((s): s is AbortSignal =>
        Boolean(s),
      ),
    ])
    try {
      this.checkPrincipal(principal)
      signal.throwIfAborted()
      if (!requestId || requestId.length > 128)
        throw new SyncError(
          'VALIDATION',
          'Request ID must contain 1 to 128 characters',
        )
      const fingerprint = createHash('sha256')
        .update(canonicalJSON({ version: this.schema.version, operation }))
        .digest('hex')
      const tx = await this.engine.newTransaction(write, signal)
      unit = new Unit(tx, write, signal)
      const accesses: Receipt['accesses'] = []
      const collections = this.collections(unit, principal, admin, accesses)
      if ('collection' in operation)
        await this.authorize(
          principal,
          operation.collection,
          write ? 'write' : 'read',
          admin,
        )
      const receipts = write
        ? await unit.store(receiptKey(this.schema.id, principal.scope))
        : undefined
      const receiptId = encoder.encode(
        canonicalJSON([principal.subject, requestId]),
      )
      if (receipts) {
        const stored = await receipts.get(receiptId, signal)
        if (stored.found) {
          const receipt = decodeJSON(stored.data) as unknown as Receipt
          if (receipt.fingerprint !== fingerprint)
            throw new SyncError(
              'CONFLICT',
              'Request ID was already used for different input',
              requestId,
            )
          await Promise.all(
            receipt.accesses.map((access) =>
              this.authorize(
                principal,
                access.collection,
                access.action,
                admin,
              ),
            ),
          )
          // A previous response may have failed while its durability fence was pending.
          await unit.release()
          unit = undefined
          if (!(await this.engine.sync()))
            throw new SyncError('STORAGE', 'Durable storage is unavailable')
          return receipt.result
        }
      }
      let result: JsonValue
      if (operation.kind === 'mutate') {
        const definition = Object.hasOwn(this.schema.mutations, operation.name)
          ? this.schema.mutations[operation.name]
          : undefined
        const handler = Object.hasOwn(this.options.mutations, operation.name)
          ? this.options.mutations[operation.name]
          : undefined
        if (!definition || !handler)
          throw new SyncError(
            'VALIDATION',
            'Mutation is not declared and implemented',
          )
        const input = await validate(
          definition.input,
          operation.input,
          'Mutation input',
        )
        const output = await handler(
          { ...collections, principal, scope: principal.scope, signal },
          input,
        )
        result = await validate(definition.output, output, 'Mutation output')
      } else {
        const collection = collections.collection(operation.collection)
        switch (operation.kind) {
          case 'get': {
            const value = await collection.get(operation.key)
            result =
              value === undefined
                ? { found: false }
                : { found: true, value: value as JsonValue }
            break
          }
          case 'scan':
            result = (await collection.scan({
              prefix: operation.prefix,
            })) as unknown as JsonValue
            break
          case 'put':
            await collection.put(operation.key, operation.value)
            result = null
            break
          case 'delete':
            await collection.delete(operation.key)
            result = null
            break
        }
      }
      if (write) {
        const metadata = await unit.store(metadataKey)
        await metadata!.set(
          encoder.encode(`scope/${canonicalJSON(principal.scope)}`),
          encodeJSON(principal.scope),
          signal,
        )
        await receipts!.set(
          receiptId,
          encodeJSON({ fingerprint, result, accesses }),
          signal,
        )
        this.checkPrincipal(principal)
        signal.throwIfAborted()
        await unit.commit(() => {
          committing = true
        })
        if (!(await this.engine.sync()))
          throw new SyncError('STORAGE', 'Durable storage is unavailable')
      }
      return result
    } catch (error) {
      if (error instanceof KvObjectTypeError)
        throw new SyncError('SCHEMA_MISMATCH', error.message)
      if (committing)
        throw new SyncError(
          'UNCERTAIN',
          'Acceptance could not be confirmed; retry the same request ID and input',
          requestId,
        )
      if (error instanceof KvScanLimitError)
        throw new SyncError('QUERY_LIMIT', error.message)
      if (signal.aborted) {
        this.checkPrincipal(principal)
        throw new SyncError(
          this.signal.aborted ? 'CLOSED' : 'UNAVAILABLE',
          'Operation was canceled',
        )
      }
      throw publicError(error)
    } finally {
      await unit?.release()
    }
  }

  private collections(
    unit: Unit,
    principal: P,
    admin: boolean,
    accesses: Receipt['accesses'],
  ): Transaction<S> {
    return {
      collection: <K extends keyof S['collections'] & string>(name: K) => {
        const validator = this.schema.collections[name]
        const access = async (action: 'read' | 'write') => {
          await this.authorize(principal, name, action, admin)
          if (
            !accesses.some(
              (access) =>
                access.collection === name && access.action === action,
            )
          )
            accesses.push({ collection: name, action })
          return unit.store(
            collectionKey(this.schema.id, principal.scope, name),
          )
        }
        return {
          get: async (key) => {
            const bytes = this.recordKey(key)
            const store = await access('read')
            const entry = await store?.get(bytes, unit.signal)
            return entry?.found ? decodeJSON(entry.data) : undefined
          },
          scan: async (query = {}) => {
            const prefix = query.prefix ?? ''
            const bytes = prefix ? this.recordKey(prefix) : new Uint8Array()
            const store = await access('read')
            if (!store) return []
            const entries = await store.scanRecords(
              bytes,
              {
                maxRecords: this.limits.maxRecords,
                maxBytes: this.limits.maxSnapshotBytes,
              },
              unit.signal,
            )
            return entries.map((entry) => ({
              key: decoder.decode(entry.key),
              value: decodeJSON(entry.value),
            }))
          },
          put: async (key, value) => {
            const bytes = this.recordKey(key)
            const store = await access('write')
            if (!unit.write || !store)
              throw new SyncError('DENIED', 'Transaction is read-only')
            const validated = await validate(validator, value, 'Record')
            const encoded = encodeJSON(validated)
            if (encoded.length > this.limits.maxRecordBytes)
              throw new SyncError(
                'QUERY_LIMIT',
                'Record exceeds the configured byte limit',
              )
            await store.set(bytes, encoded, unit.signal)
          },
          delete: async (key) => {
            const bytes = this.recordKey(key)
            const store = await access('write')
            if (!unit.write || !store)
              throw new SyncError('DENIED', 'Transaction is read-only')
            await store.delete(bytes, unit.signal)
          },
        } as TransactionCollection<S['collections'][K]>
      },
    }
  }

  private recordKey(key: string): Uint8Array {
    try {
      return encodeRecordKey(key)
    } catch {
      throw new SyncError(
        'VALIDATION',
        'Record key must be nonempty UTF-8 within 1024 bytes',
      )
    }
  }
}

// Unit keeps typed KV transactions subordinate to one World writer lifetime.
class Unit {
  private readonly stores = new Map<
    string,
    Promise<{ store: KvStore; tx: KvTransaction } | undefined>
  >()
  private released = false

  constructor(
    private readonly tx: Tx,
    readonly write: boolean,
    readonly signal: AbortSignal,
  ) {}

  async store(key: string): Promise<KvTransaction | undefined> {
    if (this.released) throw new SyncError('CLOSED', 'Transaction is closed')
    let pending = this.stores.get(key)
    if (!pending) {
      pending = (async () => {
        const store = await openWorldKvStore(
          this.tx,
          key,
          this.write,
          this.signal,
        )
        if (!store) return undefined
        try {
          return {
            store,
            tx: await store.openTransaction(this.write, this.signal),
          }
        } catch (error) {
          store.release()
          throw error
        }
      })()
      this.stores.set(key, pending)
    }
    return (await pending)?.tx
  }

  async commit(onCommit?: () => void): Promise<void> {
    const stores = await Promise.all(this.stores.values())
    for (const store of stores) {
      // Each root publication uses the same physical SQLite writer.
      // eslint-disable-next-line react-doctor/async-await-in-loop
      await store?.tx.commit(this.signal)
    }
    this.signal.throwIfAborted()
    onCommit?.()
    // Once sent, the World commit and durability fence must finish despite caller cancellation.
    await this.tx.commit()
  }

  async release(): Promise<void> {
    if (this.released) return
    this.released = true
    const stores = await Promise.allSettled(this.stores.values())
    await Promise.all(
      stores.map(async (result) => {
        if (result.status !== 'fulfilled' || !result.value) return
        try {
          await result.value.tx.discard()
        } catch {
          /* Transport failure already rejects the operation. */
        } finally {
          result.value.store.release()
        }
      }),
    )
    try {
      await this.tx.discard()
    } catch {
      /* Release the local ref even after transport failure. */
    } finally {
      this.tx.release()
    }
  }
}
