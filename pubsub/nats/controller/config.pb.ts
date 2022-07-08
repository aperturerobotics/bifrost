/* eslint-disable */
import Long from 'long'
import { Config as Config1 } from '../nats.pb.js'
import * as _m0 from 'protobufjs/minimal'

export const protobufPackage = 'nats.controller'

/** Config is the nats controller config. */
export interface Config {
  /**
   * PeerID sets the peer ID to attach the server to.
   * Must be set.
   * If set to special value: "any" - binds to any peer.
   */
  peerId: string
  /** NatsConfig configures nats. */
  natsConfig: Config1 | undefined
}

function createBaseConfig(): Config {
  return { peerId: '', natsConfig: undefined }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.peerId !== '') {
      writer.uint32(10).string(message.peerId)
    }
    if (message.natsConfig !== undefined) {
      Config1.encode(message.natsConfig, writer.uint32(18).fork()).ldelim()
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
          message.peerId = reader.string()
          break
        case 2:
          message.natsConfig = Config1.decode(reader, reader.uint32())
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
      peerId: isSet(object.peerId) ? String(object.peerId) : '',
      natsConfig: isSet(object.natsConfig)
        ? Config1.fromJSON(object.natsConfig)
        : undefined,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    message.peerId !== undefined && (obj.peerId = message.peerId)
    message.natsConfig !== undefined &&
      (obj.natsConfig = message.natsConfig
        ? Config1.toJSON(message.natsConfig)
        : undefined)
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.peerId = object.peerId ?? ''
    message.natsConfig =
      object.natsConfig !== undefined && object.natsConfig !== null
        ? Config1.fromPartial(object.natsConfig)
        : undefined
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
