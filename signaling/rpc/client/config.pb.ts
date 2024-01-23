/* eslint-disable */
import { Backoff } from '@go/github.com/aperturerobotics/util/backoff/backoff.pb.js'
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { Config as Config1 } from '../../../stream/srpc/client/client.pb.js'

export const protobufPackage = 'signaling.rpc.client'

/** Config configures a client for the Signaling SRPC service. */
export interface Config {
  /**
   * SignalingId is the signaling channel ID.
   * Filters which SignalPeer directives will be handled.
   */
  signalingId: string
  /**
   * PeerId is the local peer id to use for the client.
   * Can be empty to use any local peer.
   */
  peerId: string
  /**
   * Client contains srpc.client configuration for the signaling RPC client.
   * The local peer ID is overridden with the peer ID of the looked-up peer.
   */
  client: Config1 | undefined
  /**
   * ProtocolId overrides the default protocol id for the signaling client.
   * Default: bifrost/signaling
   */
  protocolId: string
  /**
   * ServiceId overrides the default service id for the signaling client.
   * Default: signaling.rpc.Signaling
   */
  serviceId: string
  /**
   * Backoff is the backoff config for connecting to the service.
   * If unset, defaults to reasonable defaults.
   */
  backoff: Backoff | undefined
  /**
   * DisableListen disables listening for incoming sessions.
   * If set, we will only call out, not accept incoming sessions.
   * If false, client will emit HandleSignalPeer directives for incoming sessions.
   */
  disableListen: boolean
}

function createBaseConfig(): Config {
  return {
    signalingId: '',
    peerId: '',
    client: undefined,
    protocolId: '',
    serviceId: '',
    backoff: undefined,
    disableListen: false,
  }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.signalingId !== '') {
      writer.uint32(10).string(message.signalingId)
    }
    if (message.peerId !== '') {
      writer.uint32(18).string(message.peerId)
    }
    if (message.client !== undefined) {
      Config1.encode(message.client, writer.uint32(26).fork()).ldelim()
    }
    if (message.protocolId !== '') {
      writer.uint32(34).string(message.protocolId)
    }
    if (message.serviceId !== '') {
      writer.uint32(42).string(message.serviceId)
    }
    if (message.backoff !== undefined) {
      Backoff.encode(message.backoff, writer.uint32(50).fork()).ldelim()
    }
    if (message.disableListen === true) {
      writer.uint32(56).bool(message.disableListen)
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

          message.signalingId = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.peerId = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.client = Config1.decode(reader, reader.uint32())
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          message.protocolId = reader.string()
          continue
        case 5:
          if (tag !== 42) {
            break
          }

          message.serviceId = reader.string()
          continue
        case 6:
          if (tag !== 50) {
            break
          }

          message.backoff = Backoff.decode(reader, reader.uint32())
          continue
        case 7:
          if (tag !== 56) {
            break
          }

          message.disableListen = reader.bool()
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
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Config.encode(p).finish()]
        }
      } else {
        yield* [Config.encode(pkt as any).finish()]
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
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Config.decode(p)]
        }
      } else {
        yield* [Config.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): Config {
    return {
      signalingId: isSet(object.signalingId)
        ? globalThis.String(object.signalingId)
        : '',
      peerId: isSet(object.peerId) ? globalThis.String(object.peerId) : '',
      client: isSet(object.client)
        ? Config1.fromJSON(object.client)
        : undefined,
      protocolId: isSet(object.protocolId)
        ? globalThis.String(object.protocolId)
        : '',
      serviceId: isSet(object.serviceId)
        ? globalThis.String(object.serviceId)
        : '',
      backoff: isSet(object.backoff)
        ? Backoff.fromJSON(object.backoff)
        : undefined,
      disableListen: isSet(object.disableListen)
        ? globalThis.Boolean(object.disableListen)
        : false,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    if (message.signalingId !== '') {
      obj.signalingId = message.signalingId
    }
    if (message.peerId !== '') {
      obj.peerId = message.peerId
    }
    if (message.client !== undefined) {
      obj.client = Config1.toJSON(message.client)
    }
    if (message.protocolId !== '') {
      obj.protocolId = message.protocolId
    }
    if (message.serviceId !== '') {
      obj.serviceId = message.serviceId
    }
    if (message.backoff !== undefined) {
      obj.backoff = Backoff.toJSON(message.backoff)
    }
    if (message.disableListen === true) {
      obj.disableListen = message.disableListen
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.signalingId = object.signalingId ?? ''
    message.peerId = object.peerId ?? ''
    message.client =
      object.client !== undefined && object.client !== null
        ? Config1.fromPartial(object.client)
        : undefined
    message.protocolId = object.protocolId ?? ''
    message.serviceId = object.serviceId ?? ''
    message.backoff =
      object.backoff !== undefined && object.backoff !== null
        ? Backoff.fromPartial(object.backoff)
        : undefined
    message.disableListen = object.disableListen ?? false
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
