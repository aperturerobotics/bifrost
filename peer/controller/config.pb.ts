/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'peer.controller'

/** Config is the peer controller config. */
export interface Config {
  /**
   * PrivKey is the peer private key in either b58 or PEM format.
   * See confparse.MarshalPrivateKey.
   * If not set, the peer private key will be unavailable.
   */
  privKey: string
  /**
   * PubKey is the peer public key.
   * Ignored if priv_key is set.
   */
  pubKey: string
  /**
   * PeerId is the peer identifier.
   * Ignored if priv_key or pub_key are set.
   * The peer ID should contain the public key.
   */
  peerId: string
}

function createBaseConfig(): Config {
  return { privKey: '', pubKey: '', peerId: '' }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.privKey !== '') {
      writer.uint32(10).string(message.privKey)
    }
    if (message.pubKey !== '') {
      writer.uint32(18).string(message.pubKey)
    }
    if (message.peerId !== '') {
      writer.uint32(26).string(message.peerId)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Config {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseConfig()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.privKey = reader.string()
          break
        case 2:
          message.pubKey = reader.string()
          break
        case 3:
          message.peerId = reader.string()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
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
      privKey: isSet(object.privKey) ? String(object.privKey) : '',
      pubKey: isSet(object.pubKey) ? String(object.pubKey) : '',
      peerId: isSet(object.peerId) ? String(object.peerId) : '',
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    message.privKey !== undefined && (obj.privKey = message.privKey)
    message.pubKey !== undefined && (obj.pubKey = message.pubKey)
    message.peerId !== undefined && (obj.peerId = message.peerId)
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.privKey = object.privKey ?? ''
    message.pubKey = object.pubKey ?? ''
    message.peerId = object.peerId ?? ''
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
