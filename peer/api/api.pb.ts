/* eslint-disable */
import {
  ControllerStatus,
  controllerStatusFromJSON,
  controllerStatusToJSON,
} from '@go/github.com/aperturerobotics/controllerbus/controller/exec/exec.pb.js'
import Long from 'long'
import { Config } from '../controller/config.pb.js'
import _m0 from 'protobufjs/minimal.js'

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
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.config !== undefined) {
      Config.encode(message.config, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): IdentifyRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseIdentifyRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.config = Config.decode(reader, reader.uint32())
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<IdentifyRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<IdentifyRequest | IdentifyRequest[]>
      | Iterable<IdentifyRequest | IdentifyRequest[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [IdentifyRequest.encode(p).finish()]
        }
      } else {
        yield* [IdentifyRequest.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, IdentifyRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<IdentifyRequest> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [IdentifyRequest.decode(p)]
        }
      } else {
        yield* [IdentifyRequest.decode(pkt)]
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
    message.config !== undefined &&
      (obj.config = message.config ? Config.toJSON(message.config) : undefined)
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<IdentifyRequest>, I>>(
    object: I
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
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.controllerStatus !== 0) {
      writer.uint32(8).int32(message.controllerStatus)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): IdentifyResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseIdentifyResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.controllerStatus = reader.int32() as any
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<IdentifyResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<IdentifyResponse | IdentifyResponse[]>
      | Iterable<IdentifyResponse | IdentifyResponse[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [IdentifyResponse.encode(p).finish()]
        }
      } else {
        yield* [IdentifyResponse.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, IdentifyResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<IdentifyResponse> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [IdentifyResponse.decode(p)]
        }
      } else {
        yield* [IdentifyResponse.decode(pkt)]
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
    message.controllerStatus !== undefined &&
      (obj.controllerStatus = controllerStatusToJSON(message.controllerStatus))
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<IdentifyResponse>, I>>(
    object: I
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
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.peerId !== '') {
      writer.uint32(10).string(message.peerId)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPeerInfoRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseGetPeerInfoRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.peerId = reader.string()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<GetPeerInfoRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<GetPeerInfoRequest | GetPeerInfoRequest[]>
      | Iterable<GetPeerInfoRequest | GetPeerInfoRequest[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [GetPeerInfoRequest.encode(p).finish()]
        }
      } else {
        yield* [GetPeerInfoRequest.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, GetPeerInfoRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<GetPeerInfoRequest> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [GetPeerInfoRequest.decode(p)]
        }
      } else {
        yield* [GetPeerInfoRequest.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): GetPeerInfoRequest {
    return {
      peerId: isSet(object.peerId) ? String(object.peerId) : '',
    }
  },

  toJSON(message: GetPeerInfoRequest): unknown {
    const obj: any = {}
    message.peerId !== undefined && (obj.peerId = message.peerId)
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<GetPeerInfoRequest>, I>>(
    object: I
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
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.peerId !== '') {
      writer.uint32(10).string(message.peerId)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PeerInfo {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBasePeerInfo()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.peerId = reader.string()
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<PeerInfo, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<PeerInfo | PeerInfo[]>
      | Iterable<PeerInfo | PeerInfo[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [PeerInfo.encode(p).finish()]
        }
      } else {
        yield* [PeerInfo.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, PeerInfo>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<PeerInfo> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [PeerInfo.decode(p)]
        }
      } else {
        yield* [PeerInfo.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): PeerInfo {
    return {
      peerId: isSet(object.peerId) ? String(object.peerId) : '',
    }
  },

  toJSON(message: PeerInfo): unknown {
    const obj: any = {}
    message.peerId !== undefined && (obj.peerId = message.peerId)
    return obj
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
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.localPeers) {
      PeerInfo.encode(v!, writer.uint32(10).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPeerInfoResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseGetPeerInfoResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          message.localPeers.push(PeerInfo.decode(reader, reader.uint32()))
          break
        default:
          reader.skipType(tag & 7)
          break
      }
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<GetPeerInfoResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<GetPeerInfoResponse | GetPeerInfoResponse[]>
      | Iterable<GetPeerInfoResponse | GetPeerInfoResponse[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [GetPeerInfoResponse.encode(p).finish()]
        }
      } else {
        yield* [GetPeerInfoResponse.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, GetPeerInfoResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<GetPeerInfoResponse> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [GetPeerInfoResponse.decode(p)]
        }
      } else {
        yield* [GetPeerInfoResponse.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): GetPeerInfoResponse {
    return {
      localPeers: Array.isArray(object?.localPeers)
        ? object.localPeers.map((e: any) => PeerInfo.fromJSON(e))
        : [],
    }
  },

  toJSON(message: GetPeerInfoResponse): unknown {
    const obj: any = {}
    if (message.localPeers) {
      obj.localPeers = message.localPeers.map((e) =>
        e ? PeerInfo.toJSON(e) : undefined
      )
    } else {
      obj.localPeers = []
    }
    return obj
  },

  fromPartial<I extends Exact<DeepPartial<GetPeerInfoResponse>, I>>(
    object: I
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
  Identify(request: IdentifyRequest): AsyncIterable<IdentifyResponse>
  /** GetPeerInfo returns information about attached peers. */
  GetPeerInfo(request: GetPeerInfoRequest): Promise<GetPeerInfoResponse>
}

export class PeerServiceClientImpl implements PeerService {
  private readonly rpc: Rpc
  constructor(rpc: Rpc) {
    this.rpc = rpc
    this.Identify = this.Identify.bind(this)
    this.GetPeerInfo = this.GetPeerInfo.bind(this)
  }
  Identify(request: IdentifyRequest): AsyncIterable<IdentifyResponse> {
    const data = IdentifyRequest.encode(request).finish()
    const result = this.rpc.serverStreamingRequest(
      'peer.api.PeerService',
      'Identify',
      data
    )
    return IdentifyResponse.decodeTransform(result)
  }

  GetPeerInfo(request: GetPeerInfoRequest): Promise<GetPeerInfoResponse> {
    const data = GetPeerInfoRequest.encode(request).finish()
    const promise = this.rpc.request(
      'peer.api.PeerService',
      'GetPeerInfo',
      data
    )
    return promise.then((data) =>
      GetPeerInfoResponse.decode(new _m0.Reader(data))
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
    data: Uint8Array
  ): Promise<Uint8Array>
  clientStreamingRequest(
    service: string,
    method: string,
    data: AsyncIterable<Uint8Array>
  ): Promise<Uint8Array>
  serverStreamingRequest(
    service: string,
    method: string,
    data: Uint8Array
  ): AsyncIterable<Uint8Array>
  bidirectionalStreamingRequest(
    service: string,
    method: string,
    data: AsyncIterable<Uint8Array>
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
