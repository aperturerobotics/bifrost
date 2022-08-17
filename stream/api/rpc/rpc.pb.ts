/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'stream.api.rpc'

/** StreamState is state for the stream related calls. */
export enum StreamState {
  /** StreamState_NONE - StreamState_NONE indicates nothing about the state */
  StreamState_NONE = 0,
  /** StreamState_ESTABLISHING - StreamState_ESTABLISHING indicates the stream is connecting. */
  StreamState_ESTABLISHING = 1,
  /** StreamState_ESTABLISHED - StreamState_ESTABLISHED indicates the stream is established. */
  StreamState_ESTABLISHED = 2,
  UNRECOGNIZED = -1,
}

export function streamStateFromJSON(object: any): StreamState {
  switch (object) {
    case 0:
    case 'StreamState_NONE':
      return StreamState.StreamState_NONE
    case 1:
    case 'StreamState_ESTABLISHING':
      return StreamState.StreamState_ESTABLISHING
    case 2:
    case 'StreamState_ESTABLISHED':
      return StreamState.StreamState_ESTABLISHED
    case -1:
    case 'UNRECOGNIZED':
    default:
      return StreamState.UNRECOGNIZED
  }
}

export function streamStateToJSON(object: StreamState): string {
  switch (object) {
    case StreamState.StreamState_NONE:
      return 'StreamState_NONE'
    case StreamState.StreamState_ESTABLISHING:
      return 'StreamState_ESTABLISHING'
    case StreamState.StreamState_ESTABLISHED:
      return 'StreamState_ESTABLISHED'
    case StreamState.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}

/** Data is a data packet. */
export interface Data {
  /**
   * State indicates stream state in-band.
   * Data is packet data from the remote.
   */
  data: Uint8Array
  /** State indicates the stream state. */
  state: StreamState
}

function createBaseData(): Data {
  return { data: new Uint8Array(), state: 0 }
}

export const Data = {
  encode(message: Data, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.data.length !== 0) {
      writer.uint32(10).bytes(message.data)
    }
    if (message.state !== 0) {
      writer.uint32(16).int32(message.state)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Data {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseData()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.data = reader.bytes()
          break
        case 2:
          message.state = reader.int32() as any
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Data, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Data | Data[]> | Iterable<Data | Data[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Data.encode(p).finish()]
        }
      } else {
        yield* [Data.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Data>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Data> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Data.decode(p)]
        }
      } else {
        yield* [Data.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Data {
    return {
      data: isSet(object.data)
        ? bytesFromBase64(object.data)
        : new Uint8Array(),
      state: isSet(object.state) ? streamStateFromJSON(object.state) : 0,
    }
  },

  toJSON(message: Data): unknown {
    const obj: any = {}
    message.data !== undefined &&
      (obj.data = base64FromBytes(
        message.data !== undefined ? message.data : new Uint8Array()
      ))
    message.state !== undefined &&
      (obj.state = streamStateToJSON(message.state))
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<Data>, I>>(object: I): Data {
    const message = createBaseData()
    message.data = object.data ?? new Uint8Array()
    message.state = object.state ?? 0
    return message
  },
}

declare var self: any | undefined
declare var window: any | undefined
declare var global: any | undefined
var globalThis: any = (() => {
  if (typeof globalThis !== 'undefined') return globalThis
  if (typeof self !== 'undefined') return self
  if (typeof window !== 'undefined') return window
  if (typeof global !== 'undefined') return global
  throw 'Unable to locate global object'
})()

const atob: (b64: string) => string =
  globalThis.atob ||
  ((b64) => globalThis.Buffer.from(b64, 'base64').toString('binary'))
function bytesFromBase64(b64: string): Uint8Array {
  const bin = atob(b64)
  const arr = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; ++i) {
    arr[i] = bin.charCodeAt(i)
  }
  return arr
}

const btoa: (bin: string) => string =
  globalThis.btoa ||
  ((bin) => globalThis.Buffer.from(bin, 'binary').toString('base64'))
function base64FromBytes(arr: Uint8Array): string {
  const bin: string[] = []
  arr.forEach((byte) => {
    bin.push(String.fromCharCode(byte))
  })
  return btoa(bin.join(''))
}

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined

export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Long
  ? string | number | Long
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends { $case: string }
  ? { [K in keyof Omit<T, '$case'>]?: DeepPartial<T[K]> } & {
      $case: T['$case']
    }
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>

type KeysOfUnion<T> = T extends T ? keyof T : never
export type Exact<P, I extends P> = P extends Builtin
  ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & Record<
        Exclude<keyof I, KeysOfUnion<P>>,
        never
      >

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any
  _m0.configure()
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined
}