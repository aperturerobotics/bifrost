import type { ClientResourceRef } from '@aptre/bldr-sdk/resource/client.js'
import { Resource } from '@aptre/bldr-sdk/resource/resource.js'
import { pushable, type Pushable } from 'it-pushable'
import { Client as SRPCClient, openRpcStream } from 'starpc'
import type { Message } from '@aptre/protobuf-es-lite'
import { ItState } from '../../bldr/web/bldr/it-state.js'

import { KvtxClient, KvtxOpsClient } from '../../db/kvtx/rpc/kvtx_srpc.pb.js'
import type {
  KvtxTransactionRequest,
  KvtxTransactionResponse,
} from '../../db/kvtx/rpc/kvtx.pb.js'

// KvStoreTypeID is the world ObjectType id for KVTX stores.
export const KvStoreTypeID = 'kv/store'

// KvKeyEntry is a key and its value byte length from a store scan.
export interface KvKeyEntry {
  // key is the raw key bytes.
  key: Uint8Array
  // byteLength is the length of the key's stored value in bytes.
  byteLength: number
}

export interface KvRecordEntry {
  key: Uint8Array
  value: Uint8Array
}

export interface KvScanLimits {
  maxRecords: number
  maxBytes: number
}

export class KvScanLimitError extends Error {
  constructor() {
    super(
      'The snapshot exceeds its record or byte limit; use a narrower prefix',
    )
    this.name = 'KvScanLimitError'
  }
}

// IKvStore contains the bytes-only KV store interface.
export interface IKvStore {
  keyCount(abortSignal?: AbortSignal): Promise<bigint>
  scanKeys(prefix: Uint8Array, abortSignal?: AbortSignal): Promise<KvKeyEntry[]>
  watchKeys(
    prefix: Uint8Array,
    abortSignal?: AbortSignal,
  ): AsyncIterable<KvKeyEntry[]>
  watchRecords(
    prefix: Uint8Array,
    limits: KvScanLimits,
    abortSignal?: AbortSignal,
  ): AsyncIterable<KvRecordEntry[]>
  get(
    key: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<{ data: Uint8Array; found: boolean }>
  exists(key: Uint8Array, abortSignal?: AbortSignal): Promise<boolean>
  withTransaction<T>(
    write: boolean,
    fn: (tx: KvTransaction) => Promise<T>,
    abortSignal?: AbortSignal,
  ): Promise<T>
  release(): void
  [Symbol.dispose](): void
}

// KvStore is the typed SDK handle for a kv/store object.
export class KvStore extends Resource implements IKvStore {
  private service: KvtxClient

  constructor(resourceRef: ClientResourceRef) {
    super(resourceRef)
    this.service = new KvtxClient(resourceRef.client)
  }

  // keyCount returns the number of keys in the store.
  public async keyCount(abortSignal?: AbortSignal): Promise<bigint> {
    return this.withTransaction(
      false,
      async (tx) => tx.keyCount(abortSignal),
      abortSignal,
    )
  }

  // scanKeys returns every key with the given prefix and its value byte length.
  public async scanKeys(
    prefix: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<KvKeyEntry[]> {
    return this.withTransaction(
      false,
      async (tx) => tx.scanKeys(prefix, abortSignal),
      abortSignal,
    )
  }

  // watchKeys streams every key with the given prefix after committed changes.
  public async *watchKeys(
    prefix: Uint8Array,
    abortSignal?: AbortSignal,
  ): AsyncIterable<KvKeyEntry[]> {
    const stream = this.service.Watch({ prefix, onlyKeys: false }, abortSignal)
    for await (const resp of stream) {
      if (resp.error) throw new Error(resp.error)
      yield (resp.entries ?? []).map((entry) => ({
        key: entry.key ?? new Uint8Array(),
        byteLength: entry.value?.length ?? 0,
      }))
    }
  }

  // watchRecords retains at most one pending complete snapshot during backpressure.
  public async *watchRecords(
    prefix: Uint8Array,
    limits: KvScanLimits,
    abortSignal?: AbortSignal,
  ): AsyncIterable<KvRecordEntry[]> {
    const controller = new AbortController()
    const signal = abortSignal
      ? AbortSignal.any([controller.signal, abortSignal])
      : controller.signal
    signal.throwIfAborted()
    const state = new ItState<KvRecordEntry[]>(undefined, {
      mostRecentOnly: true,
    })
    const iterator = state.getIterable()[Symbol.asyncIterator]()
    const stop = () => {
      void iterator.return?.()
    }
    signal.addEventListener('abort', stop, { once: true })
    let failure: unknown
    const pump = (async () => {
      try {
        for await (const response of this.service.Watch(
          {
            prefix,
            maxRecords: BigInt(limits.maxRecords),
            maxBytes: BigInt(limits.maxBytes),
          },
          signal,
        )) {
          if (response.limitExceeded) throw new KvScanLimitError()
          if (response.error) throw new Error(response.error)
          const entries = (response.entries ?? []).map((entry) => ({
            key: entry.key ?? new Uint8Array(),
            value: entry.value ?? new Uint8Array(),
          }))
          const bytes = entries.reduce(
            (sum, entry) => sum + entry.key.length + entry.value.length,
            0,
          )
          if (entries.length > limits.maxRecords || bytes > limits.maxBytes)
            throw new KvScanLimitError()
          state.pushChangeEvent(entries)
        }
        if (!signal.aborted) throw new Error('kv/store: watch ended')
      } catch (error) {
        failure = error
      } finally {
        await iterator.return?.()
      }
    })()
    try {
      for (;;) {
        const next = await iterator.next()
        if (next.done) break
        yield next.value
      }
      if (failure && !signal.aborted) throw failure
    } finally {
      controller.abort()
      signal.removeEventListener('abort', stop)
      await iterator.return?.()
      await pump
    }
  }

  // get returns a key value, if present.
  public async get(
    key: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<{ data: Uint8Array; found: boolean }> {
    return this.withTransaction(
      false,
      async (tx) => tx.get(key, abortSignal),
      abortSignal,
    )
  }

  // exists checks whether a key exists.
  public async exists(
    key: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<boolean> {
    return this.withTransaction(
      false,
      async (tx) => tx.exists(key, abortSignal),
      abortSignal,
    )
  }

  // withTransaction opens a KVTX transaction and commits or discards it.
  public async withTransaction<T>(
    write: boolean,
    fn: (tx: KvTransaction) => Promise<T>,
    abortSignal?: AbortSignal,
  ): Promise<T> {
    const tx = await this.openTransaction(write, abortSignal)
    try {
      const result = await fn(tx)
      if (write) {
        await tx.commit(abortSignal)
      } else {
        await tx.discard()
      }
      return result
    } catch (err) {
      await tx.discard()
      throw err
    }
  }

  // openTransaction opens a KVTX transaction.
  public async openTransaction(
    write: boolean,
    abortSignal?: AbortSignal,
  ): Promise<KvTransaction> {
    const requests = pushable<Message<KvtxTransactionRequest>>({
      objectMode: true,
    })
    const responses = this.service.KvtxTransaction(requests, abortSignal)
    const iterator = responses[Symbol.asyncIterator]()

    requests.push({
      body: {
        case: 'init',
        value: { write },
      },
    })
    const ack = await nextTransactionResponse(iterator)
    if (ack.body?.case !== 'ack') {
      requests.end()
      throw new Error('kv/store: transaction did not return ack')
    }
    if (ack.body.value.error) {
      requests.end()
      throw new Error(ack.body.value.error)
    }
    const transactionId = ack.body.value.transactionId ?? ''
    if (!transactionId) {
      requests.end()
      throw new Error('kv/store: transaction ack missing transaction id')
    }

    const opsRpc = new SRPCClient(() =>
      openRpcStream(
        transactionId,
        (stream) => this.service.KvtxTransactionRpc(stream, abortSignal),
        false,
      ),
    )
    return new KvTransaction(requests, iterator, new KvtxOpsClient(opsRpc))
  }
}

// KvTransaction wraps a KVTX transaction control stream and ops client.
export class KvTransaction {
  private released = false

  constructor(
    private readonly requests: Pushable<Message<KvtxTransactionRequest>>,
    private readonly responses: AsyncIterator<Message<KvtxTransactionResponse>>,
    private readonly ops: KvtxOpsClient,
  ) {}

  // keyCount returns the number of keys visible in the transaction.
  public async keyCount(abortSignal?: AbortSignal): Promise<bigint> {
    const resp = await this.ops.KeyCount({}, abortSignal)
    return resp.keyCount ?? 0n
  }

  // scanKeys returns every key with the given prefix and its value byte length.
  public async scanKeys(
    prefix: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<KvKeyEntry[]> {
    const entries: KvKeyEntry[] = []
    const stream = this.ops.ScanPrefix({ prefix, onlyKeys: false }, abortSignal)
    for await (const resp of stream) {
      if (resp.error) throw new Error(resp.error)
      if (!resp.key) continue
      entries.push({ key: resp.key, byteLength: resp.value?.length ?? 0 })
    }
    return entries
  }

  // scanRecords retains complete values and rejects an incomplete bounded scan.
  public async scanRecords(
    prefix: Uint8Array,
    limits: KvScanLimits,
    abortSignal?: AbortSignal,
  ): Promise<KvRecordEntry[]> {
    const entries: KvRecordEntry[] = []
    let bytes = 0
    const controller = new AbortController()
    const signal = abortSignal
      ? AbortSignal.any([controller.signal, abortSignal])
      : controller.signal
    try {
      const stream = this.ops.ScanPrefix({ prefix, onlyKeys: false }, signal)
      for await (const response of stream) {
        if (response.error) throw new Error(response.error)
        if (!response.key) continue
        const value = response.value ?? new Uint8Array()
        bytes += response.key.length + value.length
        if (entries.length >= limits.maxRecords || bytes > limits.maxBytes) {
          throw new KvScanLimitError()
        }
        entries.push({ key: response.key, value })
      }
      return entries
    } finally {
      controller.abort()
    }
  }

  // get returns a key value, if present.
  public async get(
    key: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<{ data: Uint8Array; found: boolean }> {
    const resp = await this.ops.KeyData({ key }, abortSignal)
    if (resp.error) throw new Error(resp.error)
    return {
      data: resp.data ?? new Uint8Array(),
      found: resp.found ?? false,
    }
  }

  // exists checks whether a key exists.
  public async exists(
    key: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<boolean> {
    const resp = await this.ops.KeyExists({ key }, abortSignal)
    if (resp.error) throw new Error(resp.error)
    return resp.found ?? false
  }

  // set sets a key value in a write transaction.
  public async set(
    key: Uint8Array,
    value: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<void> {
    const resp = await this.ops.SetKey({ key, value }, abortSignal)
    if (resp.error) throw new Error(resp.error)
  }

  // delete removes a key in a write transaction.
  public async delete(
    key: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<void> {
    const resp = await this.ops.DeleteKey({ key }, abortSignal)
    if (resp.error) throw new Error(resp.error)
  }

  // commit commits the transaction.
  public async commit(abortSignal?: AbortSignal): Promise<void> {
    if (this.released) throw new Error('kv/store: transaction already released')
    this.released = true
    this.requests.push({
      body: {
        case: 'commit',
        value: true,
      },
    })
    try {
      const resp = await nextTransactionResponse(this.responses)
      const complete = resp.body?.case === 'complete' ? resp.body.value : null
      if (!complete) throw new Error('kv/store: transaction did not complete')
      if (complete.error) throw new Error(complete.error)
      if (!complete.committed)
        throw new Error('kv/store: transaction discarded')
    } finally {
      this.requests.end()
      abortSignal?.throwIfAborted()
    }
  }

  // discard discards the transaction.
  public async discard(): Promise<void> {
    if (this.released) return
    this.released = true
    this.requests.push({
      body: {
        case: 'discard',
        value: true,
      },
    })
    try {
      await nextTransactionResponse(this.responses)
    } finally {
      this.requests.end()
    }
  }
}

async function nextTransactionResponse(
  iterator: AsyncIterator<Message<KvtxTransactionResponse>>,
): Promise<Message<KvtxTransactionResponse>> {
  const next = await iterator.next()
  if (next.done) throw new Error('kv/store: transaction stream closed')
  return next.value
}
