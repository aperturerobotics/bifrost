/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { Timestamp } from '../../../../../../google/protobuf/timestamp.pb.js'

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
  return { data: new Uint8Array(0), channel: '', timestamp: undefined }
}

export const PubMessageInner = {
  encode(
    message: PubMessageInner,
    writer: _m0.Writer = _m0.Writer.create(),
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
        writer.uint32(26).fork(),
      ).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PubMessageInner {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePubMessageInner()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.data = reader.bytes()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.channel = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.timestamp = fromTimestamp(
            Timestamp.decode(reader, reader.uint32()),
          )
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<PubMessageInner, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<PubMessageInner | PubMessageInner[]>
      | Iterable<PubMessageInner | PubMessageInner[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [PubMessageInner.encode(p).finish()]
        }
      } else {
        yield* [PubMessageInner.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, PubMessageInner>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<PubMessageInner> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [PubMessageInner.decode(p)]
        }
      } else {
        yield* [PubMessageInner.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): PubMessageInner {
    return {
      data: isSet(object.data)
        ? bytesFromBase64(object.data)
        : new Uint8Array(0),
      channel: isSet(object.channel) ? globalThis.String(object.channel) : '',
      timestamp: isSet(object.timestamp)
        ? fromJsonTimestamp(object.timestamp)
        : undefined,
    }
  },

  toJSON(message: PubMessageInner): unknown {
    const obj: any = {}
    if (message.data.length !== 0) {
      obj.data = base64FromBytes(message.data)
    }
    if (message.channel !== '') {
      obj.channel = message.channel
    }
    if (message.timestamp !== undefined) {
      obj.timestamp = message.timestamp.toISOString()
    }
    return obj
  },

  create<I extends Exact<DeepPartial<PubMessageInner>, I>>(
    base?: I,
  ): PubMessageInner {
    return PubMessageInner.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<PubMessageInner>, I>>(
    object: I,
  ): PubMessageInner {
    const message = createBasePubMessageInner()
    message.data = object.data ?? new Uint8Array(0)
    message.channel = object.channel ?? ''
    message.timestamp = object.timestamp ?? undefined
    return message
  },
}

function bytesFromBase64(b64: string): Uint8Array {
  if (globalThis.Buffer) {
    return Uint8Array.from(globalThis.Buffer.from(b64, 'base64'))
  } else {
    const bin = globalThis.atob(b64)
    const arr = new Uint8Array(bin.length)
    for (let i = 0; i < bin.length; ++i) {
      arr[i] = bin.charCodeAt(i)
    }
    return arr
  }
}

function base64FromBytes(arr: Uint8Array): string {
  if (globalThis.Buffer) {
    return globalThis.Buffer.from(arr).toString('base64')
  } else {
    const bin: string[] = []
    arr.forEach((byte) => {
      bin.push(globalThis.String.fromCharCode(byte))
    })
    return globalThis.btoa(bin.join(''))
  }
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
    : T extends globalThis.Array<infer U>
      ? globalThis.Array<DeepPartial<U>>
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
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & {
      [K in Exclude<keyof I, KeysOfUnion<P>>]: never
    }

function toTimestamp(date: Date): Timestamp {
  const seconds = numberToLong(Math.trunc(date.getTime() / 1_000))
  const nanos = (date.getTime() % 1_000) * 1_000_000
  return { seconds, nanos }
}

function fromTimestamp(t: Timestamp): Date {
  let millis = (t.seconds.toNumber() || 0) * 1_000
  millis += (t.nanos || 0) / 1_000_000
  return new globalThis.Date(millis)
}

function fromJsonTimestamp(o: any): Date {
  if (o instanceof globalThis.Date) {
    return o
  } else if (typeof o === 'string') {
    return new globalThis.Date(o)
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
