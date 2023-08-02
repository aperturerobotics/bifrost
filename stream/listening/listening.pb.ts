/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'stream.listening'

/** Config configures the listening controller. */
export interface Config {
  /**
   * LocalPeerId is the peer ID to forward incoming connections with.
   * Can be empty.
   */
  localPeerId: string
  /** RemotePeerId is the peer ID to forward incoming connections to. */
  remotePeerId: string
  /** ProtocolId is the protocol ID to assign to incoming connections. */
  protocolId: string
  /** ListenMultiaddr is the listening multiaddress. */
  listenMultiaddr: string
  /**
   * TransportId sets a transport ID constraint.
   * Can be empty.
   */
  transportId: Long
  /** Reliable indicates the stream should be reliable. */
  reliable: boolean
  /** Encrypted indicates the stream should be encrypted. */
  encrypted: boolean
}

function createBaseConfig(): Config {
  return {
    localPeerId: '',
    remotePeerId: '',
    protocolId: '',
    listenMultiaddr: '',
    transportId: Long.UZERO,
    reliable: false,
    encrypted: false,
  }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.localPeerId !== '') {
      writer.uint32(10).string(message.localPeerId)
    }
    if (message.remotePeerId !== '') {
      writer.uint32(18).string(message.remotePeerId)
    }
    if (message.protocolId !== '') {
      writer.uint32(26).string(message.protocolId)
    }
    if (message.listenMultiaddr !== '') {
      writer.uint32(34).string(message.listenMultiaddr)
    }
    if (!message.transportId.isZero()) {
      writer.uint32(40).uint64(message.transportId)
    }
    if (message.reliable === true) {
      writer.uint32(48).bool(message.reliable)
    }
    if (message.encrypted === true) {
      writer.uint32(56).bool(message.encrypted)
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
          if (tag !== 10) {
            break
          }

          message.localPeerId = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.remotePeerId = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.protocolId = reader.string()
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          message.listenMultiaddr = reader.string()
          continue
        case 5:
          if (tag !== 40) {
            break
          }

          message.transportId = reader.uint64() as Long
          continue
        case 6:
          if (tag !== 48) {
            break
          }

          message.reliable = reader.bool()
          continue
        case 7:
          if (tag !== 56) {
            break
          }

          message.encrypted = reader.bool()
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
  // Transform<Config, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Config | Config[]> | Iterable<Config | Config[]>,
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
      | Iterable<Uint8Array | Uint8Array[]>,
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
      remotePeerId: isSet(object.remotePeerId)
        ? String(object.remotePeerId)
        : '',
      protocolId: isSet(object.protocolId) ? String(object.protocolId) : '',
      listenMultiaddr: isSet(object.listenMultiaddr)
        ? String(object.listenMultiaddr)
        : '',
      transportId: isSet(object.transportId)
        ? Long.fromValue(object.transportId)
        : Long.UZERO,
      reliable: isSet(object.reliable) ? Boolean(object.reliable) : false,
      encrypted: isSet(object.encrypted) ? Boolean(object.encrypted) : false,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    if (message.localPeerId !== '') {
      obj.localPeerId = message.localPeerId
    }
    if (message.remotePeerId !== '') {
      obj.remotePeerId = message.remotePeerId
    }
    if (message.protocolId !== '') {
      obj.protocolId = message.protocolId
    }
    if (message.listenMultiaddr !== '') {
      obj.listenMultiaddr = message.listenMultiaddr
    }
    if (!message.transportId.isZero()) {
      obj.transportId = (message.transportId || Long.UZERO).toString()
    }
    if (message.reliable === true) {
      obj.reliable = message.reliable
    }
    if (message.encrypted === true) {
      obj.encrypted = message.encrypted
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.localPeerId = object.localPeerId ?? ''
    message.remotePeerId = object.remotePeerId ?? ''
    message.protocolId = object.protocolId ?? ''
    message.listenMultiaddr = object.listenMultiaddr ?? ''
    message.transportId =
      object.transportId !== undefined && object.transportId !== null
        ? Long.fromValue(object.transportId)
        : Long.UZERO
    message.reliable = object.reliable ?? false
    message.encrypted = object.encrypted ?? false
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
