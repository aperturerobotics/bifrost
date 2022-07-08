/* eslint-disable */
import Long from 'long'
import { Config as Config1 } from '../api.pb.js'
import { Config as Config2 } from '../../../vendor/github.com/aperturerobotics/controllerbus/bus/api/api.pb.js'
import * as _m0 from 'protobufjs/minimal'

export const protobufPackage = 'bifrost.api.controller'

/** Config configures the API. */
export interface Config {
  /** ListenAddr is the address to listen on for connections. */
  listenAddr: string
  /** ApiConfig are api config options. */
  apiConfig: Config1 | undefined
  /** DisableBusApi disables the bus api. */
  disableBusApi: boolean
  /**
   * BusApiConfig are controller-bus bus api config options.
   * BusApiConfig are options for controller bus api.
   */
  busApiConfig: Config2 | undefined
}

function createBaseConfig(): Config {
  return {
    listenAddr: '',
    apiConfig: undefined,
    disableBusApi: false,
    busApiConfig: undefined,
  }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.listenAddr !== '') {
      writer.uint32(10).string(message.listenAddr)
    }
    if (message.apiConfig !== undefined) {
      Config1.encode(message.apiConfig, writer.uint32(18).fork()).ldelim()
    }
    if (message.disableBusApi === true) {
      writer.uint32(24).bool(message.disableBusApi)
    }
    if (message.busApiConfig !== undefined) {
      Config2.encode(message.busApiConfig, writer.uint32(34).fork()).ldelim()
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
          message.listenAddr = reader.string()
          break
        case 2:
          message.apiConfig = Config1.decode(reader, reader.uint32())
          break
        case 3:
          message.disableBusApi = reader.bool()
          break
        case 4:
          message.busApiConfig = Config2.decode(reader, reader.uint32())
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
      listenAddr: isSet(object.listenAddr) ? String(object.listenAddr) : '',
      apiConfig: isSet(object.apiConfig)
        ? Config1.fromJSON(object.apiConfig)
        : undefined,
      disableBusApi: isSet(object.disableBusApi)
        ? Boolean(object.disableBusApi)
        : false,
      busApiConfig: isSet(object.busApiConfig)
        ? Config2.fromJSON(object.busApiConfig)
        : undefined,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    message.listenAddr !== undefined && (obj.listenAddr = message.listenAddr)
    message.apiConfig !== undefined &&
      (obj.apiConfig = message.apiConfig
        ? Config1.toJSON(message.apiConfig)
        : undefined)
    message.disableBusApi !== undefined &&
      (obj.disableBusApi = message.disableBusApi)
    message.busApiConfig !== undefined &&
      (obj.busApiConfig = message.busApiConfig
        ? Config2.toJSON(message.busApiConfig)
        : undefined)
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.listenAddr = object.listenAddr ?? ''
    message.apiConfig =
      object.apiConfig !== undefined && object.apiConfig !== null
        ? Config1.fromPartial(object.apiConfig)
        : undefined
    message.disableBusApi = object.disableBusApi ?? false
    message.busApiConfig =
      object.busApiConfig !== undefined && object.busApiConfig !== null
        ? Config2.fromPartial(object.busApiConfig)
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
