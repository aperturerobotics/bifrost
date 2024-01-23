/* eslint-disable */
import {
  ControllerStatus,
  controllerStatusFromJSON,
  controllerStatusToJSON,
} from '@go/github.com/aperturerobotics/controllerbus/controller/exec/exec.pb.js'
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { Config } from '../forwarding/forwarding.pb.js'
import { Config as Config1 } from '../listening/listening.pb.js'
import { Config as Config2 } from './accept/accept.pb.js'
import { Config as Config3 } from './dial/dial.pb.js'
import { Data } from './rpc/rpc.pb.js'

export const protobufPackage = 'stream.api'

/** ForwardStreamsRequest is the request type for ForwardStreams. */
export interface ForwardStreamsRequest {
  forwardingConfig: Config | undefined
}

/** ForwardStreamsResponse is the response type for ForwardStreams. */
export interface ForwardStreamsResponse {
  /** ControllerStatus is the status of the forwarding controller. */
  controllerStatus: ControllerStatus
}

/** ListenStreamsRequest is the request type for ListenStreams. */
export interface ListenStreamsRequest {
  listeningConfig: Config1 | undefined
}

/** ListenStreamsResponse is the response type for ListenStreams. */
export interface ListenStreamsResponse {
  /** ControllerStatus is the status of the forwarding controller. */
  controllerStatus: ControllerStatus
}

/** AcceptStreamRequest is the request type for AcceptStream. */
export interface AcceptStreamRequest {
  /**
   * Config is the configuration for the accept.
   * The first packet will contain this value.
   */
  config: Config2 | undefined
  /** Data is a data packet. */
  data: Data | undefined
}

/** AcceptStreamResponse is the response type for AcceptStream. */
export interface AcceptStreamResponse {
  /** Data is a data packet. */
  data: Data | undefined
}

/** DialStreamRequest is the request type for DialStream. */
export interface DialStreamRequest {
  /**
   * Config is the configuration for the dial.
   * The first packet will contain this value.
   */
  config: Config3 | undefined
  /** Data is a data packet. */
  data: Data | undefined
}

/** DialStreamResponse is the response type for DialStream. */
export interface DialStreamResponse {
  /** Data is a data packet. */
  data: Data | undefined
}

function createBaseForwardStreamsRequest(): ForwardStreamsRequest {
  return { forwardingConfig: undefined }
}

export const ForwardStreamsRequest = {
  encode(
    message: ForwardStreamsRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.forwardingConfig !== undefined) {
      Config.encode(message.forwardingConfig, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): ForwardStreamsRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseForwardStreamsRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.forwardingConfig = Config.decode(reader, reader.uint32())
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
  // Transform<ForwardStreamsRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ForwardStreamsRequest | ForwardStreamsRequest[]>
      | Iterable<ForwardStreamsRequest | ForwardStreamsRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ForwardStreamsRequest.encode(p).finish()]
        }
      } else {
        yield* [ForwardStreamsRequest.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ForwardStreamsRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<ForwardStreamsRequest> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ForwardStreamsRequest.decode(p)]
        }
      } else {
        yield* [ForwardStreamsRequest.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): ForwardStreamsRequest {
    return {
      forwardingConfig: isSet(object.forwardingConfig)
        ? Config.fromJSON(object.forwardingConfig)
        : undefined,
    }
  },

  toJSON(message: ForwardStreamsRequest): unknown {
    const obj: any = {}
    if (message.forwardingConfig !== undefined) {
      obj.forwardingConfig = Config.toJSON(message.forwardingConfig)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<ForwardStreamsRequest>, I>>(
    base?: I,
  ): ForwardStreamsRequest {
    return ForwardStreamsRequest.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<ForwardStreamsRequest>, I>>(
    object: I,
  ): ForwardStreamsRequest {
    const message = createBaseForwardStreamsRequest()
    message.forwardingConfig =
      object.forwardingConfig !== undefined && object.forwardingConfig !== null
        ? Config.fromPartial(object.forwardingConfig)
        : undefined
    return message
  },
}

function createBaseForwardStreamsResponse(): ForwardStreamsResponse {
  return { controllerStatus: 0 }
}

export const ForwardStreamsResponse = {
  encode(
    message: ForwardStreamsResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.controllerStatus !== 0) {
      writer.uint32(8).int32(message.controllerStatus)
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): ForwardStreamsResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseForwardStreamsResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.controllerStatus = reader.int32() as any
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
  // Transform<ForwardStreamsResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ForwardStreamsResponse | ForwardStreamsResponse[]>
      | Iterable<ForwardStreamsResponse | ForwardStreamsResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ForwardStreamsResponse.encode(p).finish()]
        }
      } else {
        yield* [ForwardStreamsResponse.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ForwardStreamsResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<ForwardStreamsResponse> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ForwardStreamsResponse.decode(p)]
        }
      } else {
        yield* [ForwardStreamsResponse.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): ForwardStreamsResponse {
    return {
      controllerStatus: isSet(object.controllerStatus)
        ? controllerStatusFromJSON(object.controllerStatus)
        : 0,
    }
  },

  toJSON(message: ForwardStreamsResponse): unknown {
    const obj: any = {}
    if (message.controllerStatus !== 0) {
      obj.controllerStatus = controllerStatusToJSON(message.controllerStatus)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<ForwardStreamsResponse>, I>>(
    base?: I,
  ): ForwardStreamsResponse {
    return ForwardStreamsResponse.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<ForwardStreamsResponse>, I>>(
    object: I,
  ): ForwardStreamsResponse {
    const message = createBaseForwardStreamsResponse()
    message.controllerStatus = object.controllerStatus ?? 0
    return message
  },
}

function createBaseListenStreamsRequest(): ListenStreamsRequest {
  return { listeningConfig: undefined }
}

export const ListenStreamsRequest = {
  encode(
    message: ListenStreamsRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.listeningConfig !== undefined) {
      Config1.encode(message.listeningConfig, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): ListenStreamsRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseListenStreamsRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.listeningConfig = Config1.decode(reader, reader.uint32())
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
  // Transform<ListenStreamsRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ListenStreamsRequest | ListenStreamsRequest[]>
      | Iterable<ListenStreamsRequest | ListenStreamsRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ListenStreamsRequest.encode(p).finish()]
        }
      } else {
        yield* [ListenStreamsRequest.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ListenStreamsRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<ListenStreamsRequest> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ListenStreamsRequest.decode(p)]
        }
      } else {
        yield* [ListenStreamsRequest.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): ListenStreamsRequest {
    return {
      listeningConfig: isSet(object.listeningConfig)
        ? Config1.fromJSON(object.listeningConfig)
        : undefined,
    }
  },

  toJSON(message: ListenStreamsRequest): unknown {
    const obj: any = {}
    if (message.listeningConfig !== undefined) {
      obj.listeningConfig = Config1.toJSON(message.listeningConfig)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<ListenStreamsRequest>, I>>(
    base?: I,
  ): ListenStreamsRequest {
    return ListenStreamsRequest.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<ListenStreamsRequest>, I>>(
    object: I,
  ): ListenStreamsRequest {
    const message = createBaseListenStreamsRequest()
    message.listeningConfig =
      object.listeningConfig !== undefined && object.listeningConfig !== null
        ? Config1.fromPartial(object.listeningConfig)
        : undefined
    return message
  },
}

function createBaseListenStreamsResponse(): ListenStreamsResponse {
  return { controllerStatus: 0 }
}

export const ListenStreamsResponse = {
  encode(
    message: ListenStreamsResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.controllerStatus !== 0) {
      writer.uint32(8).int32(message.controllerStatus)
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): ListenStreamsResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseListenStreamsResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.controllerStatus = reader.int32() as any
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
  // Transform<ListenStreamsResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<ListenStreamsResponse | ListenStreamsResponse[]>
      | Iterable<ListenStreamsResponse | ListenStreamsResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ListenStreamsResponse.encode(p).finish()]
        }
      } else {
        yield* [ListenStreamsResponse.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, ListenStreamsResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<ListenStreamsResponse> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [ListenStreamsResponse.decode(p)]
        }
      } else {
        yield* [ListenStreamsResponse.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): ListenStreamsResponse {
    return {
      controllerStatus: isSet(object.controllerStatus)
        ? controllerStatusFromJSON(object.controllerStatus)
        : 0,
    }
  },

  toJSON(message: ListenStreamsResponse): unknown {
    const obj: any = {}
    if (message.controllerStatus !== 0) {
      obj.controllerStatus = controllerStatusToJSON(message.controllerStatus)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<ListenStreamsResponse>, I>>(
    base?: I,
  ): ListenStreamsResponse {
    return ListenStreamsResponse.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<ListenStreamsResponse>, I>>(
    object: I,
  ): ListenStreamsResponse {
    const message = createBaseListenStreamsResponse()
    message.controllerStatus = object.controllerStatus ?? 0
    return message
  },
}

function createBaseAcceptStreamRequest(): AcceptStreamRequest {
  return { config: undefined, data: undefined }
}

export const AcceptStreamRequest = {
  encode(
    message: AcceptStreamRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.config !== undefined) {
      Config2.encode(message.config, writer.uint32(10).fork()).ldelim()
    }
    if (message.data !== undefined) {
      Data.encode(message.data, writer.uint32(18).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AcceptStreamRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseAcceptStreamRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.config = Config2.decode(reader, reader.uint32())
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.data = Data.decode(reader, reader.uint32())
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
  // Transform<AcceptStreamRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<AcceptStreamRequest | AcceptStreamRequest[]>
      | Iterable<AcceptStreamRequest | AcceptStreamRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [AcceptStreamRequest.encode(p).finish()]
        }
      } else {
        yield* [AcceptStreamRequest.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, AcceptStreamRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<AcceptStreamRequest> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [AcceptStreamRequest.decode(p)]
        }
      } else {
        yield* [AcceptStreamRequest.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): AcceptStreamRequest {
    return {
      config: isSet(object.config)
        ? Config2.fromJSON(object.config)
        : undefined,
      data: isSet(object.data) ? Data.fromJSON(object.data) : undefined,
    }
  },

  toJSON(message: AcceptStreamRequest): unknown {
    const obj: any = {}
    if (message.config !== undefined) {
      obj.config = Config2.toJSON(message.config)
    }
    if (message.data !== undefined) {
      obj.data = Data.toJSON(message.data)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<AcceptStreamRequest>, I>>(
    base?: I,
  ): AcceptStreamRequest {
    return AcceptStreamRequest.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<AcceptStreamRequest>, I>>(
    object: I,
  ): AcceptStreamRequest {
    const message = createBaseAcceptStreamRequest()
    message.config =
      object.config !== undefined && object.config !== null
        ? Config2.fromPartial(object.config)
        : undefined
    message.data =
      object.data !== undefined && object.data !== null
        ? Data.fromPartial(object.data)
        : undefined
    return message
  },
}

function createBaseAcceptStreamResponse(): AcceptStreamResponse {
  return { data: undefined }
}

export const AcceptStreamResponse = {
  encode(
    message: AcceptStreamResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.data !== undefined) {
      Data.encode(message.data, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): AcceptStreamResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseAcceptStreamResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.data = Data.decode(reader, reader.uint32())
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
  // Transform<AcceptStreamResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<AcceptStreamResponse | AcceptStreamResponse[]>
      | Iterable<AcceptStreamResponse | AcceptStreamResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [AcceptStreamResponse.encode(p).finish()]
        }
      } else {
        yield* [AcceptStreamResponse.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, AcceptStreamResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<AcceptStreamResponse> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [AcceptStreamResponse.decode(p)]
        }
      } else {
        yield* [AcceptStreamResponse.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): AcceptStreamResponse {
    return { data: isSet(object.data) ? Data.fromJSON(object.data) : undefined }
  },

  toJSON(message: AcceptStreamResponse): unknown {
    const obj: any = {}
    if (message.data !== undefined) {
      obj.data = Data.toJSON(message.data)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<AcceptStreamResponse>, I>>(
    base?: I,
  ): AcceptStreamResponse {
    return AcceptStreamResponse.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<AcceptStreamResponse>, I>>(
    object: I,
  ): AcceptStreamResponse {
    const message = createBaseAcceptStreamResponse()
    message.data =
      object.data !== undefined && object.data !== null
        ? Data.fromPartial(object.data)
        : undefined
    return message
  },
}

function createBaseDialStreamRequest(): DialStreamRequest {
  return { config: undefined, data: undefined }
}

export const DialStreamRequest = {
  encode(
    message: DialStreamRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.config !== undefined) {
      Config3.encode(message.config, writer.uint32(10).fork()).ldelim()
    }
    if (message.data !== undefined) {
      Data.encode(message.data, writer.uint32(18).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DialStreamRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseDialStreamRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.config = Config3.decode(reader, reader.uint32())
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.data = Data.decode(reader, reader.uint32())
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
  // Transform<DialStreamRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<DialStreamRequest | DialStreamRequest[]>
      | Iterable<DialStreamRequest | DialStreamRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [DialStreamRequest.encode(p).finish()]
        }
      } else {
        yield* [DialStreamRequest.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, DialStreamRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<DialStreamRequest> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [DialStreamRequest.decode(p)]
        }
      } else {
        yield* [DialStreamRequest.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): DialStreamRequest {
    return {
      config: isSet(object.config)
        ? Config3.fromJSON(object.config)
        : undefined,
      data: isSet(object.data) ? Data.fromJSON(object.data) : undefined,
    }
  },

  toJSON(message: DialStreamRequest): unknown {
    const obj: any = {}
    if (message.config !== undefined) {
      obj.config = Config3.toJSON(message.config)
    }
    if (message.data !== undefined) {
      obj.data = Data.toJSON(message.data)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<DialStreamRequest>, I>>(
    base?: I,
  ): DialStreamRequest {
    return DialStreamRequest.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<DialStreamRequest>, I>>(
    object: I,
  ): DialStreamRequest {
    const message = createBaseDialStreamRequest()
    message.config =
      object.config !== undefined && object.config !== null
        ? Config3.fromPartial(object.config)
        : undefined
    message.data =
      object.data !== undefined && object.data !== null
        ? Data.fromPartial(object.data)
        : undefined
    return message
  },
}

function createBaseDialStreamResponse(): DialStreamResponse {
  return { data: undefined }
}

export const DialStreamResponse = {
  encode(
    message: DialStreamResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.data !== undefined) {
      Data.encode(message.data, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DialStreamResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseDialStreamResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.data = Data.decode(reader, reader.uint32())
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
  // Transform<DialStreamResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<DialStreamResponse | DialStreamResponse[]>
      | Iterable<DialStreamResponse | DialStreamResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [DialStreamResponse.encode(p).finish()]
        }
      } else {
        yield* [DialStreamResponse.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, DialStreamResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<DialStreamResponse> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [DialStreamResponse.decode(p)]
        }
      } else {
        yield* [DialStreamResponse.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): DialStreamResponse {
    return { data: isSet(object.data) ? Data.fromJSON(object.data) : undefined }
  },

  toJSON(message: DialStreamResponse): unknown {
    const obj: any = {}
    if (message.data !== undefined) {
      obj.data = Data.toJSON(message.data)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<DialStreamResponse>, I>>(
    base?: I,
  ): DialStreamResponse {
    return DialStreamResponse.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<DialStreamResponse>, I>>(
    object: I,
  ): DialStreamResponse {
    const message = createBaseDialStreamResponse()
    message.data =
      object.data !== undefined && object.data !== null
        ? Data.fromPartial(object.data)
        : undefined
    return message
  },
}

/** StreamService is the bifrost stream service. */
export interface StreamService {
  /**
   * ForwardStreams forwards streams to the target multiaddress.
   * Handles HandleMountedStream directives by contacting the target.
   */
  ForwardStreams(
    request: ForwardStreamsRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<ForwardStreamsResponse>
  /**
   * ListenStreams listens for connections to the multiaddress.
   * Forwards the connections to a remote peer with a protocol ID.
   */
  ListenStreams(
    request: ListenStreamsRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<ListenStreamsResponse>
  /**
   * AcceptStream accepts an incoming stream.
   * Stream data is sent over the request / response streams.
   */
  AcceptStream(
    request: AsyncIterable<AcceptStreamRequest>,
    abortSignal?: AbortSignal,
  ): AsyncIterable<AcceptStreamResponse>
  /**
   * DialStream dials a outgoing stream.
   * Stream data is sent over the request / response streams.
   */
  DialStream(
    request: AsyncIterable<DialStreamRequest>,
    abortSignal?: AbortSignal,
  ): AsyncIterable<DialStreamResponse>
}

export const StreamServiceServiceName = 'stream.api.StreamService'
export class StreamServiceClientImpl implements StreamService {
  private readonly rpc: Rpc
  private readonly service: string
  constructor(rpc: Rpc, opts?: { service?: string }) {
    this.service = opts?.service || StreamServiceServiceName
    this.rpc = rpc
    this.ForwardStreams = this.ForwardStreams.bind(this)
    this.ListenStreams = this.ListenStreams.bind(this)
    this.AcceptStream = this.AcceptStream.bind(this)
    this.DialStream = this.DialStream.bind(this)
  }
  ForwardStreams(
    request: ForwardStreamsRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<ForwardStreamsResponse> {
    const data = ForwardStreamsRequest.encode(request).finish()
    const result = this.rpc.serverStreamingRequest(
      this.service,
      'ForwardStreams',
      data,
      abortSignal || undefined,
    )
    return ForwardStreamsResponse.decodeTransform(result)
  }

  ListenStreams(
    request: ListenStreamsRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<ListenStreamsResponse> {
    const data = ListenStreamsRequest.encode(request).finish()
    const result = this.rpc.serverStreamingRequest(
      this.service,
      'ListenStreams',
      data,
      abortSignal || undefined,
    )
    return ListenStreamsResponse.decodeTransform(result)
  }

  AcceptStream(
    request: AsyncIterable<AcceptStreamRequest>,
    abortSignal?: AbortSignal,
  ): AsyncIterable<AcceptStreamResponse> {
    const data = AcceptStreamRequest.encodeTransform(request)
    const result = this.rpc.bidirectionalStreamingRequest(
      this.service,
      'AcceptStream',
      data,
      abortSignal || undefined,
    )
    return AcceptStreamResponse.decodeTransform(result)
  }

  DialStream(
    request: AsyncIterable<DialStreamRequest>,
    abortSignal?: AbortSignal,
  ): AsyncIterable<DialStreamResponse> {
    const data = DialStreamRequest.encodeTransform(request)
    const result = this.rpc.bidirectionalStreamingRequest(
      this.service,
      'DialStream',
      data,
      abortSignal || undefined,
    )
    return DialStreamResponse.decodeTransform(result)
  }
}

/** StreamService is the bifrost stream service. */
export type StreamServiceDefinition = typeof StreamServiceDefinition
export const StreamServiceDefinition = {
  name: 'StreamService',
  fullName: 'stream.api.StreamService',
  methods: {
    /**
     * ForwardStreams forwards streams to the target multiaddress.
     * Handles HandleMountedStream directives by contacting the target.
     */
    forwardStreams: {
      name: 'ForwardStreams',
      requestType: ForwardStreamsRequest,
      requestStream: false,
      responseType: ForwardStreamsResponse,
      responseStream: true,
      options: {},
    },
    /**
     * ListenStreams listens for connections to the multiaddress.
     * Forwards the connections to a remote peer with a protocol ID.
     */
    listenStreams: {
      name: 'ListenStreams',
      requestType: ListenStreamsRequest,
      requestStream: false,
      responseType: ListenStreamsResponse,
      responseStream: true,
      options: {},
    },
    /**
     * AcceptStream accepts an incoming stream.
     * Stream data is sent over the request / response streams.
     */
    acceptStream: {
      name: 'AcceptStream',
      requestType: AcceptStreamRequest,
      requestStream: true,
      responseType: AcceptStreamResponse,
      responseStream: true,
      options: {},
    },
    /**
     * DialStream dials a outgoing stream.
     * Stream data is sent over the request / response streams.
     */
    dialStream: {
      name: 'DialStream',
      requestType: DialStreamRequest,
      requestStream: true,
      responseType: DialStreamResponse,
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
