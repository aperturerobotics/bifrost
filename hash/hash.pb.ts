/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'hash'

/** HashType identifies the hash type in use. */
export enum HashType {
  /** HashType_UNKNOWN - HashType_UNKNOWN is an unknown hash type. */
  HashType_UNKNOWN = 0,
  /** HashType_SHA256 - HashType_SHA256 is the sha256 hash type. */
  HashType_SHA256 = 1,
  /**
   * HashType_SHA1 - HashType_SHA1 is the sha1 hash type.
   * Note: this is not recommended for use outside of backwards-compat.
   */
  HashType_SHA1 = 2,
  /**
   * HashType_BLAKE3 - HashType_BLAKE3 is the blake3 hash type.
   * Uses a 32-byte digest size.
   */
  HashType_BLAKE3 = 3,
  UNRECOGNIZED = -1,
}

export function hashTypeFromJSON(object: any): HashType {
  switch (object) {
    case 0:
    case 'HashType_UNKNOWN':
      return HashType.HashType_UNKNOWN
    case 1:
    case 'HashType_SHA256':
      return HashType.HashType_SHA256
    case 2:
    case 'HashType_SHA1':
      return HashType.HashType_SHA1
    case 3:
    case 'HashType_BLAKE3':
      return HashType.HashType_BLAKE3
    case -1:
    case 'UNRECOGNIZED':
    default:
      return HashType.UNRECOGNIZED
  }
}

export function hashTypeToJSON(object: HashType): string {
  switch (object) {
    case HashType.HashType_UNKNOWN:
      return 'HashType_UNKNOWN'
    case HashType.HashType_SHA256:
      return 'HashType_SHA256'
    case HashType.HashType_SHA1:
      return 'HashType_SHA1'
    case HashType.HashType_BLAKE3:
      return 'HashType_BLAKE3'
    case HashType.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}

/** Hash is a hash of a binary blob. */
export interface Hash {
  /** HashType is the hash type in use. */
  hashType: HashType
  /** Hash is the hash value. */
  hash: Uint8Array
}

function createBaseHash(): Hash {
  return { hashType: 0, hash: new Uint8Array(0) }
}

export const Hash = {
  encode(message: Hash, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.hashType !== 0) {
      writer.uint32(8).int32(message.hashType)
    }
    if (message.hash.length !== 0) {
      writer.uint32(18).bytes(message.hash)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Hash {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseHash()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.hashType = reader.int32() as any
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.hash = reader.bytes()
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
  // Transform<Hash, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Hash | Hash[]> | Iterable<Hash | Hash[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Hash.encode(p).finish()]
        }
      } else {
        yield* [Hash.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Hash>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Hash> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Hash.decode(p)]
        }
      } else {
        yield* [Hash.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): Hash {
    return {
      hashType: isSet(object.hashType) ? hashTypeFromJSON(object.hashType) : 0,
      hash: isSet(object.hash)
        ? bytesFromBase64(object.hash)
        : new Uint8Array(0),
    }
  },

  toJSON(message: Hash): unknown {
    const obj: any = {}
    if (message.hashType !== 0) {
      obj.hashType = hashTypeToJSON(message.hashType)
    }
    if (message.hash.length !== 0) {
      obj.hash = base64FromBytes(message.hash)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Hash>, I>>(base?: I): Hash {
    return Hash.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Hash>, I>>(object: I): Hash {
    const message = createBaseHash()
    message.hashType = object.hashType ?? 0
    message.hash = object.hash ?? new Uint8Array(0)
    return message
  },
}

function bytesFromBase64(b64: string): Uint8Array {
  if (globalThis.Buffer) {
    return Uint8Array.from(globalThis.Buffer.from(b64, 'base64'))
  } else {
    const bin = globalThis.atob(b64)
    const arr = new Uint8Array(bin.length)
    for (let i = 0; i < bin.length; ++i) {
      arr[i] = bin.charCodeAt(i)
    }
    return arr
  }
}

function base64FromBytes(arr: Uint8Array): string {
  if (globalThis.Buffer) {
    return globalThis.Buffer.from(arr).toString('base64')
  } else {
    const bin: string[] = []
    arr.forEach((byte) => {
      bin.push(globalThis.String.fromCharCode(byte))
    })
    return globalThis.btoa(bin.join(''))
  }
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
