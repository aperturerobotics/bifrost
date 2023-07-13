/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import {
  BlockCompress,
  blockCompressFromJSON,
  blockCompressToJSON,
} from '../../../util/blockcompress/blockcompress.pb.js'
import {
  BlockCrypt,
  blockCryptFromJSON,
  blockCryptToJSON,
} from '../../../util/blockcrypt/blockcrypt.pb.js'

export const protobufPackage = 'kcp'

/** KCPMode is the mode to set KCP to. */
export enum KCPMode {
  /** KCPMode_UNKNOWN - KCPMode_UNKNOWN defaults to normal mode. */
  KCPMode_UNKNOWN = 0,
  /**
   * KCPMode_NORMAL - KCPMode_NORMAL is the normal mode.
   * NoDelay = 0
   * Interval = 40
   * Resend = 2
   * NoCongestion = 1
   */
  KCPMode_NORMAL = 1,
  /**
   * KCPMode_FAST - KCPMode_FAST is the "fast" mode.
   * NoDelay = 0
   * Interval = 30
   * Resend = 2
   * NoCongestion = 1
   */
  KCPMode_FAST = 2,
  /**
   * KCPMode_FAST2 - KCPMode_FAST2 is the "fast2" mode.
   * NoDelay = 1
   * Interval = 20
   * Resend = 2
   * NoCongestion = 1
   */
  KCPMode_FAST2 = 3,
  /**
   * KCPMode_FAST3 - KCPMode_FAST3 is the "fast3" mode.
   * NoDelay = 1
   * Interval = 10
   * Resend = 2
   * NoCongestion = 1
   */
  KCPMode_FAST3 = 4,
  /**
   * KCPMode_SLOW1 - KCPMode_SLOW1 is the slow 1 mode.
   * NoDelay = 0
   * Interval = 100
   * Resend = 0
   * NoCongestion = 0
   */
  KCPMode_SLOW1 = 5,
  UNRECOGNIZED = -1,
}

export function kCPModeFromJSON(object: any): KCPMode {
  switch (object) {
    case 0:
    case 'KCPMode_UNKNOWN':
      return KCPMode.KCPMode_UNKNOWN
    case 1:
    case 'KCPMode_NORMAL':
      return KCPMode.KCPMode_NORMAL
    case 2:
    case 'KCPMode_FAST':
      return KCPMode.KCPMode_FAST
    case 3:
    case 'KCPMode_FAST2':
      return KCPMode.KCPMode_FAST2
    case 4:
    case 'KCPMode_FAST3':
      return KCPMode.KCPMode_FAST3
    case 5:
    case 'KCPMode_SLOW1':
      return KCPMode.KCPMode_SLOW1
    case -1:
    case 'UNRECOGNIZED':
    default:
      return KCPMode.UNRECOGNIZED
  }
}

export function kCPModeToJSON(object: KCPMode): string {
  switch (object) {
    case KCPMode.KCPMode_UNKNOWN:
      return 'KCPMode_UNKNOWN'
    case KCPMode.KCPMode_NORMAL:
      return 'KCPMode_NORMAL'
    case KCPMode.KCPMode_FAST:
      return 'KCPMode_FAST'
    case KCPMode.KCPMode_FAST2:
      return 'KCPMode_FAST2'
    case KCPMode.KCPMode_FAST3:
      return 'KCPMode_FAST3'
    case KCPMode.KCPMode_SLOW1:
      return 'KCPMode_SLOW1'
    case KCPMode.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}

/** StreamMuxer sets the type of stream muxer to use. */
export enum StreamMuxer {
  /** StreamMuxer_UNKNOWN - StreamMuxer_UNKNOWN defaults to StreamMuxer_XTACI_SMUX */
  StreamMuxer_UNKNOWN = 0,
  /** StreamMuxer_XTACI_SMUX - StreamMuxer_XTACI_SMUX is the xtaci/smux muxer. */
  StreamMuxer_XTACI_SMUX = 1,
  /** StreamMuxer_YAMUX - StreamMuxer_YAMUX is the yamux muxer. */
  StreamMuxer_YAMUX = 2,
  UNRECOGNIZED = -1,
}

export function streamMuxerFromJSON(object: any): StreamMuxer {
  switch (object) {
    case 0:
    case 'StreamMuxer_UNKNOWN':
      return StreamMuxer.StreamMuxer_UNKNOWN
    case 1:
    case 'StreamMuxer_XTACI_SMUX':
      return StreamMuxer.StreamMuxer_XTACI_SMUX
    case 2:
    case 'StreamMuxer_YAMUX':
      return StreamMuxer.StreamMuxer_YAMUX
    case -1:
    case 'UNRECOGNIZED':
    default:
      return StreamMuxer.UNRECOGNIZED
  }
}

export function streamMuxerToJSON(object: StreamMuxer): string {
  switch (object) {
    case StreamMuxer.StreamMuxer_UNKNOWN:
      return 'StreamMuxer_UNKNOWN'
    case StreamMuxer.StreamMuxer_XTACI_SMUX:
      return 'StreamMuxer_XTACI_SMUX'
    case StreamMuxer.StreamMuxer_YAMUX:
      return 'StreamMuxer_YAMUX'
    case StreamMuxer.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}

/** Opts are extra options for the packet conn. */
export interface Opts {
  /**
   * DataShards are the number of FEC data shards to use. By adding t check
   * symbols to the data, a Reed–Solomon code can detect any combination of up
   * to t erroneous symbols, or correct up to ⌊t/2⌋ symbols. As an erasure code,
   * it can correct up to t known erasures, or it can detect and correct
   * combinations of errors and erasures. Furthermore, Reed–Solomon codes are
   * suitable as multiple-burst bit-error correcting codes, since a sequence of
   * b + 1 consecutive bit errors can affect at most two symbols of size b. The
   * choice of t is up to the designer of the code, and may be selected within
   * wide limits. Maximum is 256.
   * Recommended: 10
   * If zero, FEC is disabled.
   */
  dataShards: number
  /**
   * ParityShards are the number of FEC parity shards to use.
   * Recommended: 3
   */
  parityShards: number
  /**
   * Mtu is the maximum transmission unit to use.
   * Defaults to 1350 (UDP safe packet size).
   */
  mtu: number
  /** KcpMode is the KCP mode. */
  kcpMode: KCPMode
  /**
   * BlockCrypt is the block crypto to use.
   * Defaults to AES256.
   * Uses the handshake-negotiated session key.
   */
  blockCrypt: BlockCrypt
  /** BlockCompress is the block compression to use. */
  blockCompress: BlockCompress
  /**
   * StreamMuxer is the stream muxer to use.
   * Defaults to smux.
   */
  streamMuxer: StreamMuxer
}

function createBaseOpts(): Opts {
  return {
    dataShards: 0,
    parityShards: 0,
    mtu: 0,
    kcpMode: 0,
    blockCrypt: 0,
    blockCompress: 0,
    streamMuxer: 0,
  }
}

export const Opts = {
  encode(message: Opts, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.dataShards !== 0) {
      writer.uint32(8).uint32(message.dataShards)
    }
    if (message.parityShards !== 0) {
      writer.uint32(16).uint32(message.parityShards)
    }
    if (message.mtu !== 0) {
      writer.uint32(24).uint32(message.mtu)
    }
    if (message.kcpMode !== 0) {
      writer.uint32(32).int32(message.kcpMode)
    }
    if (message.blockCrypt !== 0) {
      writer.uint32(40).int32(message.blockCrypt)
    }
    if (message.blockCompress !== 0) {
      writer.uint32(48).int32(message.blockCompress)
    }
    if (message.streamMuxer !== 0) {
      writer.uint32(56).int32(message.streamMuxer)
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
          if (tag !== 8) {
            break
          }

          message.dataShards = reader.uint32()
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.parityShards = reader.uint32()
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

          message.kcpMode = reader.int32() as any
          continue
        case 5:
          if (tag !== 40) {
            break
          }

          message.blockCrypt = reader.int32() as any
          continue
        case 6:
          if (tag !== 48) {
            break
          }

          message.blockCompress = reader.int32() as any
          continue
        case 7:
          if (tag !== 56) {
            break
          }

          message.streamMuxer = reader.int32() as any
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
      dataShards: isSet(object.dataShards) ? Number(object.dataShards) : 0,
      parityShards: isSet(object.parityShards)
        ? Number(object.parityShards)
        : 0,
      mtu: isSet(object.mtu) ? Number(object.mtu) : 0,
      kcpMode: isSet(object.kcpMode) ? kCPModeFromJSON(object.kcpMode) : 0,
      blockCrypt: isSet(object.blockCrypt)
        ? blockCryptFromJSON(object.blockCrypt)
        : 0,
      blockCompress: isSet(object.blockCompress)
        ? blockCompressFromJSON(object.blockCompress)
        : 0,
      streamMuxer: isSet(object.streamMuxer)
        ? streamMuxerFromJSON(object.streamMuxer)
        : 0,
    }
  },

  toJSON(message: Opts): unknown {
    const obj: any = {}
    message.dataShards !== undefined &&
      (obj.dataShards = Math.round(message.dataShards))
    message.parityShards !== undefined &&
      (obj.parityShards = Math.round(message.parityShards))
    message.mtu !== undefined && (obj.mtu = Math.round(message.mtu))
    message.kcpMode !== undefined &&
      (obj.kcpMode = kCPModeToJSON(message.kcpMode))
    message.blockCrypt !== undefined &&
      (obj.blockCrypt = blockCryptToJSON(message.blockCrypt))
    message.blockCompress !== undefined &&
      (obj.blockCompress = blockCompressToJSON(message.blockCompress))
    message.streamMuxer !== undefined &&
      (obj.streamMuxer = streamMuxerToJSON(message.streamMuxer))
    return obj
  },

  create<I extends Exact<DeepPartial<Opts>, I>>(base?: I): Opts {
    return Opts.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Opts>, I>>(object: I): Opts {
    const message = createBaseOpts()
    message.dataShards = object.dataShards ?? 0
    message.parityShards = object.parityShards ?? 0
    message.mtu = object.mtu ?? 0
    message.kcpMode = object.kcpMode ?? 0
    message.blockCrypt = object.blockCrypt ?? 0
    message.blockCompress = object.blockCompress ?? 0
    message.streamMuxer = object.streamMuxer ?? 0
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
