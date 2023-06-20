/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'pubsub.api'

/** SubcribeRequest is a pubsub subscription request message. */
export interface SubscribeRequest {
  /**
   * ChannelId is the channel id to subscribe to.
   * Must be sent before / with publish.
   * Cannot change the channel ID after first transmission.
   */
  channelId: string
  /**
   * PeerId is the peer identifier of the publisher/subscriber.
   * The peer ID will be used to acquire the peer private key.
   */
  peerId: string
  /**
   * PrivKeyPem is an alternate to PeerId, specify private key inline.
   * Overrides PeerId if set.
   */
  privKeyPem: string
  /** PublishRequest contains a publish message request. */
  publishRequest: PublishRequest | undefined
}

/** PublishRequest is a message published via the subscribe channel. */
export interface PublishRequest {
  /** Data is the published data. */
  data: Uint8Array
  /**
   * Identifier is a uint32 identifier to use for outgoing status.
   * If zero, no outgoing status response will be sent.
   */
  identifier: number
}

/** SubcribeResponse is a pubsub subscription response message. */
export interface SubscribeResponse {
  /** IncomingMessage is an incoming message. */
  incomingMessage: IncomingMessage | undefined
  /**
   * OutgoingStatus is status of an outgoing message.
   * Sent when a Publish request finishes.
   */
  outgoingStatus: OutgoingStatus | undefined
  /** SubscriptionStatus is the status of the subscription */
  subscriptionStatus: SubscriptionStatus | undefined
}

/** SubscripionStatus is the status of the subscription handle. */
export interface SubscriptionStatus {
  /** Subscribed indicates the subscription is established. */
  subscribed: boolean
}

/** OutgoingStatus is status of an outgoing message. */
export interface OutgoingStatus {
  /** Identifier is the request-provided identifier for the message. */
  identifier: number
  /** Sent indicates if the message was sent. */
  sent: boolean
}

/** IncomingMessage implements Message with a proto object. */
export interface IncomingMessage {
  /** FromPeerId is the peer identifier of the sender. */
  fromPeerId: string
  /** Authenticated indicates if the message is verified to be from the sender. */
  authenticated: boolean
  /** Data is the inner data. */
  data: Uint8Array
}

function createBaseSubscribeRequest(): SubscribeRequest {
  return {
    channelId: '',
    peerId: '',
    privKeyPem: '',
    publishRequest: undefined,
  }
}

export const SubscribeRequest = {
  encode(
    message: SubscribeRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.channelId !== '') {
      writer.uint32(10).string(message.channelId)
    }
    if (message.peerId !== '') {
      writer.uint32(18).string(message.peerId)
    }
    if (message.privKeyPem !== '') {
      writer.uint32(26).string(message.privKeyPem)
    }
    if (message.publishRequest !== undefined) {
      PublishRequest.encode(
        message.publishRequest,
        writer.uint32(34).fork()
      ).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SubscribeRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSubscribeRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.channelId = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.peerId = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.privKeyPem = reader.string()
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          message.publishRequest = PublishRequest.decode(
            reader,
            reader.uint32()
          )
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
  // Transform<SubscribeRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<SubscribeRequest | SubscribeRequest[]>
      | Iterable<SubscribeRequest | SubscribeRequest[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SubscribeRequest.encode(p).finish()]
        }
      } else {
        yield* [SubscribeRequest.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, SubscribeRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<SubscribeRequest> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SubscribeRequest.decode(p)]
        }
      } else {
        yield* [SubscribeRequest.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): SubscribeRequest {
    return {
      channelId: isSet(object.channelId) ? String(object.channelId) : '',
      peerId: isSet(object.peerId) ? String(object.peerId) : '',
      privKeyPem: isSet(object.privKeyPem) ? String(object.privKeyPem) : '',
      publishRequest: isSet(object.publishRequest)
        ? PublishRequest.fromJSON(object.publishRequest)
        : undefined,
    }
  },

  toJSON(message: SubscribeRequest): unknown {
    const obj: any = {}
    message.channelId !== undefined && (obj.channelId = message.channelId)
    message.peerId !== undefined && (obj.peerId = message.peerId)
    message.privKeyPem !== undefined && (obj.privKeyPem = message.privKeyPem)
    message.publishRequest !== undefined &&
      (obj.publishRequest = message.publishRequest
        ? PublishRequest.toJSON(message.publishRequest)
        : undefined)
    return obj
  },

  create<I extends Exact<DeepPartial<SubscribeRequest>, I>>(
    base?: I
  ): SubscribeRequest {
    return SubscribeRequest.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<SubscribeRequest>, I>>(
    object: I
  ): SubscribeRequest {
    const message = createBaseSubscribeRequest()
    message.channelId = object.channelId ?? ''
    message.peerId = object.peerId ?? ''
    message.privKeyPem = object.privKeyPem ?? ''
    message.publishRequest =
      object.publishRequest !== undefined && object.publishRequest !== null
        ? PublishRequest.fromPartial(object.publishRequest)
        : undefined
    return message
  },
}

function createBasePublishRequest(): PublishRequest {
  return { data: new Uint8Array(0), identifier: 0 }
}

export const PublishRequest = {
  encode(
    message: PublishRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.data.length !== 0) {
      writer.uint32(10).bytes(message.data)
    }
    if (message.identifier !== 0) {
      writer.uint32(16).uint32(message.identifier)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PublishRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePublishRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.data = reader.bytes()
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.identifier = reader.uint32()
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
  // Transform<PublishRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<PublishRequest | PublishRequest[]>
      | Iterable<PublishRequest | PublishRequest[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [PublishRequest.encode(p).finish()]
        }
      } else {
        yield* [PublishRequest.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, PublishRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<PublishRequest> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [PublishRequest.decode(p)]
        }
      } else {
        yield* [PublishRequest.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): PublishRequest {
    return {
      data: isSet(object.data)
        ? bytesFromBase64(object.data)
        : new Uint8Array(0),
      identifier: isSet(object.identifier) ? Number(object.identifier) : 0,
    }
  },

  toJSON(message: PublishRequest): unknown {
    const obj: any = {}
    message.data !== undefined &&
      (obj.data = base64FromBytes(
        message.data !== undefined ? message.data : new Uint8Array(0)
      ))
    message.identifier !== undefined &&
      (obj.identifier = Math.round(message.identifier))
    return obj
  },

  create<I extends Exact<DeepPartial<PublishRequest>, I>>(
    base?: I
  ): PublishRequest {
    return PublishRequest.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<PublishRequest>, I>>(
    object: I
  ): PublishRequest {
    const message = createBasePublishRequest()
    message.data = object.data ?? new Uint8Array(0)
    message.identifier = object.identifier ?? 0
    return message
  },
}

function createBaseSubscribeResponse(): SubscribeResponse {
  return {
    incomingMessage: undefined,
    outgoingStatus: undefined,
    subscriptionStatus: undefined,
  }
}

export const SubscribeResponse = {
  encode(
    message: SubscribeResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.incomingMessage !== undefined) {
      IncomingMessage.encode(
        message.incomingMessage,
        writer.uint32(10).fork()
      ).ldelim()
    }
    if (message.outgoingStatus !== undefined) {
      OutgoingStatus.encode(
        message.outgoingStatus,
        writer.uint32(18).fork()
      ).ldelim()
    }
    if (message.subscriptionStatus !== undefined) {
      SubscriptionStatus.encode(
        message.subscriptionStatus,
        writer.uint32(26).fork()
      ).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SubscribeResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSubscribeResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.incomingMessage = IncomingMessage.decode(
            reader,
            reader.uint32()
          )
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.outgoingStatus = OutgoingStatus.decode(
            reader,
            reader.uint32()
          )
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.subscriptionStatus = SubscriptionStatus.decode(
            reader,
            reader.uint32()
          )
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
  // Transform<SubscribeResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<SubscribeResponse | SubscribeResponse[]>
      | Iterable<SubscribeResponse | SubscribeResponse[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SubscribeResponse.encode(p).finish()]
        }
      } else {
        yield* [SubscribeResponse.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, SubscribeResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<SubscribeResponse> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SubscribeResponse.decode(p)]
        }
      } else {
        yield* [SubscribeResponse.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): SubscribeResponse {
    return {
      incomingMessage: isSet(object.incomingMessage)
        ? IncomingMessage.fromJSON(object.incomingMessage)
        : undefined,
      outgoingStatus: isSet(object.outgoingStatus)
        ? OutgoingStatus.fromJSON(object.outgoingStatus)
        : undefined,
      subscriptionStatus: isSet(object.subscriptionStatus)
        ? SubscriptionStatus.fromJSON(object.subscriptionStatus)
        : undefined,
    }
  },

  toJSON(message: SubscribeResponse): unknown {
    const obj: any = {}
    message.incomingMessage !== undefined &&
      (obj.incomingMessage = message.incomingMessage
        ? IncomingMessage.toJSON(message.incomingMessage)
        : undefined)
    message.outgoingStatus !== undefined &&
      (obj.outgoingStatus = message.outgoingStatus
        ? OutgoingStatus.toJSON(message.outgoingStatus)
        : undefined)
    message.subscriptionStatus !== undefined &&
      (obj.subscriptionStatus = message.subscriptionStatus
        ? SubscriptionStatus.toJSON(message.subscriptionStatus)
        : undefined)
    return obj
  },

  create<I extends Exact<DeepPartial<SubscribeResponse>, I>>(
    base?: I
  ): SubscribeResponse {
    return SubscribeResponse.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<SubscribeResponse>, I>>(
    object: I
  ): SubscribeResponse {
    const message = createBaseSubscribeResponse()
    message.incomingMessage =
      object.incomingMessage !== undefined && object.incomingMessage !== null
        ? IncomingMessage.fromPartial(object.incomingMessage)
        : undefined
    message.outgoingStatus =
      object.outgoingStatus !== undefined && object.outgoingStatus !== null
        ? OutgoingStatus.fromPartial(object.outgoingStatus)
        : undefined
    message.subscriptionStatus =
      object.subscriptionStatus !== undefined &&
      object.subscriptionStatus !== null
        ? SubscriptionStatus.fromPartial(object.subscriptionStatus)
        : undefined
    return message
  },
}

function createBaseSubscriptionStatus(): SubscriptionStatus {
  return { subscribed: false }
}

export const SubscriptionStatus = {
  encode(
    message: SubscriptionStatus,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.subscribed === true) {
      writer.uint32(8).bool(message.subscribed)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SubscriptionStatus {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseSubscriptionStatus()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.subscribed = reader.bool()
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
  // Transform<SubscriptionStatus, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<SubscriptionStatus | SubscriptionStatus[]>
      | Iterable<SubscriptionStatus | SubscriptionStatus[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SubscriptionStatus.encode(p).finish()]
        }
      } else {
        yield* [SubscriptionStatus.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, SubscriptionStatus>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<SubscriptionStatus> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [SubscriptionStatus.decode(p)]
        }
      } else {
        yield* [SubscriptionStatus.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): SubscriptionStatus {
    return {
      subscribed: isSet(object.subscribed) ? Boolean(object.subscribed) : false,
    }
  },

  toJSON(message: SubscriptionStatus): unknown {
    const obj: any = {}
    message.subscribed !== undefined && (obj.subscribed = message.subscribed)
    return obj
  },

  create<I extends Exact<DeepPartial<SubscriptionStatus>, I>>(
    base?: I
  ): SubscriptionStatus {
    return SubscriptionStatus.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<SubscriptionStatus>, I>>(
    object: I
  ): SubscriptionStatus {
    const message = createBaseSubscriptionStatus()
    message.subscribed = object.subscribed ?? false
    return message
  },
}

function createBaseOutgoingStatus(): OutgoingStatus {
  return { identifier: 0, sent: false }
}

export const OutgoingStatus = {
  encode(
    message: OutgoingStatus,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.identifier !== 0) {
      writer.uint32(8).uint32(message.identifier)
    }
    if (message.sent === true) {
      writer.uint32(16).bool(message.sent)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OutgoingStatus {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseOutgoingStatus()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.identifier = reader.uint32()
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.sent = reader.bool()
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
  // Transform<OutgoingStatus, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<OutgoingStatus | OutgoingStatus[]>
      | Iterable<OutgoingStatus | OutgoingStatus[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [OutgoingStatus.encode(p).finish()]
        }
      } else {
        yield* [OutgoingStatus.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, OutgoingStatus>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<OutgoingStatus> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [OutgoingStatus.decode(p)]
        }
      } else {
        yield* [OutgoingStatus.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): OutgoingStatus {
    return {
      identifier: isSet(object.identifier) ? Number(object.identifier) : 0,
      sent: isSet(object.sent) ? Boolean(object.sent) : false,
    }
  },

  toJSON(message: OutgoingStatus): unknown {
    const obj: any = {}
    message.identifier !== undefined &&
      (obj.identifier = Math.round(message.identifier))
    message.sent !== undefined && (obj.sent = message.sent)
    return obj
  },

  create<I extends Exact<DeepPartial<OutgoingStatus>, I>>(
    base?: I
  ): OutgoingStatus {
    return OutgoingStatus.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<OutgoingStatus>, I>>(
    object: I
  ): OutgoingStatus {
    const message = createBaseOutgoingStatus()
    message.identifier = object.identifier ?? 0
    message.sent = object.sent ?? false
    return message
  },
}

function createBaseIncomingMessage(): IncomingMessage {
  return { fromPeerId: '', authenticated: false, data: new Uint8Array(0) }
}

export const IncomingMessage = {
  encode(
    message: IncomingMessage,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.fromPeerId !== '') {
      writer.uint32(10).string(message.fromPeerId)
    }
    if (message.authenticated === true) {
      writer.uint32(16).bool(message.authenticated)
    }
    if (message.data.length !== 0) {
      writer.uint32(26).bytes(message.data)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): IncomingMessage {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseIncomingMessage()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.fromPeerId = reader.string()
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.authenticated = reader.bool()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.data = reader.bytes()
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
  // Transform<IncomingMessage, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<IncomingMessage | IncomingMessage[]>
      | Iterable<IncomingMessage | IncomingMessage[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [IncomingMessage.encode(p).finish()]
        }
      } else {
        yield* [IncomingMessage.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, IncomingMessage>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<IncomingMessage> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [IncomingMessage.decode(p)]
        }
      } else {
        yield* [IncomingMessage.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): IncomingMessage {
    return {
      fromPeerId: isSet(object.fromPeerId) ? String(object.fromPeerId) : '',
      authenticated: isSet(object.authenticated)
        ? Boolean(object.authenticated)
        : false,
      data: isSet(object.data)
        ? bytesFromBase64(object.data)
        : new Uint8Array(0),
    }
  },

  toJSON(message: IncomingMessage): unknown {
    const obj: any = {}
    message.fromPeerId !== undefined && (obj.fromPeerId = message.fromPeerId)
    message.authenticated !== undefined &&
      (obj.authenticated = message.authenticated)
    message.data !== undefined &&
      (obj.data = base64FromBytes(
        message.data !== undefined ? message.data : new Uint8Array(0)
      ))
    return obj
  },

  create<I extends Exact<DeepPartial<IncomingMessage>, I>>(
    base?: I
  ): IncomingMessage {
    return IncomingMessage.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<IncomingMessage>, I>>(
    object: I
  ): IncomingMessage {
    const message = createBaseIncomingMessage()
    message.fromPeerId = object.fromPeerId ?? ''
    message.authenticated = object.authenticated ?? false
    message.data = object.data ?? new Uint8Array(0)
    return message
  },
}

/** PubSubService is the bifrost pubsub service. */
export interface PubSubService {
  /**
   * Subscribe subscribes to a channel, allowing the subscriber to publish
   * messages over the same channel.
   */
  Subscribe(
    request: AsyncIterable<SubscribeRequest>,
    abortSignal?: AbortSignal
  ): AsyncIterable<SubscribeResponse>
}

export const PubSubServiceServiceName = 'pubsub.api.PubSubService'
export class PubSubServiceClientImpl implements PubSubService {
  private readonly rpc: Rpc
  private readonly service: string
  constructor(rpc: Rpc, opts?: { service?: string }) {
    this.service = opts?.service || PubSubServiceServiceName
    this.rpc = rpc
    this.Subscribe = this.Subscribe.bind(this)
  }
  Subscribe(
    request: AsyncIterable<SubscribeRequest>,
    abortSignal?: AbortSignal
  ): AsyncIterable<SubscribeResponse> {
    const data = SubscribeRequest.encodeTransform(request)
    const result = this.rpc.bidirectionalStreamingRequest(
      this.service,
      'Subscribe',
      data,
      abortSignal || undefined
    )
    return SubscribeResponse.decodeTransform(result)
  }
}

/** PubSubService is the bifrost pubsub service. */
export type PubSubServiceDefinition = typeof PubSubServiceDefinition
export const PubSubServiceDefinition = {
  name: 'PubSubService',
  fullName: 'pubsub.api.PubSubService',
  methods: {
    /**
     * Subscribe subscribes to a channel, allowing the subscriber to publish
     * messages over the same channel.
     */
    subscribe: {
      name: 'Subscribe',
      requestType: SubscribeRequest,
      requestStream: true,
      responseType: SubscribeResponse,
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
    abortSignal?: AbortSignal
  ): Promise<Uint8Array>
  clientStreamingRequest(
    service: string,
    method: string,
    data: AsyncIterable<Uint8Array>,
    abortSignal?: AbortSignal
  ): Promise<Uint8Array>
  serverStreamingRequest(
    service: string,
    method: string,
    data: Uint8Array,
    abortSignal?: AbortSignal
  ): AsyncIterable<Uint8Array>
  bidirectionalStreamingRequest(
    service: string,
    method: string,
    data: AsyncIterable<Uint8Array>,
    abortSignal?: AbortSignal
  ): AsyncIterable<Uint8Array>
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
