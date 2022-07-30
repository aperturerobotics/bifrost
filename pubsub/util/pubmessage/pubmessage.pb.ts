/* eslint-disable */
import { Timestamp } from '@aperturerobotics/ts-proto-common-types/google/protobuf/timestamp.pb.js'
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'pubmessage'

/** PubMessageInner is the signed inner portion of the message. */
export interface PubMessageInner {
  /** Data is the message data. */
  data: Uint8Array
  /** Channel is the channel. */
  channel: string
  /** Timestamp is the message timestamp. */
  timestamp: Date | undefined
}

function createBasePubMessageInner(): PubMessageInner {
  return { data: new Uint8Array(), channel: '', timestamp: undefined }
}

export const PubMessageInner = {
  encode(
    message: PubMessageInner,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.data.length !== 0) {
      writer.uint32(10).bytes(message.data)
    }
    if (message.channel !== '') {
      writer.uint32(18).string(message.channel)
    }
    if (message.timestamp !== undefined) {
      Timestamp.encode(
        toTimestamp(message.timestamp),
        writer.uint32(26).fork()
      ).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PubMessageInner {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePubMessageInner()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.data = reader.bytes()
          break
        case 2:
          message.channel = reader.string()
          break
        case 3:
          message.timestamp = fromTimestamp(
            Timestamp.decode(reader, reader.uint32())
          )
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<PubMessageInner, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<PubMessageInner | PubMessageInner[]>
      | Iterable<PubMessageInner | PubMessageInner[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [PubMessageInner.encode(p).finish()]
        }
      } else {
        yield* [PubMessageInner.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, PubMessageInner>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<PubMessageInner> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [PubMessageInner.decode(p)]
        }
      } else {
        yield* [PubMessageInner.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): PubMessageInner {
    return {
      data: isSet(object.data)
        ? bytesFromBase64(object.data)
        : new Uint8Array(),
      channel: isSet(object.channel) ? String(object.channel) : '',
      timestamp: isSet(object.timestamp)
        ? fromJsonTimestamp(object.timestamp)
        : undefined,
    }
  },

  toJSON(message: PubMessageInner): unknown {
    const obj: any = {}
    message.data !== undefined &&
      (obj.data = base64FromBytes(
        message.data !== undefined ? message.data : new Uint8Array()
      ))
    message.channel !== undefined && (obj.channel = message.channel)
    message.timestamp !== undefined &&
      (obj.timestamp = message.timestamp.toISOString())
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<PubMessageInner>, I>>(
    object: I
  ): PubMessageInner {
    const message = createBasePubMessageInner()
    message.data = object.data ?? new Uint8Array()
    message.channel = object.channel ?? ''
    message.timestamp = object.timestamp ?? undefined
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

function toTimestamp(date: Date): Timestamp {
  const seconds = numberToLong(date.getTime() / 1_000)
  const nanos = (date.getTime() % 1_000) * 1_000_000
  return { seconds, nanos }
}

function fromTimestamp(t: Timestamp): Date {
  let millis = t.seconds.toNumber() * 1_000
  millis += t.nanos / 1_000_000
  return new Date(millis)
}

function fromJsonTimestamp(o: any): Date {
  if (o instanceof Date) {
    return o
  } else if (typeof o === 'string') {
    return new Date(o)
  } else {
    return fromTimestamp(Timestamp.fromJSON(o))
  }
}

function numberToLong(number: number) {
  return Long.fromNumber(number)
}

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any
  _m0.configure()
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined
}
