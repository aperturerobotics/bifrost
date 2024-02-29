/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { Config as Config1 } from '../client.pb.js'

export const protobufPackage = 'stream.srpc.client.controller'

/**
 * Config configures mounting a bifrost srpc RPC client to a bus.
 * Resolves the LookupRpcClient directive.
 */
export interface Config {
  /** Client contains srpc.client configuration for the RPC client. */
  client: Config1 | undefined
  /**
   * ProtocolId is the protocol ID to use to contact the remote RPC service.
   * Must be set.
   */
  protocolId: string
  /**
   * ServiceIdPrefixes are the service ID prefixes to match.
   * The prefix will be stripped from the service id before being passed to the client.
   * This is used like: LookupRpcClient<remote/my/service> -> my/service.
   *
   * If empty slice or empty string: matches all LookupRpcClient calls ignoring service ID.
   * Optional.
   */
  serviceIdPrefixes: string[]
}

function createBaseConfig(): Config {
  return { client: undefined, protocolId: '', serviceIdPrefixes: [] }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.client !== undefined) {
      Config1.encode(message.client, writer.uint32(10).fork()).ldelim()
    }
    if (message.protocolId !== '') {
      writer.uint32(18).string(message.protocolId)
    }
    for (const v of message.serviceIdPrefixes) {
      writer.uint32(26).string(v!)
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

          message.client = Config1.decode(reader, reader.uint32())
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.protocolId = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.serviceIdPrefixes.push(reader.string())
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
      client: isSet(object.client)
        ? Config1.fromJSON(object.client)
        : undefined,
      protocolId: isSet(object.protocolId)
        ? globalThis.String(object.protocolId)
        : '',
      serviceIdPrefixes: globalThis.Array.isArray(object?.serviceIdPrefixes)
        ? object.serviceIdPrefixes.map((e: any) => globalThis.String(e))
        : [],
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    if (message.client !== undefined) {
      obj.client = Config1.toJSON(message.client)
    }
    if (message.protocolId !== '') {
      obj.protocolId = message.protocolId
    }
    if (message.serviceIdPrefixes?.length) {
      obj.serviceIdPrefixes = message.serviceIdPrefixes
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.client =
      object.client !== undefined && object.client !== null
        ? Config1.fromPartial(object.client)
        : undefined
    message.protocolId = object.protocolId ?? ''
    message.serviceIdPrefixes = object.serviceIdPrefixes?.map((e) => e) || []
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
