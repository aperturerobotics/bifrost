/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'stream.drpc'

/** DprcOpts are drpc connection options. */
export interface DrpcOpts {
  /** ManagerOpts are drpc manager options. */
  managerOpts: ManagerOpts | undefined
}

/** ManagerOpts are drpc manager options. */
export interface ManagerOpts {
  /**
   * WriterBufferSize controls the size of the buffer that we will fill before
   * flushing. Normal writes to streams typically issue a flush explicitly.
   */
  writerBufferSize: number
  /** StreamOpts are options for streams created by the manager. */
  streamOpts: StreamOpts | undefined
  /**
   * InactivityTimeout is the amount of time the manager will wait when creating
   * a NewServerStream. It only includes the time it is reading packets from the
   * remote client. In other words, it only includes the time that the client
   * could delay before invoking an RPC. If zero or negative, no timeout.
   */
  inactivityTimeout: string
}

/** StreamOpts are options for a drpc stream. */
export interface StreamOpts {
  /** SplitSize controls the default size we split packets into frames. */
  splitSize: number
}

function createBaseDrpcOpts(): DrpcOpts {
  return { managerOpts: undefined }
}

export const DrpcOpts = {
  encode(
    message: DrpcOpts,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.managerOpts !== undefined) {
      ManagerOpts.encode(message.managerOpts, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DrpcOpts {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseDrpcOpts()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.managerOpts = ManagerOpts.decode(reader, reader.uint32())
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
  // Transform<DrpcOpts, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<DrpcOpts | DrpcOpts[]>
      | Iterable<DrpcOpts | DrpcOpts[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [DrpcOpts.encode(p).finish()]
        }
      } else {
        yield* [DrpcOpts.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, DrpcOpts>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<DrpcOpts> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [DrpcOpts.decode(p)]
        }
      } else {
        yield* [DrpcOpts.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): DrpcOpts {
    return {
      managerOpts: isSet(object.managerOpts)
        ? ManagerOpts.fromJSON(object.managerOpts)
        : undefined,
    }
  },

  toJSON(message: DrpcOpts): unknown {
    const obj: any = {}
    if (message.managerOpts !== undefined) {
      obj.managerOpts = ManagerOpts.toJSON(message.managerOpts)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<DrpcOpts>, I>>(base?: I): DrpcOpts {
    return DrpcOpts.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<DrpcOpts>, I>>(object: I): DrpcOpts {
    const message = createBaseDrpcOpts()
    message.managerOpts =
      object.managerOpts !== undefined && object.managerOpts !== null
        ? ManagerOpts.fromPartial(object.managerOpts)
        : undefined
    return message
  },
}

function createBaseManagerOpts(): ManagerOpts {
  return { writerBufferSize: 0, streamOpts: undefined, inactivityTimeout: '' }
}

export const ManagerOpts = {
  encode(
    message: ManagerOpts,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.writerBufferSize !== 0) {
      writer.uint32(8).uint32(message.writerBufferSize)
    }
    if (message.streamOpts !== undefined) {
      StreamOpts.encode(message.streamOpts, writer.uint32(18).fork()).ldelim()
    }
    if (message.inactivityTimeout !== '') {
      writer.uint32(26).string(message.inactivityTimeout)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ManagerOpts {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseManagerOpts()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.writerBufferSize = reader.uint32()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.streamOpts = StreamOpts.decode(reader, reader.uint32())
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.inactivityTimeout = reader.string()
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
  // Transform<ManagerOpts, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ManagerOpts | ManagerOpts[]>
      | Iterable<ManagerOpts | ManagerOpts[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ManagerOpts.encode(p).finish()]
        }
      } else {
        yield* [ManagerOpts.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ManagerOpts>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<ManagerOpts> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ManagerOpts.decode(p)]
        }
      } else {
        yield* [ManagerOpts.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): ManagerOpts {
    return {
      writerBufferSize: isSet(object.writerBufferSize)
        ? globalThis.Number(object.writerBufferSize)
        : 0,
      streamOpts: isSet(object.streamOpts)
        ? StreamOpts.fromJSON(object.streamOpts)
        : undefined,
      inactivityTimeout: isSet(object.inactivityTimeout)
        ? globalThis.String(object.inactivityTimeout)
        : '',
    }
  },

  toJSON(message: ManagerOpts): unknown {
    const obj: any = {}
    if (message.writerBufferSize !== 0) {
      obj.writerBufferSize = Math.round(message.writerBufferSize)
    }
    if (message.streamOpts !== undefined) {
      obj.streamOpts = StreamOpts.toJSON(message.streamOpts)
    }
    if (message.inactivityTimeout !== '') {
      obj.inactivityTimeout = message.inactivityTimeout
    }
    return obj
  },

  create<I extends Exact<DeepPartial<ManagerOpts>, I>>(base?: I): ManagerOpts {
    return ManagerOpts.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<ManagerOpts>, I>>(
    object: I,
  ): ManagerOpts {
    const message = createBaseManagerOpts()
    message.writerBufferSize = object.writerBufferSize ?? 0
    message.streamOpts =
      object.streamOpts !== undefined && object.streamOpts !== null
        ? StreamOpts.fromPartial(object.streamOpts)
        : undefined
    message.inactivityTimeout = object.inactivityTimeout ?? ''
    return message
  },
}

function createBaseStreamOpts(): StreamOpts {
  return { splitSize: 0 }
}

export const StreamOpts = {
  encode(
    message: StreamOpts,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.splitSize !== 0) {
      writer.uint32(8).uint32(message.splitSize)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): StreamOpts {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseStreamOpts()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.splitSize = reader.uint32()
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
  // Transform<StreamOpts, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<StreamOpts | StreamOpts[]>
      | Iterable<StreamOpts | StreamOpts[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [StreamOpts.encode(p).finish()]
        }
      } else {
        yield* [StreamOpts.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, StreamOpts>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<StreamOpts> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [StreamOpts.decode(p)]
        }
      } else {
        yield* [StreamOpts.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): StreamOpts {
    return {
      splitSize: isSet(object.splitSize)
        ? globalThis.Number(object.splitSize)
        : 0,
    }
  },

  toJSON(message: StreamOpts): unknown {
    const obj: any = {}
    if (message.splitSize !== 0) {
      obj.splitSize = Math.round(message.splitSize)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<StreamOpts>, I>>(base?: I): StreamOpts {
    return StreamOpts.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<StreamOpts>, I>>(
    object: I,
  ): StreamOpts {
    const message = createBaseStreamOpts()
    message.splitSize = object.splitSize ?? 0
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
