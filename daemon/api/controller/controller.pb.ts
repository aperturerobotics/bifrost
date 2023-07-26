/* eslint-disable */
import { Config as Config2 } from '@go/github.com/aperturerobotics/controllerbus/bus/api/api.pb.js'
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { Config as Config1 } from '../api.pb.js'

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
    writer: _m0.Writer = _m0.Writer.create(),
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

          message.listenAddr = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.apiConfig = Config1.decode(reader, reader.uint32())
          continue
        case 3:
          if (tag !== 24) {
            break
          }

          message.disableBusApi = reader.bool()
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          message.busApiConfig = Config2.decode(reader, reader.uint32())
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
    if (message.listenAddr !== '') {
      obj.listenAddr = message.listenAddr
    }
    if (message.apiConfig !== undefined) {
      obj.apiConfig = Config1.toJSON(message.apiConfig)
    }
    if (message.disableBusApi === true) {
      obj.disableBusApi = message.disableBusApi
    }
    if (message.busApiConfig !== undefined) {
      obj.busApiConfig = Config2.toJSON(message.busApiConfig)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? {})
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
