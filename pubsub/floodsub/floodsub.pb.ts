/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import {
  HashType,
  hashTypeFromJSON,
  hashTypeToJSON,
} from '../../hash/hash.pb.js'
import { SignedMsg } from '../../peer/peer.pb.js'

export const protobufPackage = 'floodsub'

/** Config configures the floodsub router. */
export interface Config {
  /**
   * PublishHashType is the hash type to use when signing published messages.
   * Defaults to sha256
   */
  publishHashType: HashType
}

/** Packet is the floodsub packet. */
export interface Packet {
  /** Subscriptions contains any new subscription changes. */
  subscriptions: SubscriptionOpts[]
  /** Publish contains messages we are publishing. */
  publish: SignedMsg[]
}

/** SubscriptionOpts are subscription options. */
export interface SubscriptionOpts {
  /** Subscribe indicates if we are subscribing to this channel ID. */
  subscribe: boolean
  /** ChannelId is the channel to subscribe to. */
  channelId: string
}

function createBaseConfig(): Config {
  return { publishHashType: 0 }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.publishHashType !== 0) {
      writer.uint32(8).int32(message.publishHashType)
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
          if (tag !== 8) {
            break
          }

          message.publishHashType = reader.int32() as any
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
      publishHashType: isSet(object.publishHashType)
        ? hashTypeFromJSON(object.publishHashType)
        : 0,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    if (message.publishHashType !== 0) {
      obj.publishHashType = hashTypeToJSON(message.publishHashType)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.publishHashType = object.publishHashType ?? 0
    return message
  },
}

function createBasePacket(): Packet {
  return { subscriptions: [], publish: [] }
}

export const Packet = {
  encode(
    message: Packet,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.subscriptions) {
      SubscriptionOpts.encode(v!, writer.uint32(10).fork()).ldelim()
    }
    for (const v of message.publish) {
      SignedMsg.encode(v!, writer.uint32(18).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Packet {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.subscriptions.push(
            SubscriptionOpts.decode(reader, reader.uint32()),
          )
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.publish.push(SignedMsg.decode(reader, reader.uint32()))
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
  // Transform<Packet, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Packet | Packet[]> | Iterable<Packet | Packet[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet.encode(p).finish()]
        }
      } else {
        yield* [Packet.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Packet> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet.decode(p)]
        }
      } else {
        yield* [Packet.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Packet {
    return {
      subscriptions: Array.isArray(object?.subscriptions)
        ? object.subscriptions.map((e: any) => SubscriptionOpts.fromJSON(e))
        : [],
      publish: Array.isArray(object?.publish)
        ? object.publish.map((e: any) => SignedMsg.fromJSON(e))
        : [],
    }
  },

  toJSON(message: Packet): unknown {
    const obj: any = {}
    if (message.subscriptions?.length) {
      obj.subscriptions = message.subscriptions.map((e) =>
        SubscriptionOpts.toJSON(e),
      )
    }
    if (message.publish?.length) {
      obj.publish = message.publish.map((e) => SignedMsg.toJSON(e))
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Packet>, I>>(base?: I): Packet {
    return Packet.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Packet>, I>>(object: I): Packet {
    const message = createBasePacket()
    message.subscriptions =
      object.subscriptions?.map((e) => SubscriptionOpts.fromPartial(e)) || []
    message.publish = object.publish?.map((e) => SignedMsg.fromPartial(e)) || []
    return message
  },
}

function createBaseSubscriptionOpts(): SubscriptionOpts {
  return { subscribe: false, channelId: '' }
}

export const SubscriptionOpts = {
  encode(
    message: SubscriptionOpts,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.subscribe === true) {
      writer.uint32(8).bool(message.subscribe)
    }
    if (message.channelId !== '') {
      writer.uint32(18).string(message.channelId)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SubscriptionOpts {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSubscriptionOpts()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.subscribe = reader.bool()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.channelId = reader.string()
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
  // Transform<SubscriptionOpts, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<SubscriptionOpts | SubscriptionOpts[]>
      | Iterable<SubscriptionOpts | SubscriptionOpts[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SubscriptionOpts.encode(p).finish()]
        }
      } else {
        yield* [SubscriptionOpts.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, SubscriptionOpts>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<SubscriptionOpts> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SubscriptionOpts.decode(p)]
        }
      } else {
        yield* [SubscriptionOpts.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): SubscriptionOpts {
    return {
      subscribe: isSet(object.subscribe) ? Boolean(object.subscribe) : false,
      channelId: isSet(object.channelId) ? String(object.channelId) : '',
    }
  },

  toJSON(message: SubscriptionOpts): unknown {
    const obj: any = {}
    if (message.subscribe === true) {
      obj.subscribe = message.subscribe
    }
    if (message.channelId !== '') {
      obj.channelId = message.channelId
    }
    return obj
  },

  create<I extends Exact<DeepPartial<SubscriptionOpts>, I>>(
    base?: I,
  ): SubscriptionOpts {
    return SubscriptionOpts.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<SubscriptionOpts>, I>>(
    object: I,
  ): SubscriptionOpts {
    const message = createBaseSubscriptionOpts()
    message.subscribe = object.subscribe ?? false
    message.channelId = object.channelId ?? ''
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
