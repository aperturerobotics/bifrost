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
    writer: _m0.Writer = _m0.Writer.create(),
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
        writer.uint32(26).fork(),
      ).ldelim()
    }
    if (message.completePkt !== undefined) {
      Packet_Complete.encode(
        message.completePkt,
        writer.uint32(34).fork(),
      ).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Packet {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.packetType = reader.int32() as any
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.initPkt = Packet_Init.decode(reader, reader.uint32())
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.initAckPkt = Packet_InitAck.decode(reader, reader.uint32())
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          message.completePkt = Packet_Complete.decode(reader, reader.uint32())
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
  // Transform<Packet, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Packet | Packet[]> | Iterable<Packet | Packet[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet.encode(p).finish()]
        }
      } else {
        yield* [Packet.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Packet> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet.decode(p)]
        }
      } else {
        yield* [Packet.decode(pkt as any)]
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
    if (message.packetType !== 0) {
      obj.packetType = packetTypeToJSON(message.packetType)
    }
    if (message.initPkt !== undefined) {
      obj.initPkt = Packet_Init.toJSON(message.initPkt)
    }
    if (message.initAckPkt !== undefined) {
      obj.initAckPkt = Packet_InitAck.toJSON(message.initAckPkt)
    }
    if (message.completePkt !== undefined) {
      obj.completePkt = Packet_Complete.toJSON(message.completePkt)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Packet>, I>>(base?: I): Packet {
    return Packet.fromPartial(base ?? ({} as any))
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
    senderPeerId: new Uint8Array(0),
    receiverPeerId: new Uint8Array(0),
    senderEphPub: new Uint8Array(0),
  }
}

export const Packet_Init = {
  encode(
    message: Packet_Init,
    writer: _m0.Writer = _m0.Writer.create(),
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
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket_Init()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.senderPeerId = reader.bytes()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.receiverPeerId = reader.bytes()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.senderEphPub = reader.bytes()
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
  // Transform<Packet_Init, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Packet_Init | Packet_Init[]>
      | Iterable<Packet_Init | Packet_Init[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet_Init.encode(p).finish()]
        }
      } else {
        yield* [Packet_Init.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet_Init>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Packet_Init> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet_Init.decode(p)]
        }
      } else {
        yield* [Packet_Init.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): Packet_Init {
    return {
      senderPeerId: isSet(object.senderPeerId)
        ? bytesFromBase64(object.senderPeerId)
        : new Uint8Array(0),
      receiverPeerId: isSet(object.receiverPeerId)
        ? bytesFromBase64(object.receiverPeerId)
        : new Uint8Array(0),
      senderEphPub: isSet(object.senderEphPub)
        ? bytesFromBase64(object.senderEphPub)
        : new Uint8Array(0),
    }
  },

  toJSON(message: Packet_Init): unknown {
    const obj: any = {}
    if (message.senderPeerId.length !== 0) {
      obj.senderPeerId = base64FromBytes(message.senderPeerId)
    }
    if (message.receiverPeerId.length !== 0) {
      obj.receiverPeerId = base64FromBytes(message.receiverPeerId)
    }
    if (message.senderEphPub.length !== 0) {
      obj.senderEphPub = base64FromBytes(message.senderEphPub)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Packet_Init>, I>>(base?: I): Packet_Init {
    return Packet_Init.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Packet_Init>, I>>(
    object: I,
  ): Packet_Init {
    const message = createBasePacket_Init()
    message.senderPeerId = object.senderPeerId ?? new Uint8Array(0)
    message.receiverPeerId = object.receiverPeerId ?? new Uint8Array(0)
    message.senderEphPub = object.senderEphPub ?? new Uint8Array(0)
    return message
  },
}

function createBasePacket_InitAck(): Packet_InitAck {
  return { senderEphPub: new Uint8Array(0), ciphertext: new Uint8Array(0) }
}

export const Packet_InitAck = {
  encode(
    message: Packet_InitAck,
    writer: _m0.Writer = _m0.Writer.create(),
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
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket_InitAck()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.senderEphPub = reader.bytes()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.ciphertext = reader.bytes()
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
  // Transform<Packet_InitAck, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Packet_InitAck | Packet_InitAck[]>
      | Iterable<Packet_InitAck | Packet_InitAck[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet_InitAck.encode(p).finish()]
        }
      } else {
        yield* [Packet_InitAck.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet_InitAck>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Packet_InitAck> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet_InitAck.decode(p)]
        }
      } else {
        yield* [Packet_InitAck.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): Packet_InitAck {
    return {
      senderEphPub: isSet(object.senderEphPub)
        ? bytesFromBase64(object.senderEphPub)
        : new Uint8Array(0),
      ciphertext: isSet(object.ciphertext)
        ? bytesFromBase64(object.ciphertext)
        : new Uint8Array(0),
    }
  },

  toJSON(message: Packet_InitAck): unknown {
    const obj: any = {}
    if (message.senderEphPub.length !== 0) {
      obj.senderEphPub = base64FromBytes(message.senderEphPub)
    }
    if (message.ciphertext.length !== 0) {
      obj.ciphertext = base64FromBytes(message.ciphertext)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Packet_InitAck>, I>>(
    base?: I,
  ): Packet_InitAck {
    return Packet_InitAck.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Packet_InitAck>, I>>(
    object: I,
  ): Packet_InitAck {
    const message = createBasePacket_InitAck()
    message.senderEphPub = object.senderEphPub ?? new Uint8Array(0)
    message.ciphertext = object.ciphertext ?? new Uint8Array(0)
    return message
  },
}

function createBasePacket_Complete(): Packet_Complete {
  return { ciphertext: new Uint8Array(0) }
}

export const Packet_Complete = {
  encode(
    message: Packet_Complete,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.ciphertext.length !== 0) {
      writer.uint32(10).bytes(message.ciphertext)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Packet_Complete {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket_Complete()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.ciphertext = reader.bytes()
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
  // Transform<Packet_Complete, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Packet_Complete | Packet_Complete[]>
      | Iterable<Packet_Complete | Packet_Complete[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet_Complete.encode(p).finish()]
        }
      } else {
        yield* [Packet_Complete.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet_Complete>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Packet_Complete> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet_Complete.decode(p)]
        }
      } else {
        yield* [Packet_Complete.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): Packet_Complete {
    return {
      ciphertext: isSet(object.ciphertext)
        ? bytesFromBase64(object.ciphertext)
        : new Uint8Array(0),
    }
  },

  toJSON(message: Packet_Complete): unknown {
    const obj: any = {}
    if (message.ciphertext.length !== 0) {
      obj.ciphertext = base64FromBytes(message.ciphertext)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Packet_Complete>, I>>(
    base?: I,
  ): Packet_Complete {
    return Packet_Complete.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Packet_Complete>, I>>(
    object: I,
  ): Packet_Complete {
    const message = createBasePacket_Complete()
    message.ciphertext = object.ciphertext ?? new Uint8Array(0)
    return message
  },
}

function createBasePacket_Ciphertext(): Packet_Ciphertext {
  return {
    tupleSignature: new Uint8Array(0),
    senderPubKey: new Uint8Array(0),
    receiverKeyKnown: false,
    extraInfo: new Uint8Array(0),
  }
}

export const Packet_Ciphertext = {
  encode(
    message: Packet_Ciphertext,
    writer: _m0.Writer = _m0.Writer.create(),
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
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePacket_Ciphertext()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.tupleSignature = reader.bytes()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.senderPubKey = reader.bytes()
          continue
        case 3:
          if (tag !== 24) {
            break
          }

          message.receiverKeyKnown = reader.bool()
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          message.extraInfo = reader.bytes()
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
  // Transform<Packet_Ciphertext, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Packet_Ciphertext | Packet_Ciphertext[]>
      | Iterable<Packet_Ciphertext | Packet_Ciphertext[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet_Ciphertext.encode(p).finish()]
        }
      } else {
        yield* [Packet_Ciphertext.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Packet_Ciphertext>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Packet_Ciphertext> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Packet_Ciphertext.decode(p)]
        }
      } else {
        yield* [Packet_Ciphertext.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): Packet_Ciphertext {
    return {
      tupleSignature: isSet(object.tupleSignature)
        ? bytesFromBase64(object.tupleSignature)
        : new Uint8Array(0),
      senderPubKey: isSet(object.senderPubKey)
        ? bytesFromBase64(object.senderPubKey)
        : new Uint8Array(0),
      receiverKeyKnown: isSet(object.receiverKeyKnown)
        ? globalThis.Boolean(object.receiverKeyKnown)
        : false,
      extraInfo: isSet(object.extraInfo)
        ? bytesFromBase64(object.extraInfo)
        : new Uint8Array(0),
    }
  },

  toJSON(message: Packet_Ciphertext): unknown {
    const obj: any = {}
    if (message.tupleSignature.length !== 0) {
      obj.tupleSignature = base64FromBytes(message.tupleSignature)
    }
    if (message.senderPubKey.length !== 0) {
      obj.senderPubKey = base64FromBytes(message.senderPubKey)
    }
    if (message.receiverKeyKnown === true) {
      obj.receiverKeyKnown = message.receiverKeyKnown
    }
    if (message.extraInfo.length !== 0) {
      obj.extraInfo = base64FromBytes(message.extraInfo)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Packet_Ciphertext>, I>>(
    base?: I,
  ): Packet_Ciphertext {
    return Packet_Ciphertext.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Packet_Ciphertext>, I>>(
    object: I,
  ): Packet_Ciphertext {
    const message = createBasePacket_Ciphertext()
    message.tupleSignature = object.tupleSignature ?? new Uint8Array(0)
    message.senderPubKey = object.senderPubKey ?? new Uint8Array(0)
    message.receiverKeyKnown = object.receiverKeyKnown ?? false
    message.extraInfo = object.extraInfo ?? new Uint8Array(0)
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
