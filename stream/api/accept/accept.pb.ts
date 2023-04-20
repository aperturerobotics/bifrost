/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'stream.api.accept'

/** Config configures the accept controller. */
export interface Config {
  /**
   * LocalPeerId is the peer ID to accept incoming connections with.
   * Can be empty to accept any peer.
   */
  localPeerId: string
  /**
   * RemotePeerIds are peer IDs to accept incoming connections from.
   * Can be empty to accept any remote peer IDs.
   */
  remotePeerIds: string[]
  /** ProtocolId is the protocol ID to accept. */
  protocolId: string
  /**
   * TransportId constrains the transport ID to accept from.
   * Can be empty.
   */
  transportId: Long
}

function createBaseConfig(): Config {
  return {
    localPeerId: '',
    remotePeerIds: [],
    protocolId: '',
    transportId: Long.UZERO,
  }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.localPeerId !== '') {
      writer.uint32(10).string(message.localPeerId)
    }
    for (const v of message.remotePeerIds) {
      writer.uint32(18).string(v!)
    }
    if (message.protocolId !== '') {
      writer.uint32(26).string(message.protocolId)
    }
    if (!message.transportId.isZero()) {
      writer.uint32(32).uint64(message.transportId)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Config {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseConfig()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag != 10) {
            break
          }

          message.localPeerId = reader.string()
          continue
        case 2:
          if (tag != 18) {
            break
          }

          message.remotePeerIds.push(reader.string())
          continue
        case 3:
          if (tag != 26) {
            break
          }

          message.protocolId = reader.string()
          continue
        case 4:
          if (tag != 32) {
            break
          }

          message.transportId = reader.uint64() as Long
          continue
      }
      if ((tag & 7) == 4 || tag == 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Config, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Config | Config[]> | Iterable<Config | Config[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config.encode(p).finish()]
        }
      } else {
        yield* [Config.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Config>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Config> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config.decode(p)]
        }
      } else {
        yield* [Config.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Config {
    return {
      localPeerId: isSet(object.localPeerId) ? String(object.localPeerId) : '',
      remotePeerIds: Array.isArray(object?.remotePeerIds)
        ? object.remotePeerIds.map((e: any) => String(e))
        : [],
      protocolId: isSet(object.protocolId) ? String(object.protocolId) : '',
      transportId: isSet(object.transportId)
        ? Long.fromValue(object.transportId)
        : Long.UZERO,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    message.localPeerId !== undefined && (obj.localPeerId = message.localPeerId)
    if (message.remotePeerIds) {
      obj.remotePeerIds = message.remotePeerIds.map((e) => e)
    } else {
      obj.remotePeerIds = []
    }
    message.protocolId !== undefined && (obj.protocolId = message.protocolId)
    message.transportId !== undefined &&
      (obj.transportId = (message.transportId || Long.UZERO).toString())
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.localPeerId = object.localPeerId ?? ''
    message.remotePeerIds = object.remotePeerIds?.map((e) => e) || []
    message.protocolId = object.protocolId ?? ''
    message.transportId =
      object.transportId !== undefined && object.transportId !== null
        ? Long.fromValue(object.transportId)
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
