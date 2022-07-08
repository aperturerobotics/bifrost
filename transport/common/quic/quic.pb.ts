/* eslint-disable */
import Long from 'long'
import * as _m0 from 'protobufjs/minimal'

export const protobufPackage = 'transport.quic'

export interface Opts {
  /**
   * MaxIdleTimeoutDur is the duration of idle after which conn is closed.
   *
   * If unset, the max idle timeout will be disabled.
   */
  maxIdleTimeoutDur: string
  /**
   * MaxIncomingStreams is the maximum number of concurrent bidirectional
   * streams that a peer is allowed to open.
   *
   * If unset or negative, defaults to 100000.
   */
  maxIncomingStreams: number
  /** DisableKeepAlive disables the keep alive packets. */
  disableKeepAlive: boolean
  /**
   * DisableDatagrams disables the unreliable datagrams feature.
   * Both peers must support it for it to be enabled, regardless of this flag.
   */
  disableDatagrams: boolean
  /** DisablePathMtuDiscovery disables sending packets to discover max packet size. */
  disablePathMtuDiscovery: boolean
  /**
   * Verbose indicates to use verbose logging.
   * Note: this is VERY verbose, logs every packet sent.
   */
  verbose: boolean
}

function createBaseOpts(): Opts {
  return {
    maxIdleTimeoutDur: '',
    maxIncomingStreams: 0,
    disableKeepAlive: false,
    disableDatagrams: false,
    disablePathMtuDiscovery: false,
    verbose: false,
  }
}

export const Opts = {
  encode(message: Opts, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.maxIdleTimeoutDur !== '') {
      writer.uint32(10).string(message.maxIdleTimeoutDur)
    }
    if (message.maxIncomingStreams !== 0) {
      writer.uint32(16).int32(message.maxIncomingStreams)
    }
    if (message.disableKeepAlive === true) {
      writer.uint32(24).bool(message.disableKeepAlive)
    }
    if (message.disableDatagrams === true) {
      writer.uint32(32).bool(message.disableDatagrams)
    }
    if (message.disablePathMtuDiscovery === true) {
      writer.uint32(40).bool(message.disablePathMtuDiscovery)
    }
    if (message.verbose === true) {
      writer.uint32(48).bool(message.verbose)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Opts {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseOpts()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.maxIdleTimeoutDur = reader.string()
          break
        case 2:
          message.maxIncomingStreams = reader.int32()
          break
        case 3:
          message.disableKeepAlive = reader.bool()
          break
        case 4:
          message.disableDatagrams = reader.bool()
          break
        case 5:
          message.disablePathMtuDiscovery = reader.bool()
          break
        case 6:
          message.verbose = reader.bool()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Opts, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Opts | Opts[]> | Iterable<Opts | Opts[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Opts.encode(p).finish()]
        }
      } else {
        yield* [Opts.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Opts>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Opts> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Opts.decode(p)]
        }
      } else {
        yield* [Opts.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Opts {
    return {
      maxIdleTimeoutDur: isSet(object.maxIdleTimeoutDur)
        ? String(object.maxIdleTimeoutDur)
        : '',
      maxIncomingStreams: isSet(object.maxIncomingStreams)
        ? Number(object.maxIncomingStreams)
        : 0,
      disableKeepAlive: isSet(object.disableKeepAlive)
        ? Boolean(object.disableKeepAlive)
        : false,
      disableDatagrams: isSet(object.disableDatagrams)
        ? Boolean(object.disableDatagrams)
        : false,
      disablePathMtuDiscovery: isSet(object.disablePathMtuDiscovery)
        ? Boolean(object.disablePathMtuDiscovery)
        : false,
      verbose: isSet(object.verbose) ? Boolean(object.verbose) : false,
    }
  },

  toJSON(message: Opts): unknown {
    const obj: any = {}
    message.maxIdleTimeoutDur !== undefined &&
      (obj.maxIdleTimeoutDur = message.maxIdleTimeoutDur)
    message.maxIncomingStreams !== undefined &&
      (obj.maxIncomingStreams = Math.round(message.maxIncomingStreams))
    message.disableKeepAlive !== undefined &&
      (obj.disableKeepAlive = message.disableKeepAlive)
    message.disableDatagrams !== undefined &&
      (obj.disableDatagrams = message.disableDatagrams)
    message.disablePathMtuDiscovery !== undefined &&
      (obj.disablePathMtuDiscovery = message.disablePathMtuDiscovery)
    message.verbose !== undefined && (obj.verbose = message.verbose)
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<Opts>, I>>(object: I): Opts {
    const message = createBaseOpts()
    message.maxIdleTimeoutDur = object.maxIdleTimeoutDur ?? ''
    message.maxIncomingStreams = object.maxIncomingStreams ?? 0
    message.disableKeepAlive = object.disableKeepAlive ?? false
    message.disableDatagrams = object.disableDatagrams ?? false
    message.disablePathMtuDiscovery = object.disablePathMtuDiscovery ?? false
    message.verbose = object.verbose ?? false
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
