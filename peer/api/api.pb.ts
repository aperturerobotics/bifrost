/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import {
  ControllerStatus,
  controllerStatusFromJSON,
  controllerStatusToJSON,
} from '../../../controllerbus/controller/exec/exec.pb.js'
import { Config } from '../controller/config.pb.js'

export const protobufPackage = 'peer.api'

/** IdentifyRequest is a request to load an identity. */
export interface IdentifyRequest {
  /** Config is the request to configure the peer controller. */
  config: Config | undefined
}

/** IdentifyResponse is a response to an identify request. */
export interface IdentifyResponse {
  /** ControllerStatus is the status of the peer controller. */
  controllerStatus: ControllerStatus
}

/** GetPeerInfoRequest is the request type for GetPeerInfo. */
export interface GetPeerInfoRequest {
  /** PeerId restricts the response to a specific peer ID. */
  peerId: string
}

/** PeerInfo is basic information about a peer. */
export interface PeerInfo {
  /** PeerId is the b58 peer ID. */
  peerId: string
}

/** GetPeerInfoResponse is the response type for GetPeerInfo. */
export interface GetPeerInfoResponse {
  /** LocalPeers is the set of peers loaded. */
  localPeers: PeerInfo[]
}

function createBaseIdentifyRequest(): IdentifyRequest {
  return { config: undefined }
}

export const IdentifyRequest = {
  encode(
    message: IdentifyRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.config !== undefined) {
      Config.encode(message.config, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): IdentifyRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseIdentifyRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.config = Config.decode(reader, reader.uint32())
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
  // Transform<IdentifyRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<IdentifyRequest | IdentifyRequest[]>
      | Iterable<IdentifyRequest | IdentifyRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [IdentifyRequest.encode(p).finish()]
        }
      } else {
        yield* [IdentifyRequest.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, IdentifyRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<IdentifyRequest> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [IdentifyRequest.decode(p)]
        }
      } else {
        yield* [IdentifyRequest.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): IdentifyRequest {
    return {
      config: isSet(object.config) ? Config.fromJSON(object.config) : undefined,
    }
  },

  toJSON(message: IdentifyRequest): unknown {
    const obj: any = {}
    if (message.config !== undefined) {
      obj.config = Config.toJSON(message.config)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<IdentifyRequest>, I>>(
    base?: I,
  ): IdentifyRequest {
    return IdentifyRequest.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<IdentifyRequest>, I>>(
    object: I,
  ): IdentifyRequest {
    const message = createBaseIdentifyRequest()
    message.config =
      object.config !== undefined && object.config !== null
        ? Config.fromPartial(object.config)
        : undefined
    return message
  },
}

function createBaseIdentifyResponse(): IdentifyResponse {
  return { controllerStatus: 0 }
}

export const IdentifyResponse = {
  encode(
    message: IdentifyResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.controllerStatus !== 0) {
      writer.uint32(8).int32(message.controllerStatus)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): IdentifyResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseIdentifyResponse()
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
  // Transform<IdentifyResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<IdentifyResponse | IdentifyResponse[]>
      | Iterable<IdentifyResponse | IdentifyResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [IdentifyResponse.encode(p).finish()]
        }
      } else {
        yield* [IdentifyResponse.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, IdentifyResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<IdentifyResponse> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [IdentifyResponse.decode(p)]
        }
      } else {
        yield* [IdentifyResponse.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): IdentifyResponse {
    return {
      controllerStatus: isSet(object.controllerStatus)
        ? controllerStatusFromJSON(object.controllerStatus)
        : 0,
    }
  },

  toJSON(message: IdentifyResponse): unknown {
    const obj: any = {}
    if (message.controllerStatus !== 0) {
      obj.controllerStatus = controllerStatusToJSON(message.controllerStatus)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<IdentifyResponse>, I>>(
    base?: I,
  ): IdentifyResponse {
    return IdentifyResponse.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<IdentifyResponse>, I>>(
    object: I,
  ): IdentifyResponse {
    const message = createBaseIdentifyResponse()
    message.controllerStatus = object.controllerStatus ?? 0
    return message
  },
}

function createBaseGetPeerInfoRequest(): GetPeerInfoRequest {
  return { peerId: '' }
}

export const GetPeerInfoRequest = {
  encode(
    message: GetPeerInfoRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.peerId !== '') {
      writer.uint32(10).string(message.peerId)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPeerInfoRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseGetPeerInfoRequest()
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
  // Transform<GetPeerInfoRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<GetPeerInfoRequest | GetPeerInfoRequest[]>
      | Iterable<GetPeerInfoRequest | GetPeerInfoRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [GetPeerInfoRequest.encode(p).finish()]
        }
      } else {
        yield* [GetPeerInfoRequest.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, GetPeerInfoRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<GetPeerInfoRequest> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [GetPeerInfoRequest.decode(p)]
        }
      } else {
        yield* [GetPeerInfoRequest.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): GetPeerInfoRequest {
    return {
      peerId: isSet(object.peerId) ? globalThis.String(object.peerId) : '',
    }
  },

  toJSON(message: GetPeerInfoRequest): unknown {
    const obj: any = {}
    if (message.peerId !== '') {
      obj.peerId = message.peerId
    }
    return obj
  },

  create<I extends Exact<DeepPartial<GetPeerInfoRequest>, I>>(
    base?: I,
  ): GetPeerInfoRequest {
    return GetPeerInfoRequest.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<GetPeerInfoRequest>, I>>(
    object: I,
  ): GetPeerInfoRequest {
    const message = createBaseGetPeerInfoRequest()
    message.peerId = object.peerId ?? ''
    return message
  },
}

function createBasePeerInfo(): PeerInfo {
  return { peerId: '' }
}

export const PeerInfo = {
  encode(
    message: PeerInfo,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.peerId !== '') {
      writer.uint32(10).string(message.peerId)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PeerInfo {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePeerInfo()
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
  // Transform<PeerInfo, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<PeerInfo | PeerInfo[]>
      | Iterable<PeerInfo | PeerInfo[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [PeerInfo.encode(p).finish()]
        }
      } else {
        yield* [PeerInfo.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, PeerInfo>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<PeerInfo> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [PeerInfo.decode(p)]
        }
      } else {
        yield* [PeerInfo.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): PeerInfo {
    return {
      peerId: isSet(object.peerId) ? globalThis.String(object.peerId) : '',
    }
  },

  toJSON(message: PeerInfo): unknown {
    const obj: any = {}
    if (message.peerId !== '') {
      obj.peerId = message.peerId
    }
    return obj
  },

  create<I extends Exact<DeepPartial<PeerInfo>, I>>(base?: I): PeerInfo {
    return PeerInfo.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<PeerInfo>, I>>(object: I): PeerInfo {
    const message = createBasePeerInfo()
    message.peerId = object.peerId ?? ''
    return message
  },
}

function createBaseGetPeerInfoResponse(): GetPeerInfoResponse {
  return { localPeers: [] }
}

export const GetPeerInfoResponse = {
  encode(
    message: GetPeerInfoResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.localPeers) {
      PeerInfo.encode(v!, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPeerInfoResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseGetPeerInfoResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.localPeers.push(PeerInfo.decode(reader, reader.uint32()))
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
  // Transform<GetPeerInfoResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<GetPeerInfoResponse | GetPeerInfoResponse[]>
      | Iterable<GetPeerInfoResponse | GetPeerInfoResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [GetPeerInfoResponse.encode(p).finish()]
        }
      } else {
        yield* [GetPeerInfoResponse.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, GetPeerInfoResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<GetPeerInfoResponse> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [GetPeerInfoResponse.decode(p)]
        }
      } else {
        yield* [GetPeerInfoResponse.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): GetPeerInfoResponse {
    return {
      localPeers: globalThis.Array.isArray(object?.localPeers)
        ? object.localPeers.map((e: any) => PeerInfo.fromJSON(e))
        : [],
    }
  },

  toJSON(message: GetPeerInfoResponse): unknown {
    const obj: any = {}
    if (message.localPeers?.length) {
      obj.localPeers = message.localPeers.map((e) => PeerInfo.toJSON(e))
    }
    return obj
  },

  create<I extends Exact<DeepPartial<GetPeerInfoResponse>, I>>(
    base?: I,
  ): GetPeerInfoResponse {
    return GetPeerInfoResponse.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<GetPeerInfoResponse>, I>>(
    object: I,
  ): GetPeerInfoResponse {
    const message = createBaseGetPeerInfoResponse()
    message.localPeers =
      object.localPeers?.map((e) => PeerInfo.fromPartial(e)) || []
    return message
  },
}

/** PeerService implements a bifrost peer service. */
export interface PeerService {
  /** Identify loads and manages a private key identity. */
  Identify(
    request: IdentifyRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<IdentifyResponse>
  /** GetPeerInfo returns information about attached peers. */
  GetPeerInfo(
    request: GetPeerInfoRequest,
    abortSignal?: AbortSignal,
  ): Promise<GetPeerInfoResponse>
}

export const PeerServiceServiceName = 'peer.api.PeerService'
export class PeerServiceClientImpl implements PeerService {
  private readonly rpc: Rpc
  private readonly service: string
  constructor(rpc: Rpc, opts?: { service?: string }) {
    this.service = opts?.service || PeerServiceServiceName
    this.rpc = rpc
    this.Identify = this.Identify.bind(this)
    this.GetPeerInfo = this.GetPeerInfo.bind(this)
  }
  Identify(
    request: IdentifyRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<IdentifyResponse> {
    const data = IdentifyRequest.encode(request).finish()
    const result = this.rpc.serverStreamingRequest(
      this.service,
      'Identify',
      data,
      abortSignal || undefined,
    )
    return IdentifyResponse.decodeTransform(result)
  }

  GetPeerInfo(
    request: GetPeerInfoRequest,
    abortSignal?: AbortSignal,
  ): Promise<GetPeerInfoResponse> {
    const data = GetPeerInfoRequest.encode(request).finish()
    const promise = this.rpc.request(
      this.service,
      'GetPeerInfo',
      data,
      abortSignal || undefined,
    )
    return promise.then((data) =>
      GetPeerInfoResponse.decode(_m0.Reader.create(data)),
    )
  }
}

/** PeerService implements a bifrost peer service. */
export type PeerServiceDefinition = typeof PeerServiceDefinition
export const PeerServiceDefinition = {
  name: 'PeerService',
  fullName: 'peer.api.PeerService',
  methods: {
    /** Identify loads and manages a private key identity. */
    identify: {
      name: 'Identify',
      requestType: IdentifyRequest,
      requestStream: false,
      responseType: IdentifyResponse,
      responseStream: true,
      options: {},
    },
    /** GetPeerInfo returns information about attached peers. */
    getPeerInfo: {
      name: 'GetPeerInfo',
      requestType: GetPeerInfoRequest,
      requestStream: false,
      responseType: GetPeerInfoResponse,
      responseStream: false,
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
