/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { DrpcOpts } from '../drpc.pb.js'

export const protobufPackage = 'stream.drpc.server'

/** Config configures the server for the drpc service. */
export interface Config {
  /**
   * PeerIds are the list of peer IDs to listen on.
   * If empty, allows any incoming peer id w/ the protocol id.
   */
  peerIds: string[]
  /** DrpcOpts are options passed to drpc. */
  drpcOpts: DrpcOpts | undefined
}

function createBaseConfig(): Config {
  return { peerIds: [], drpcOpts: undefined }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.peerIds) {
      writer.uint32(10).string(v!)
    }
    if (message.drpcOpts !== undefined) {
      DrpcOpts.encode(message.drpcOpts, writer.uint32(18).fork()).ldelim()
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

          message.peerIds.push(reader.string())
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.drpcOpts = DrpcOpts.decode(reader, reader.uint32())
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
      peerIds: Array.isArray(object?.peerIds)
        ? object.peerIds.map((e: any) => String(e))
        : [],
      drpcOpts: isSet(object.drpcOpts)
        ? DrpcOpts.fromJSON(object.drpcOpts)
        : undefined,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    if (message.peerIds?.length) {
      obj.peerIds = message.peerIds
    }
    if (message.drpcOpts !== undefined) {
      obj.drpcOpts = DrpcOpts.toJSON(message.drpcOpts)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.peerIds = object.peerIds?.map((e) => e) || []
    message.drpcOpts =
      object.drpcOpts !== undefined && object.drpcOpts !== null
        ? DrpcOpts.fromPartial(object.drpcOpts)
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
