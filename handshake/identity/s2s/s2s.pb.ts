/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 's2s'

export enum PacketType {
  /** PacketType_INIT - INIT initializes the handshake. */
  PacketType_INIT = 0,
  /** PacketType_INIT_ACK - INIT_ACK is the reply to the init. */
  PacketType_INIT_ACK = 1,
  /** PacketType_COMPLETE - COMPLETE is the completion of the handshake. */
  PacketType_COMPLETE = 2,
  UNRECOGNIZED = -1,
}

export function packetTypeFromJSON(object: any): PacketType {
  switch (object) {
    case 0:
    case 'PacketType_INIT':
      return PacketType.PacketType_INIT
    case 1:
    case 'PacketType_INIT_ACK':
      return PacketType.PacketType_INIT_ACK
    case 2:
    case 'PacketType_COMPLETE':
      return PacketType.PacketType_COMPLETE
    case -1:
    case 'UNRECOGNIZED':
    default:
      return PacketType.UNRECOGNIZED
  }
}

export function packetTypeToJSON(object: PacketType): string {
  switch (object) {
    case PacketType.PacketType_INIT:
      return 'PacketType_INIT'
    case PacketType.PacketType_INIT_ACK:
      return 'PacketType_INIT_ACK'
    case PacketType.PacketType_COMPLETE:
      return 'PacketType_COMPLETE'
    case PacketType.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}

/** Packet is a handshake packet. */
export interface Packet {
  /** PacketType is the packet type. */
  packetType: PacketType
  /** InitPkt is the init packet. */
  initPkt: Packet_Init | undefined
  /** InitAck is the init-ack packet. */
  initAckPkt: Packet_InitAck | undefined
  /** Complete is the complete packet. */
  completePkt: Packet_Complete | undefined
}

export interface Packet_Init {
  /** SenderPeerID is the peer ID of the sender. */
  senderPeerId: Uint8Array
  /**
   * ReceiverPeerID is the receiver peer ID, if known.
   * If this does not match, the public key is included in the next message.
   */
  receiverPeerId: Uint8Array
  /** SenderEphPub is the ephemeral public key of the sender. */
  senderEphPub: Uint8Array
}

export interface Packet_InitAck {
  /**
   * SenderEphPub is the ephemeral public key of the sender.
   * This is used to compute the shared secret and decode AckInner.
   */
  senderEphPub: Uint8Array
  /** Ciphertext is a Ciphertext message encoded and encrypted with the shared key. */
  ciphertext: Uint8Array
}

export interface Packet_Complete {
  /** Ciphertext is a Ciphertext message encoded and encrypted with the shared key. */
  ciphertext: Uint8Array
}

export interface Packet_Ciphertext {
  /**
   * TupleSignature is the signature of the two ephemeral pub keys.
   * The signature is made using the sender's public key.
   * The keys are concatinated as AB
   */
  tupleSignature: Uint8Array
  /** SenderPubKey contains B's public key if necessary. */
  senderPubKey: Uint8Array
  /** ReceiverKeyKnown indicates that A's public key is known. */
  receiverKeyKnown: boolean
  /**
   * ExtraInfo contains extra information supplied by the transport.
   * Example: in UDP this is information about what port to dial KCP on.
   */
  extraInfo: Uint8Array
}

function createBasePacket(): Packet {
  return {
    packetType: 0,
    initPkt: undefined,
    initAckPkt: undefined,
    completePkt: undefined,
  }
}

export const Packet = {
  encode(
    message: Packet,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.packetType !== 0) {
      writer.uint32(8).int32(message.packetType)
    }
    if (message.initPkt !== undefined) {
      Packet_Init.encode(message.initPkt, writer.uint32(18).fork()).ldelim()
    }
    if (message.initAckPkt !== undefined) {
      Packet_InitAck.encode(
        message.initAckPkt,
        writer.uint32(26).fork()
      ).ldelim()
    }
    if (message.completePkt !== undefined) {
      Packet_Complete.encode(
        message.completePkt,
        writer.uint32(34).fork()
      ).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Packet {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.packetType = reader.int32() as any
          break
        case 2:
          message.initPkt = Packet_Init.decode(reader, reader.uint32())
          break
        case 3:
          message.initAckPkt = Packet_InitAck.decode(reader, reader.uint32())
          break
        case 4:
          message.completePkt = Packet_Complete.decode(reader, reader.uint32())
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Packet, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Packet | Packet[]> | Iterable<Packet | Packet[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet.encode(p).finish()]
        }
      } else {
        yield* [Packet.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Packet> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet.decode(p)]
        }
      } else {
        yield* [Packet.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Packet {
    return {
      packetType: isSet(object.packetType)
        ? packetTypeFromJSON(object.packetType)
        : 0,
      initPkt: isSet(object.initPkt)
        ? Packet_Init.fromJSON(object.initPkt)
        : undefined,
      initAckPkt: isSet(object.initAckPkt)
        ? Packet_InitAck.fromJSON(object.initAckPkt)
        : undefined,
      completePkt: isSet(object.completePkt)
        ? Packet_Complete.fromJSON(object.completePkt)
        : undefined,
    }
  },

  toJSON(message: Packet): unknown {
    const obj: any = {}
    message.packetType !== undefined &&
      (obj.packetType = packetTypeToJSON(message.packetType))
    message.initPkt !== undefined &&
      (obj.initPkt = message.initPkt
        ? Packet_Init.toJSON(message.initPkt)
        : undefined)
    message.initAckPkt !== undefined &&
      (obj.initAckPkt = message.initAckPkt
        ? Packet_InitAck.toJSON(message.initAckPkt)
        : undefined)
    message.completePkt !== undefined &&
      (obj.completePkt = message.completePkt
        ? Packet_Complete.toJSON(message.completePkt)
        : undefined)
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<Packet>, I>>(object: I): Packet {
    const message = createBasePacket()
    message.packetType = object.packetType ?? 0
    message.initPkt =
      object.initPkt !== undefined && object.initPkt !== null
        ? Packet_Init.fromPartial(object.initPkt)
        : undefined
    message.initAckPkt =
      object.initAckPkt !== undefined && object.initAckPkt !== null
        ? Packet_InitAck.fromPartial(object.initAckPkt)
        : undefined
    message.completePkt =
      object.completePkt !== undefined && object.completePkt !== null
        ? Packet_Complete.fromPartial(object.completePkt)
        : undefined
    return message
  },
}

function createBasePacket_Init(): Packet_Init {
  return {
    senderPeerId: new Uint8Array(),
    receiverPeerId: new Uint8Array(),
    senderEphPub: new Uint8Array(),
  }
}

export const Packet_Init = {
  encode(
    message: Packet_Init,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.senderPeerId.length !== 0) {
      writer.uint32(10).bytes(message.senderPeerId)
    }
    if (message.receiverPeerId.length !== 0) {
      writer.uint32(18).bytes(message.receiverPeerId)
    }
    if (message.senderEphPub.length !== 0) {
      writer.uint32(26).bytes(message.senderEphPub)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Packet_Init {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket_Init()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.senderPeerId = reader.bytes()
          break
        case 2:
          message.receiverPeerId = reader.bytes()
          break
        case 3:
          message.senderEphPub = reader.bytes()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Packet_Init, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Packet_Init | Packet_Init[]>
      | Iterable<Packet_Init | Packet_Init[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet_Init.encode(p).finish()]
        }
      } else {
        yield* [Packet_Init.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet_Init>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Packet_Init> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet_Init.decode(p)]
        }
      } else {
        yield* [Packet_Init.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Packet_Init {
    return {
      senderPeerId: isSet(object.senderPeerId)
        ? bytesFromBase64(object.senderPeerId)
        : new Uint8Array(),
      receiverPeerId: isSet(object.receiverPeerId)
        ? bytesFromBase64(object.receiverPeerId)
        : new Uint8Array(),
      senderEphPub: isSet(object.senderEphPub)
        ? bytesFromBase64(object.senderEphPub)
        : new Uint8Array(),
    }
  },

  toJSON(message: Packet_Init): unknown {
    const obj: any = {}
    message.senderPeerId !== undefined &&
      (obj.senderPeerId = base64FromBytes(
        message.senderPeerId !== undefined
          ? message.senderPeerId
          : new Uint8Array()
      ))
    message.receiverPeerId !== undefined &&
      (obj.receiverPeerId = base64FromBytes(
        message.receiverPeerId !== undefined
          ? message.receiverPeerId
          : new Uint8Array()
      ))
    message.senderEphPub !== undefined &&
      (obj.senderEphPub = base64FromBytes(
        message.senderEphPub !== undefined
          ? message.senderEphPub
          : new Uint8Array()
      ))
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<Packet_Init>, I>>(
    object: I
  ): Packet_Init {
    const message = createBasePacket_Init()
    message.senderPeerId = object.senderPeerId ?? new Uint8Array()
    message.receiverPeerId = object.receiverPeerId ?? new Uint8Array()
    message.senderEphPub = object.senderEphPub ?? new Uint8Array()
    return message
  },
}

function createBasePacket_InitAck(): Packet_InitAck {
  return { senderEphPub: new Uint8Array(), ciphertext: new Uint8Array() }
}

export const Packet_InitAck = {
  encode(
    message: Packet_InitAck,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.senderEphPub.length !== 0) {
      writer.uint32(10).bytes(message.senderEphPub)
    }
    if (message.ciphertext.length !== 0) {
      writer.uint32(18).bytes(message.ciphertext)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Packet_InitAck {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket_InitAck()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.senderEphPub = reader.bytes()
          break
        case 2:
          message.ciphertext = reader.bytes()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Packet_InitAck, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Packet_InitAck | Packet_InitAck[]>
      | Iterable<Packet_InitAck | Packet_InitAck[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet_InitAck.encode(p).finish()]
        }
      } else {
        yield* [Packet_InitAck.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet_InitAck>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Packet_InitAck> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet_InitAck.decode(p)]
        }
      } else {
        yield* [Packet_InitAck.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Packet_InitAck {
    return {
      senderEphPub: isSet(object.senderEphPub)
        ? bytesFromBase64(object.senderEphPub)
        : new Uint8Array(),
      ciphertext: isSet(object.ciphertext)
        ? bytesFromBase64(object.ciphertext)
        : new Uint8Array(),
    }
  },

  toJSON(message: Packet_InitAck): unknown {
    const obj: any = {}
    message.senderEphPub !== undefined &&
      (obj.senderEphPub = base64FromBytes(
        message.senderEphPub !== undefined
          ? message.senderEphPub
          : new Uint8Array()
      ))
    message.ciphertext !== undefined &&
      (obj.ciphertext = base64FromBytes(
        message.ciphertext !== undefined ? message.ciphertext : new Uint8Array()
      ))
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<Packet_InitAck>, I>>(
    object: I
  ): Packet_InitAck {
    const message = createBasePacket_InitAck()
    message.senderEphPub = object.senderEphPub ?? new Uint8Array()
    message.ciphertext = object.ciphertext ?? new Uint8Array()
    return message
  },
}

function createBasePacket_Complete(): Packet_Complete {
  return { ciphertext: new Uint8Array() }
}

export const Packet_Complete = {
  encode(
    message: Packet_Complete,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.ciphertext.length !== 0) {
      writer.uint32(10).bytes(message.ciphertext)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Packet_Complete {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket_Complete()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.ciphertext = reader.bytes()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Packet_Complete, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Packet_Complete | Packet_Complete[]>
      | Iterable<Packet_Complete | Packet_Complete[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet_Complete.encode(p).finish()]
        }
      } else {
        yield* [Packet_Complete.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet_Complete>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Packet_Complete> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet_Complete.decode(p)]
        }
      } else {
        yield* [Packet_Complete.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Packet_Complete {
    return {
      ciphertext: isSet(object.ciphertext)
        ? bytesFromBase64(object.ciphertext)
        : new Uint8Array(),
    }
  },

  toJSON(message: Packet_Complete): unknown {
    const obj: any = {}
    message.ciphertext !== undefined &&
      (obj.ciphertext = base64FromBytes(
        message.ciphertext !== undefined ? message.ciphertext : new Uint8Array()
      ))
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<Packet_Complete>, I>>(
    object: I
  ): Packet_Complete {
    const message = createBasePacket_Complete()
    message.ciphertext = object.ciphertext ?? new Uint8Array()
    return message
  },
}

function createBasePacket_Ciphertext(): Packet_Ciphertext {
  return {
    tupleSignature: new Uint8Array(),
    senderPubKey: new Uint8Array(),
    receiverKeyKnown: false,
    extraInfo: new Uint8Array(),
  }
}

export const Packet_Ciphertext = {
  encode(
    message: Packet_Ciphertext,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.tupleSignature.length !== 0) {
      writer.uint32(10).bytes(message.tupleSignature)
    }
    if (message.senderPubKey.length !== 0) {
      writer.uint32(18).bytes(message.senderPubKey)
    }
    if (message.receiverKeyKnown === true) {
      writer.uint32(24).bool(message.receiverKeyKnown)
    }
    if (message.extraInfo.length !== 0) {
      writer.uint32(34).bytes(message.extraInfo)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Packet_Ciphertext {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket_Ciphertext()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.tupleSignature = reader.bytes()
          break
        case 2:
          message.senderPubKey = reader.bytes()
          break
        case 3:
          message.receiverKeyKnown = reader.bool()
          break
        case 4:
          message.extraInfo = reader.bytes()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Packet_Ciphertext, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Packet_Ciphertext | Packet_Ciphertext[]>
      | Iterable<Packet_Ciphertext | Packet_Ciphertext[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet_Ciphertext.encode(p).finish()]
        }
      } else {
        yield* [Packet_Ciphertext.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet_Ciphertext>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Packet_Ciphertext> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Packet_Ciphertext.decode(p)]
        }
      } else {
        yield* [Packet_Ciphertext.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Packet_Ciphertext {
    return {
      tupleSignature: isSet(object.tupleSignature)
        ? bytesFromBase64(object.tupleSignature)
        : new Uint8Array(),
      senderPubKey: isSet(object.senderPubKey)
        ? bytesFromBase64(object.senderPubKey)
        : new Uint8Array(),
      receiverKeyKnown: isSet(object.receiverKeyKnown)
        ? Boolean(object.receiverKeyKnown)
        : false,
      extraInfo: isSet(object.extraInfo)
        ? bytesFromBase64(object.extraInfo)
        : new Uint8Array(),
    }
  },

  toJSON(message: Packet_Ciphertext): unknown {
    const obj: any = {}
    message.tupleSignature !== undefined &&
      (obj.tupleSignature = base64FromBytes(
        message.tupleSignature !== undefined
          ? message.tupleSignature
          : new Uint8Array()
      ))
    message.senderPubKey !== undefined &&
      (obj.senderPubKey = base64FromBytes(
        message.senderPubKey !== undefined
          ? message.senderPubKey
          : new Uint8Array()
      ))
    message.receiverKeyKnown !== undefined &&
      (obj.receiverKeyKnown = message.receiverKeyKnown)
    message.extraInfo !== undefined &&
      (obj.extraInfo = base64FromBytes(
        message.extraInfo !== undefined ? message.extraInfo : new Uint8Array()
      ))
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<Packet_Ciphertext>, I>>(
    object: I
  ): Packet_Ciphertext {
    const message = createBasePacket_Ciphertext()
    message.tupleSignature = object.tupleSignature ?? new Uint8Array()
    message.senderPubKey = object.senderPubKey ?? new Uint8Array()
    message.receiverKeyKnown = object.receiverKeyKnown ?? false
    message.extraInfo = object.extraInfo ?? new Uint8Array()
    return message
  },
}

declare var self: any | undefined
declare var window: any | undefined
declare var global: any | undefined
var globalThis: any = (() => {
  if (typeof globalThis !== 'undefined') return globalThis
  if (typeof self !== 'undefined') return self
  if (typeof window !== 'undefined') return window
  if (typeof global !== 'undefined') return global
  throw 'Unable to locate global object'
})()

const atob: (b64: string) => string =
  globalThis.atob ||
  ((b64) => globalThis.Buffer.from(b64, 'base64').toString('binary'))
function bytesFromBase64(b64: string): Uint8Array {
  const bin = atob(b64)
  const arr = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; ++i) {
    arr[i] = bin.charCodeAt(i)
  }
  return arr
}

const btoa: (bin: string) => string =
  globalThis.btoa ||
  ((bin) => globalThis.Buffer.from(bin, 'binary').toString('base64'))
function base64FromBytes(arr: Uint8Array): string {
  const bin: string[] = []
  arr.forEach((byte) => {
    bin.push(String.fromCharCode(byte))
  })
  return btoa(bin.join(''))
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
