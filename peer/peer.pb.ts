/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { HashType, hashTypeFromJSON, hashTypeToJSON } from '../hash/hash.pb.js'

export const protobufPackage = 'peer'

/** Signature contains a signature by a peer. */
export interface Signature {
  /**
   * PubKey is the public key of the peer.
   * May be empty if the public key is to be inferred from context.
   */
  pubKey: Uint8Array
  /**
   * HashType is the hash type used to hash the data.
   * The signature is then of the hash bytes (usually 32).
   */
  hashType: HashType
  /**
   * SigData contains the signature data.
   * The format is defined by the key type.
   */
  sigData: Uint8Array
}

/** SignedMsg is a message from a peer with a signature. */
export interface SignedMsg {
  /** FromPeerId is the peer identifier of the sender. */
  fromPeerId: string
  /**
   * Signature is the sender signature.
   * Should not contain PubKey, which is inferred from peer id.
   */
  signature: Signature | undefined
  /** Data is the signed data. */
  data: Uint8Array
}

function createBaseSignature(): Signature {
  return { pubKey: new Uint8Array(), hashType: 0, sigData: new Uint8Array() }
}

export const Signature = {
  encode(
    message: Signature,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.pubKey.length !== 0) {
      writer.uint32(10).bytes(message.pubKey)
    }
    if (message.hashType !== 0) {
      writer.uint32(16).int32(message.hashType)
    }
    if (message.sigData.length !== 0) {
      writer.uint32(26).bytes(message.sigData)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Signature {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSignature()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag != 10) {
            break
          }

          message.pubKey = reader.bytes()
          continue
        case 2:
          if (tag != 16) {
            break
          }

          message.hashType = reader.int32() as any
          continue
        case 3:
          if (tag != 26) {
            break
          }

          message.sigData = reader.bytes()
          continue
      }
      if ((tag & 7) == 4 || tag == 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Signature, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Signature | Signature[]>
      | Iterable<Signature | Signature[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Signature.encode(p).finish()]
        }
      } else {
        yield* [Signature.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Signature>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Signature> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Signature.decode(p)]
        }
      } else {
        yield* [Signature.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Signature {
    return {
      pubKey: isSet(object.pubKey)
        ? bytesFromBase64(object.pubKey)
        : new Uint8Array(),
      hashType: isSet(object.hashType) ? hashTypeFromJSON(object.hashType) : 0,
      sigData: isSet(object.sigData)
        ? bytesFromBase64(object.sigData)
        : new Uint8Array(),
    }
  },

  toJSON(message: Signature): unknown {
    const obj: any = {}
    message.pubKey !== undefined &&
      (obj.pubKey = base64FromBytes(
        message.pubKey !== undefined ? message.pubKey : new Uint8Array()
      ))
    message.hashType !== undefined &&
      (obj.hashType = hashTypeToJSON(message.hashType))
    message.sigData !== undefined &&
      (obj.sigData = base64FromBytes(
        message.sigData !== undefined ? message.sigData : new Uint8Array()
      ))
    return obj
  },

  create<I extends Exact<DeepPartial<Signature>, I>>(base?: I): Signature {
    return Signature.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Signature>, I>>(
    object: I
  ): Signature {
    const message = createBaseSignature()
    message.pubKey = object.pubKey ?? new Uint8Array()
    message.hashType = object.hashType ?? 0
    message.sigData = object.sigData ?? new Uint8Array()
    return message
  },
}

function createBaseSignedMsg(): SignedMsg {
  return { fromPeerId: '', signature: undefined, data: new Uint8Array() }
}

export const SignedMsg = {
  encode(
    message: SignedMsg,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.fromPeerId !== '') {
      writer.uint32(10).string(message.fromPeerId)
    }
    if (message.signature !== undefined) {
      Signature.encode(message.signature, writer.uint32(18).fork()).ldelim()
    }
    if (message.data.length !== 0) {
      writer.uint32(26).bytes(message.data)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SignedMsg {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSignedMsg()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag != 10) {
            break
          }

          message.fromPeerId = reader.string()
          continue
        case 2:
          if (tag != 18) {
            break
          }

          message.signature = Signature.decode(reader, reader.uint32())
          continue
        case 3:
          if (tag != 26) {
            break
          }

          message.data = reader.bytes()
          continue
      }
      if ((tag & 7) == 4 || tag == 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<SignedMsg, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<SignedMsg | SignedMsg[]>
      | Iterable<SignedMsg | SignedMsg[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SignedMsg.encode(p).finish()]
        }
      } else {
        yield* [SignedMsg.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, SignedMsg>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<SignedMsg> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SignedMsg.decode(p)]
        }
      } else {
        yield* [SignedMsg.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): SignedMsg {
    return {
      fromPeerId: isSet(object.fromPeerId) ? String(object.fromPeerId) : '',
      signature: isSet(object.signature)
        ? Signature.fromJSON(object.signature)
        : undefined,
      data: isSet(object.data)
        ? bytesFromBase64(object.data)
        : new Uint8Array(),
    }
  },

  toJSON(message: SignedMsg): unknown {
    const obj: any = {}
    message.fromPeerId !== undefined && (obj.fromPeerId = message.fromPeerId)
    message.signature !== undefined &&
      (obj.signature = message.signature
        ? Signature.toJSON(message.signature)
        : undefined)
    message.data !== undefined &&
      (obj.data = base64FromBytes(
        message.data !== undefined ? message.data : new Uint8Array()
      ))
    return obj
  },

  create<I extends Exact<DeepPartial<SignedMsg>, I>>(base?: I): SignedMsg {
    return SignedMsg.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<SignedMsg>, I>>(
    object: I
  ): SignedMsg {
    const message = createBaseSignedMsg()
    message.fromPeerId = object.fromPeerId ?? ''
    message.signature =
      object.signature !== undefined && object.signature !== null
        ? Signature.fromPartial(object.signature)
        : undefined
    message.data = object.data ?? new Uint8Array()
    return message
  },
}

declare var self: any | undefined
declare var window: any | undefined
declare var global: any | undefined
var tsProtoGlobalThis: any = (() => {
  if (typeof globalThis !== 'undefined') {
    return globalThis
  }
  if (typeof self !== 'undefined') {
    return self
  }
  if (typeof window !== 'undefined') {
    return window
  }
  if (typeof global !== 'undefined') {
    return global
  }
  throw 'Unable to locate global object'
})()

function bytesFromBase64(b64: string): Uint8Array {
  if (tsProtoGlobalThis.Buffer) {
    return Uint8Array.from(tsProtoGlobalThis.Buffer.from(b64, 'base64'))
  } else {
    const bin = tsProtoGlobalThis.atob(b64)
    const arr = new Uint8Array(bin.length)
    for (let i = 0; i < bin.length; ++i) {
      arr[i] = bin.charCodeAt(i)
    }
    return arr
  }
}

function base64FromBytes(arr: Uint8Array): string {
  if (tsProtoGlobalThis.Buffer) {
    return tsProtoGlobalThis.Buffer.from(arr).toString('base64')
  } else {
    const bin: string[] = []
    arr.forEach((byte) => {
      bin.push(String.fromCharCode(byte))
    })
    return tsProtoGlobalThis.btoa(bin.join(''))
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
