/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { SignedMsg } from '../../peer/peer.pb.js'

export const protobufPackage = 'signaling.rpc'

/** ListenRequest is the body of the Listen request. */
export interface ListenRequest {}

/** ListenResponse is a message sent in a stream in response to Listen. */
export interface ListenResponse {
  body?:
    | { $case: 'setPeer'; setPeer: string }
    | { $case: 'clearPeer'; clearPeer: string }
    | undefined
}

/** SessionRequest is a message sent from the client to the server. */
export interface SessionRequest {
  /**
   * SessionSeqno is the session sequence number.
   * If this doesn't match the current session no, this pkt will be dropped.
   * This should be zero for the init packet.
   */
  sessionSeqno: Long
  body?:
    | { $case: 'init'; init: SessionInit }
    | { $case: 'sendMsg'; sendMsg: SessionMsg }
    | { $case: 'clearMsg'; clearMsg: Long }
    | { $case: 'ackMsg'; ackMsg: Long }
    | undefined
}

/** SessionInit is a message to init a Session. */
export interface SessionInit {
  /** PeerId is the remote peer id we want to contact. */
  peerId: string
}

/** SessionMsg contains a signed message and a sequence number. */
export interface SessionMsg {
  /** SignedMsg is the signed message body. */
  signedMsg: SignedMsg | undefined
  /** Seqno is the message sequence number for clear and ack. */
  seqno: Long
}

/** SessionResponse is a message sent from the server to the client. */
export interface SessionResponse {
  body?:
    | { $case: 'opened'; opened: Long }
    | { $case: 'closed'; closed: boolean }
    | { $case: 'recvMsg'; recvMsg: SessionMsg }
    | { $case: 'clearMsg'; clearMsg: Long }
    | { $case: 'ackMsg'; ackMsg: Long }
    | undefined
}

function createBaseListenRequest(): ListenRequest {
  return {}
}

export const ListenRequest = {
  encode(
    _: ListenRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ListenRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseListenRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<ListenRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ListenRequest | ListenRequest[]>
      | Iterable<ListenRequest | ListenRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ListenRequest.encode(p).finish()]
        }
      } else {
        yield* [ListenRequest.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ListenRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<ListenRequest> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ListenRequest.decode(p)]
        }
      } else {
        yield* [ListenRequest.decode(pkt as any)]
      }
    }
  },

  fromJSON(_: any): ListenRequest {
    return {}
  },

  toJSON(_: ListenRequest): unknown {
    const obj: any = {}
    return obj
  },

  create<I extends Exact<DeepPartial<ListenRequest>, I>>(
    base?: I,
  ): ListenRequest {
    return ListenRequest.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<ListenRequest>, I>>(
    _: I,
  ): ListenRequest {
    const message = createBaseListenRequest()
    return message
  },
}

function createBaseListenResponse(): ListenResponse {
  return { body: undefined }
}

export const ListenResponse = {
  encode(
    message: ListenResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    switch (message.body?.$case) {
      case 'setPeer':
        writer.uint32(10).string(message.body.setPeer)
        break
      case 'clearPeer':
        writer.uint32(18).string(message.body.clearPeer)
        break
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ListenResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseListenResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.body = { $case: 'setPeer', setPeer: reader.string() }
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.body = { $case: 'clearPeer', clearPeer: reader.string() }
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
  // Transform<ListenResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ListenResponse | ListenResponse[]>
      | Iterable<ListenResponse | ListenResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ListenResponse.encode(p).finish()]
        }
      } else {
        yield* [ListenResponse.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ListenResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<ListenResponse> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ListenResponse.decode(p)]
        }
      } else {
        yield* [ListenResponse.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): ListenResponse {
    return {
      body: isSet(object.setPeer)
        ? { $case: 'setPeer', setPeer: globalThis.String(object.setPeer) }
        : isSet(object.clearPeer)
          ? {
              $case: 'clearPeer',
              clearPeer: globalThis.String(object.clearPeer),
            }
          : undefined,
    }
  },

  toJSON(message: ListenResponse): unknown {
    const obj: any = {}
    if (message.body?.$case === 'setPeer') {
      obj.setPeer = message.body.setPeer
    }
    if (message.body?.$case === 'clearPeer') {
      obj.clearPeer = message.body.clearPeer
    }
    return obj
  },

  create<I extends Exact<DeepPartial<ListenResponse>, I>>(
    base?: I,
  ): ListenResponse {
    return ListenResponse.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<ListenResponse>, I>>(
    object: I,
  ): ListenResponse {
    const message = createBaseListenResponse()
    if (
      object.body?.$case === 'setPeer' &&
      object.body?.setPeer !== undefined &&
      object.body?.setPeer !== null
    ) {
      message.body = { $case: 'setPeer', setPeer: object.body.setPeer }
    }
    if (
      object.body?.$case === 'clearPeer' &&
      object.body?.clearPeer !== undefined &&
      object.body?.clearPeer !== null
    ) {
      message.body = { $case: 'clearPeer', clearPeer: object.body.clearPeer }
    }
    return message
  },
}

function createBaseSessionRequest(): SessionRequest {
  return { sessionSeqno: Long.UZERO, body: undefined }
}

export const SessionRequest = {
  encode(
    message: SessionRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (!message.sessionSeqno.equals(Long.UZERO)) {
      writer.uint32(8).uint64(message.sessionSeqno)
    }
    switch (message.body?.$case) {
      case 'init':
        SessionInit.encode(message.body.init, writer.uint32(18).fork()).ldelim()
        break
      case 'sendMsg':
        SessionMsg.encode(
          message.body.sendMsg,
          writer.uint32(26).fork(),
        ).ldelim()
        break
      case 'clearMsg':
        writer.uint32(32).uint64(message.body.clearMsg)
        break
      case 'ackMsg':
        writer.uint32(40).uint64(message.body.ackMsg)
        break
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SessionRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSessionRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.sessionSeqno = reader.uint64() as Long
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.body = {
            $case: 'init',
            init: SessionInit.decode(reader, reader.uint32()),
          }
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.body = {
            $case: 'sendMsg',
            sendMsg: SessionMsg.decode(reader, reader.uint32()),
          }
          continue
        case 4:
          if (tag !== 32) {
            break
          }

          message.body = {
            $case: 'clearMsg',
            clearMsg: reader.uint64() as Long,
          }
          continue
        case 5:
          if (tag !== 40) {
            break
          }

          message.body = { $case: 'ackMsg', ackMsg: reader.uint64() as Long }
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
  // Transform<SessionRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<SessionRequest | SessionRequest[]>
      | Iterable<SessionRequest | SessionRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [SessionRequest.encode(p).finish()]
        }
      } else {
        yield* [SessionRequest.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, SessionRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<SessionRequest> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [SessionRequest.decode(p)]
        }
      } else {
        yield* [SessionRequest.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): SessionRequest {
    return {
      sessionSeqno: isSet(object.sessionSeqno)
        ? Long.fromValue(object.sessionSeqno)
        : Long.UZERO,
      body: isSet(object.init)
        ? { $case: 'init', init: SessionInit.fromJSON(object.init) }
        : isSet(object.sendMsg)
          ? { $case: 'sendMsg', sendMsg: SessionMsg.fromJSON(object.sendMsg) }
          : isSet(object.clearMsg)
            ? { $case: 'clearMsg', clearMsg: Long.fromValue(object.clearMsg) }
            : isSet(object.ackMsg)
              ? { $case: 'ackMsg', ackMsg: Long.fromValue(object.ackMsg) }
              : undefined,
    }
  },

  toJSON(message: SessionRequest): unknown {
    const obj: any = {}
    if (!message.sessionSeqno.equals(Long.UZERO)) {
      obj.sessionSeqno = (message.sessionSeqno || Long.UZERO).toString()
    }
    if (message.body?.$case === 'init') {
      obj.init = SessionInit.toJSON(message.body.init)
    }
    if (message.body?.$case === 'sendMsg') {
      obj.sendMsg = SessionMsg.toJSON(message.body.sendMsg)
    }
    if (message.body?.$case === 'clearMsg') {
      obj.clearMsg = (message.body.clearMsg || Long.UZERO).toString()
    }
    if (message.body?.$case === 'ackMsg') {
      obj.ackMsg = (message.body.ackMsg || Long.UZERO).toString()
    }
    return obj
  },

  create<I extends Exact<DeepPartial<SessionRequest>, I>>(
    base?: I,
  ): SessionRequest {
    return SessionRequest.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<SessionRequest>, I>>(
    object: I,
  ): SessionRequest {
    const message = createBaseSessionRequest()
    message.sessionSeqno =
      object.sessionSeqno !== undefined && object.sessionSeqno !== null
        ? Long.fromValue(object.sessionSeqno)
        : Long.UZERO
    if (
      object.body?.$case === 'init' &&
      object.body?.init !== undefined &&
      object.body?.init !== null
    ) {
      message.body = {
        $case: 'init',
        init: SessionInit.fromPartial(object.body.init),
      }
    }
    if (
      object.body?.$case === 'sendMsg' &&
      object.body?.sendMsg !== undefined &&
      object.body?.sendMsg !== null
    ) {
      message.body = {
        $case: 'sendMsg',
        sendMsg: SessionMsg.fromPartial(object.body.sendMsg),
      }
    }
    if (
      object.body?.$case === 'clearMsg' &&
      object.body?.clearMsg !== undefined &&
      object.body?.clearMsg !== null
    ) {
      message.body = {
        $case: 'clearMsg',
        clearMsg: Long.fromValue(object.body.clearMsg),
      }
    }
    if (
      object.body?.$case === 'ackMsg' &&
      object.body?.ackMsg !== undefined &&
      object.body?.ackMsg !== null
    ) {
      message.body = {
        $case: 'ackMsg',
        ackMsg: Long.fromValue(object.body.ackMsg),
      }
    }
    return message
  },
}

function createBaseSessionInit(): SessionInit {
  return { peerId: '' }
}

export const SessionInit = {
  encode(
    message: SessionInit,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.peerId !== '') {
      writer.uint32(10).string(message.peerId)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SessionInit {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSessionInit()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.peerId = reader.string()
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
  // Transform<SessionInit, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<SessionInit | SessionInit[]>
      | Iterable<SessionInit | SessionInit[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [SessionInit.encode(p).finish()]
        }
      } else {
        yield* [SessionInit.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, SessionInit>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<SessionInit> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [SessionInit.decode(p)]
        }
      } else {
        yield* [SessionInit.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): SessionInit {
    return {
      peerId: isSet(object.peerId) ? globalThis.String(object.peerId) : '',
    }
  },

  toJSON(message: SessionInit): unknown {
    const obj: any = {}
    if (message.peerId !== '') {
      obj.peerId = message.peerId
    }
    return obj
  },

  create<I extends Exact<DeepPartial<SessionInit>, I>>(base?: I): SessionInit {
    return SessionInit.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<SessionInit>, I>>(
    object: I,
  ): SessionInit {
    const message = createBaseSessionInit()
    message.peerId = object.peerId ?? ''
    return message
  },
}

function createBaseSessionMsg(): SessionMsg {
  return { signedMsg: undefined, seqno: Long.UZERO }
}

export const SessionMsg = {
  encode(
    message: SessionMsg,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.signedMsg !== undefined) {
      SignedMsg.encode(message.signedMsg, writer.uint32(10).fork()).ldelim()
    }
    if (!message.seqno.equals(Long.UZERO)) {
      writer.uint32(16).uint64(message.seqno)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SessionMsg {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSessionMsg()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.signedMsg = SignedMsg.decode(reader, reader.uint32())
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.seqno = reader.uint64() as Long
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
  // Transform<SessionMsg, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<SessionMsg | SessionMsg[]>
      | Iterable<SessionMsg | SessionMsg[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [SessionMsg.encode(p).finish()]
        }
      } else {
        yield* [SessionMsg.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, SessionMsg>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<SessionMsg> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [SessionMsg.decode(p)]
        }
      } else {
        yield* [SessionMsg.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): SessionMsg {
    return {
      signedMsg: isSet(object.signedMsg)
        ? SignedMsg.fromJSON(object.signedMsg)
        : undefined,
      seqno: isSet(object.seqno) ? Long.fromValue(object.seqno) : Long.UZERO,
    }
  },

  toJSON(message: SessionMsg): unknown {
    const obj: any = {}
    if (message.signedMsg !== undefined) {
      obj.signedMsg = SignedMsg.toJSON(message.signedMsg)
    }
    if (!message.seqno.equals(Long.UZERO)) {
      obj.seqno = (message.seqno || Long.UZERO).toString()
    }
    return obj
  },

  create<I extends Exact<DeepPartial<SessionMsg>, I>>(base?: I): SessionMsg {
    return SessionMsg.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<SessionMsg>, I>>(
    object: I,
  ): SessionMsg {
    const message = createBaseSessionMsg()
    message.signedMsg =
      object.signedMsg !== undefined && object.signedMsg !== null
        ? SignedMsg.fromPartial(object.signedMsg)
        : undefined
    message.seqno =
      object.seqno !== undefined && object.seqno !== null
        ? Long.fromValue(object.seqno)
        : Long.UZERO
    return message
  },
}

function createBaseSessionResponse(): SessionResponse {
  return { body: undefined }
}

export const SessionResponse = {
  encode(
    message: SessionResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    switch (message.body?.$case) {
      case 'opened':
        writer.uint32(8).uint64(message.body.opened)
        break
      case 'closed':
        writer.uint32(16).bool(message.body.closed)
        break
      case 'recvMsg':
        SessionMsg.encode(
          message.body.recvMsg,
          writer.uint32(26).fork(),
        ).ldelim()
        break
      case 'clearMsg':
        writer.uint32(32).uint64(message.body.clearMsg)
        break
      case 'ackMsg':
        writer.uint32(40).uint64(message.body.ackMsg)
        break
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SessionResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSessionResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.body = { $case: 'opened', opened: reader.uint64() as Long }
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.body = { $case: 'closed', closed: reader.bool() }
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.body = {
            $case: 'recvMsg',
            recvMsg: SessionMsg.decode(reader, reader.uint32()),
          }
          continue
        case 4:
          if (tag !== 32) {
            break
          }

          message.body = {
            $case: 'clearMsg',
            clearMsg: reader.uint64() as Long,
          }
          continue
        case 5:
          if (tag !== 40) {
            break
          }

          message.body = { $case: 'ackMsg', ackMsg: reader.uint64() as Long }
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
  // Transform<SessionResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<SessionResponse | SessionResponse[]>
      | Iterable<SessionResponse | SessionResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [SessionResponse.encode(p).finish()]
        }
      } else {
        yield* [SessionResponse.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, SessionResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<SessionResponse> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [SessionResponse.decode(p)]
        }
      } else {
        yield* [SessionResponse.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): SessionResponse {
    return {
      body: isSet(object.opened)
        ? { $case: 'opened', opened: Long.fromValue(object.opened) }
        : isSet(object.closed)
          ? { $case: 'closed', closed: globalThis.Boolean(object.closed) }
          : isSet(object.recvMsg)
            ? { $case: 'recvMsg', recvMsg: SessionMsg.fromJSON(object.recvMsg) }
            : isSet(object.clearMsg)
              ? { $case: 'clearMsg', clearMsg: Long.fromValue(object.clearMsg) }
              : isSet(object.ackMsg)
                ? { $case: 'ackMsg', ackMsg: Long.fromValue(object.ackMsg) }
                : undefined,
    }
  },

  toJSON(message: SessionResponse): unknown {
    const obj: any = {}
    if (message.body?.$case === 'opened') {
      obj.opened = (message.body.opened || Long.UZERO).toString()
    }
    if (message.body?.$case === 'closed') {
      obj.closed = message.body.closed
    }
    if (message.body?.$case === 'recvMsg') {
      obj.recvMsg = SessionMsg.toJSON(message.body.recvMsg)
    }
    if (message.body?.$case === 'clearMsg') {
      obj.clearMsg = (message.body.clearMsg || Long.UZERO).toString()
    }
    if (message.body?.$case === 'ackMsg') {
      obj.ackMsg = (message.body.ackMsg || Long.UZERO).toString()
    }
    return obj
  },

  create<I extends Exact<DeepPartial<SessionResponse>, I>>(
    base?: I,
  ): SessionResponse {
    return SessionResponse.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<SessionResponse>, I>>(
    object: I,
  ): SessionResponse {
    const message = createBaseSessionResponse()
    if (
      object.body?.$case === 'opened' &&
      object.body?.opened !== undefined &&
      object.body?.opened !== null
    ) {
      message.body = {
        $case: 'opened',
        opened: Long.fromValue(object.body.opened),
      }
    }
    if (
      object.body?.$case === 'closed' &&
      object.body?.closed !== undefined &&
      object.body?.closed !== null
    ) {
      message.body = { $case: 'closed', closed: object.body.closed }
    }
    if (
      object.body?.$case === 'recvMsg' &&
      object.body?.recvMsg !== undefined &&
      object.body?.recvMsg !== null
    ) {
      message.body = {
        $case: 'recvMsg',
        recvMsg: SessionMsg.fromPartial(object.body.recvMsg),
      }
    }
    if (
      object.body?.$case === 'clearMsg' &&
      object.body?.clearMsg !== undefined &&
      object.body?.clearMsg !== null
    ) {
      message.body = {
        $case: 'clearMsg',
        clearMsg: Long.fromValue(object.body.clearMsg),
      }
    }
    if (
      object.body?.$case === 'ackMsg' &&
      object.body?.ackMsg !== undefined &&
      object.body?.ackMsg !== null
    ) {
      message.body = {
        $case: 'ackMsg',
        ackMsg: Long.fromValue(object.body.ackMsg),
      }
    }
    return message
  },
}

/** Signaling is a service which allows peers to signal each other via a RPC server. */
export interface Signaling {
  /** Listen waits for messages to be available in our inbox from remote peers. */
  Listen(
    request: ListenRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<ListenResponse>
  /** Session opens a signaling session to send and recv messages from a remote peer. */
  Session(
    request: AsyncIterable<SessionRequest>,
    abortSignal?: AbortSignal,
  ): AsyncIterable<SessionResponse>
}

export const SignalingServiceName = 'signaling.rpc.Signaling'
export class SignalingClientImpl implements Signaling {
  private readonly rpc: Rpc
  private readonly service: string
  constructor(rpc: Rpc, opts?: { service?: string }) {
    this.service = opts?.service || SignalingServiceName
    this.rpc = rpc
    this.Listen = this.Listen.bind(this)
    this.Session = this.Session.bind(this)
  }
  Listen(
    request: ListenRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<ListenResponse> {
    const data = ListenRequest.encode(request).finish()
    const result = this.rpc.serverStreamingRequest(
      this.service,
      'Listen',
      data,
      abortSignal || undefined,
    )
    return ListenResponse.decodeTransform(result)
  }

  Session(
    request: AsyncIterable<SessionRequest>,
    abortSignal?: AbortSignal,
  ): AsyncIterable<SessionResponse> {
    const data = SessionRequest.encodeTransform(request)
    const result = this.rpc.bidirectionalStreamingRequest(
      this.service,
      'Session',
      data,
      abortSignal || undefined,
    )
    return SessionResponse.decodeTransform(result)
  }
}

/** Signaling is a service which allows peers to signal each other via a RPC server. */
export type SignalingDefinition = typeof SignalingDefinition
export const SignalingDefinition = {
  name: 'Signaling',
  fullName: 'signaling.rpc.Signaling',
  methods: {
    /** Listen waits for messages to be available in our inbox from remote peers. */
    listen: {
      name: 'Listen',
      requestType: ListenRequest,
      requestStream: false,
      responseType: ListenResponse,
      responseStream: true,
      options: {},
    },
    /** Session opens a signaling session to send and recv messages from a remote peer. */
    session: {
      name: 'Session',
      requestType: SessionRequest,
      requestStream: true,
      responseType: SessionResponse,
      responseStream: true,
      options: {},
    },
  },
} as const

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array,
    abortSignal?: AbortSignal,
  ): Promise<Uint8Array>
  clientStreamingRequest(
    service: string,
    method: string,
    data: AsyncIterable<Uint8Array>,
    abortSignal?: AbortSignal,
  ): Promise<Uint8Array>
  serverStreamingRequest(
    service: string,
    method: string,
    data: Uint8Array,
    abortSignal?: AbortSignal,
  ): AsyncIterable<Uint8Array>
  bidirectionalStreamingRequest(
    service: string,
    method: string,
    data: AsyncIterable<Uint8Array>,
    abortSignal?: AbortSignal,
  ): AsyncIterable<Uint8Array>
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
