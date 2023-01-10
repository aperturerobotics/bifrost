/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { Opts as Opts1 } from '../quic/quic.pb.js'

export const protobufPackage = 'pconn'

/** Opts are extra options for the packet conn. */
export interface Opts {
  /** Quic are the quic protocol options. */
  quic: Opts1 | undefined
  /** Verbose turns on verbose debug logging. */
  verbose: boolean
}

function createBaseOpts(): Opts {
  return { quic: undefined, verbose: false }
}

export const Opts = {
  encode(message: Opts, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.quic !== undefined) {
      Opts1.encode(message.quic, writer.uint32(10).fork()).ldelim()
    }
    if (message.verbose === true) {
      writer.uint32(16).bool(message.verbose)
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
          message.quic = Opts1.decode(reader, reader.uint32())
          break
        case 2:
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
      quic: isSet(object.quic) ? Opts1.fromJSON(object.quic) : undefined,
      verbose: isSet(object.verbose) ? Boolean(object.verbose) : false,
    }
  },

  toJSON(message: Opts): unknown {
    const obj: any = {}
    message.quic !== undefined &&
      (obj.quic = message.quic ? Opts1.toJSON(message.quic) : undefined)
    message.verbose !== undefined && (obj.verbose = message.verbose)
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
