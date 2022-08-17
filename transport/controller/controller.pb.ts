/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'transport.controller'

/**
 * StreamEstablish is the first message sent by the initiator of a stream.
 * Prefixed by a uint32 length.
 * Max size: 100kb
 */
export interface StreamEstablish {
  /** ProtocolID is the protocol identifier string for the stream. */
  protocolId: string
}

function createBaseStreamEstablish(): StreamEstablish {
  return { protocolId: '' }
}

export const StreamEstablish = {
  encode(
    message: StreamEstablish,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.protocolId !== '') {
      writer.uint32(10).string(message.protocolId)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): StreamEstablish {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseStreamEstablish()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.protocolId = reader.string()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<StreamEstablish, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<StreamEstablish | StreamEstablish[]>
      | Iterable<StreamEstablish | StreamEstablish[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [StreamEstablish.encode(p).finish()]
        }
      } else {
        yield* [StreamEstablish.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, StreamEstablish>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<StreamEstablish> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [StreamEstablish.decode(p)]
        }
      } else {
        yield* [StreamEstablish.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): StreamEstablish {
    return {
      protocolId: isSet(object.protocolId) ? String(object.protocolId) : '',
    }
  },

  toJSON(message: StreamEstablish): unknown {
    const obj: any = {}
    message.protocolId !== undefined && (obj.protocolId = message.protocolId)
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<StreamEstablish>, I>>(
    object: I
  ): StreamEstablish {
    const message = createBaseStreamEstablish()
    message.protocolId = object.protocolId ?? ''
    return message
  },
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