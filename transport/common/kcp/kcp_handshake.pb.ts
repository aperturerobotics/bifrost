/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'kcp'

/** HandshakeExtraData contains the extra data field of the pconn handshake. */
export interface HandshakeExtraData {
  /**
   * LocalTransportUuid is the transport uuid of the sender.
   * This is used for monitoring / analysis at a later time.
   * Coorelates the transport connections between two machines.
   */
  localTransportUuid: Long
}

function createBaseHandshakeExtraData(): HandshakeExtraData {
  return { localTransportUuid: Long.UZERO }
}

export const HandshakeExtraData = {
  encode(
    message: HandshakeExtraData,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (!message.localTransportUuid.isZero()) {
      writer.uint32(8).uint64(message.localTransportUuid)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): HandshakeExtraData {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseHandshakeExtraData()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.localTransportUuid = reader.uint64() as Long
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
  // Transform<HandshakeExtraData, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<HandshakeExtraData | HandshakeExtraData[]>
      | Iterable<HandshakeExtraData | HandshakeExtraData[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [HandshakeExtraData.encode(p).finish()]
        }
      } else {
        yield* [HandshakeExtraData.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, HandshakeExtraData>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<HandshakeExtraData> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [HandshakeExtraData.decode(p)]
        }
      } else {
        yield* [HandshakeExtraData.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): HandshakeExtraData {
    return {
      localTransportUuid: isSet(object.localTransportUuid)
        ? Long.fromValue(object.localTransportUuid)
        : Long.UZERO,
    }
  },

  toJSON(message: HandshakeExtraData): unknown {
    const obj: any = {}
    if (!message.localTransportUuid.isZero()) {
      obj.localTransportUuid = (
        message.localTransportUuid || Long.UZERO
      ).toString()
    }
    return obj
  },

  create<I extends Exact<DeepPartial<HandshakeExtraData>, I>>(
    base?: I,
  ): HandshakeExtraData {
    return HandshakeExtraData.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<HandshakeExtraData>, I>>(
    object: I,
  ): HandshakeExtraData {
    const message = createBaseHandshakeExtraData()
    message.localTransportUuid =
      object.localTransportUuid !== undefined &&
      object.localTransportUuid !== null
        ? Long.fromValue(object.localTransportUuid)
        : Long.UZERO
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

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any
  _m0.configure()
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined
}
