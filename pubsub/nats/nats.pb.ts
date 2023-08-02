/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import {
  HashType,
  hashTypeFromJSON,
  hashTypeToJSON,
} from '../../hash/hash.pb.js'

export const protobufPackage = 'nats'

/** NatsConnType indicates the type of nats conn a stream represents. */
export enum NatsConnType {
  /** NatsConnType_UNKNOWN - NatsConnType_UNKNOWN is the unknown type. */
  NatsConnType_UNKNOWN = 0,
  /** NatsConnType_CLIENT - NatsConnType_CLIENT is the client connection type. */
  NatsConnType_CLIENT = 1,
  /** NatsConnType_ROUTER - NatsConnType_ROUTER is the router-router connection type. */
  NatsConnType_ROUTER = 2,
  UNRECOGNIZED = -1,
}

export function natsConnTypeFromJSON(object: any): NatsConnType {
  switch (object) {
    case 0:
    case 'NatsConnType_UNKNOWN':
      return NatsConnType.NatsConnType_UNKNOWN
    case 1:
    case 'NatsConnType_CLIENT':
      return NatsConnType.NatsConnType_CLIENT
    case 2:
    case 'NatsConnType_ROUTER':
      return NatsConnType.NatsConnType_ROUTER
    case -1:
    case 'UNRECOGNIZED':
    default:
      return NatsConnType.UNRECOGNIZED
  }
}

export function natsConnTypeToJSON(object: NatsConnType): string {
  switch (object) {
    case NatsConnType.NatsConnType_UNKNOWN:
      return 'NatsConnType_UNKNOWN'
    case NatsConnType.NatsConnType_CLIENT:
      return 'NatsConnType_CLIENT'
    case NatsConnType.NatsConnType_ROUTER:
      return 'NatsConnType_ROUTER'
    case NatsConnType.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}

/**
 * Config configures the nats router, hosting a nats.io routing node.
 * This uses nats 2.0 accounts - an Account maps to a Peer.
 */
export interface Config {
  /**
   * ClusterName is the cluster ID string to use.
   * This must be the same on all nodes.
   * If unset, uses the protocol ID.
   */
  clusterName: string
  /**
   * PublishHashType is the hash type to use when signing published messages.
   * Defaults to sha256
   */
  publishHashType: HashType
  /** LogDebug turns on extended debugging logging. */
  logDebug: boolean
  /**
   * LogTrace turns on tracing logging.
   * implies log_debug.
   */
  logTrace: boolean
  /**
   * LogTraceVrebose turns on verbose tracing logging.
   * Implies log_trace and log_debug.
   */
  logTraceVerbose: boolean
}

function createBaseConfig(): Config {
  return {
    clusterName: '',
    publishHashType: 0,
    logDebug: false,
    logTrace: false,
    logTraceVerbose: false,
  }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.clusterName !== '') {
      writer.uint32(10).string(message.clusterName)
    }
    if (message.publishHashType !== 0) {
      writer.uint32(16).int32(message.publishHashType)
    }
    if (message.logDebug === true) {
      writer.uint32(24).bool(message.logDebug)
    }
    if (message.logTrace === true) {
      writer.uint32(32).bool(message.logTrace)
    }
    if (message.logTraceVerbose === true) {
      writer.uint32(40).bool(message.logTraceVerbose)
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

          message.clusterName = reader.string()
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.publishHashType = reader.int32() as any
          continue
        case 3:
          if (tag !== 24) {
            break
          }

          message.logDebug = reader.bool()
          continue
        case 4:
          if (tag !== 32) {
            break
          }

          message.logTrace = reader.bool()
          continue
        case 5:
          if (tag !== 40) {
            break
          }

          message.logTraceVerbose = reader.bool()
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
      clusterName: isSet(object.clusterName) ? String(object.clusterName) : '',
      publishHashType: isSet(object.publishHashType)
        ? hashTypeFromJSON(object.publishHashType)
        : 0,
      logDebug: isSet(object.logDebug) ? Boolean(object.logDebug) : false,
      logTrace: isSet(object.logTrace) ? Boolean(object.logTrace) : false,
      logTraceVerbose: isSet(object.logTraceVerbose)
        ? Boolean(object.logTraceVerbose)
        : false,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    if (message.clusterName !== '') {
      obj.clusterName = message.clusterName
    }
    if (message.publishHashType !== 0) {
      obj.publishHashType = hashTypeToJSON(message.publishHashType)
    }
    if (message.logDebug === true) {
      obj.logDebug = message.logDebug
    }
    if (message.logTrace === true) {
      obj.logTrace = message.logTrace
    }
    if (message.logTraceVerbose === true) {
      obj.logTraceVerbose = message.logTraceVerbose
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.clusterName = object.clusterName ?? ''
    message.publishHashType = object.publishHashType ?? 0
    message.logDebug = object.logDebug ?? false
    message.logTrace = object.logTrace ?? false
    message.logTraceVerbose = object.logTraceVerbose ?? false
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
