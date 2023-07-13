/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { Opts as Opts1 } from '../quic/quic.pb.js'

export const protobufPackage = 'conn'

/** Opts are extra options for the reliable conn. */
export interface Opts {
  /** Quic are the quic protocol options. */
  quic: Opts1 | undefined
  /** Verbose turns on verbose debug logging. */
  verbose: boolean
  /**
   * Mtu sets the maximum size for a single packet.
   * Defaults to 65000.
   */
  mtu: number
  /**
   * BufSize is the number of packets to buffer.
   *
   * Total memory cap is mtu * bufSize.
   * Defaults to 10.
   */
  bufSize: number
}

function createBaseOpts(): Opts {
  return { quic: undefined, verbose: false, mtu: 0, bufSize: 0 }
}

export const Opts = {
  encode(message: Opts, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.quic !== undefined) {
      Opts1.encode(message.quic, writer.uint32(10).fork()).ldelim()
    }
    if (message.verbose === true) {
      writer.uint32(16).bool(message.verbose)
    }
    if (message.mtu !== 0) {
      writer.uint32(24).uint32(message.mtu)
    }
    if (message.bufSize !== 0) {
      writer.uint32(32).uint32(message.bufSize)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Opts {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseOpts()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.quic = Opts1.decode(reader, reader.uint32())
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.verbose = reader.bool()
          continue
        case 3:
          if (tag !== 24) {
            break
          }

          message.mtu = reader.uint32()
          continue
        case 4:
          if (tag !== 32) {
            break
          }

          message.bufSize = reader.uint32()
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
  // Transform<Opts, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Opts | Opts[]> | Iterable<Opts | Opts[]>,
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
      | Iterable<Uint8Array | Uint8Array[]>,
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
      quic: isSet(object.quic) ? Opts1.fromJSON(object.quic) : undefined,
      verbose: isSet(object.verbose) ? Boolean(object.verbose) : false,
      mtu: isSet(object.mtu) ? Number(object.mtu) : 0,
      bufSize: isSet(object.bufSize) ? Number(object.bufSize) : 0,
    }
  },

  toJSON(message: Opts): unknown {
    const obj: any = {}
    message.quic !== undefined &&
      (obj.quic = message.quic ? Opts1.toJSON(message.quic) : undefined)
    message.verbose !== undefined && (obj.verbose = message.verbose)
    message.mtu !== undefined && (obj.mtu = Math.round(message.mtu))
    message.bufSize !== undefined && (obj.bufSize = Math.round(message.bufSize))
    return obj
  },

  create<I extends Exact<DeepPartial<Opts>, I>>(base?: I): Opts {
    return Opts.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Opts>, I>>(object: I): Opts {
    const message = createBaseOpts()
    message.quic =
      object.quic !== undefined && object.quic !== null
        ? Opts1.fromPartial(object.quic)
        : undefined
    message.verbose = object.verbose ?? false
    message.mtu = object.mtu ?? 0
    message.bufSize = object.bufSize ?? 0
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
