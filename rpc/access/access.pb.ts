/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { RpcStreamPacket } from '../../../starpc/rpcstream/rpcstream.pb.js'

export const protobufPackage = 'bifrost.rpc.access'

/** LookupRpcServiceRequest is a request to lookup an rpc service. */
export interface LookupRpcServiceRequest {
  /** ServiceId is the service identifier. */
  serviceId: string
  /**
   * ServerId is the identifier of the server requesting the service.
   * Can be empty.
   */
  serverId: string
}

/** LookupRpcServiceResponse is a response to LookupRpcService */
export interface LookupRpcServiceResponse {
  /** Idle indicates the directive is now idle. */
  idle: boolean
  /** Exists indicates we found the service on the remote. */
  exists: boolean
  /** Removed indicates the value no longer exists. */
  removed: boolean
}

function createBaseLookupRpcServiceRequest(): LookupRpcServiceRequest {
  return { serviceId: '', serverId: '' }
}

export const LookupRpcServiceRequest = {
  encode(
    message: LookupRpcServiceRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.serviceId !== '') {
      writer.uint32(10).string(message.serviceId)
    }
    if (message.serverId !== '') {
      writer.uint32(18).string(message.serverId)
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): LookupRpcServiceRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseLookupRpcServiceRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.serviceId = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.serverId = reader.string()
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
  // Transform<LookupRpcServiceRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<LookupRpcServiceRequest | LookupRpcServiceRequest[]>
      | Iterable<LookupRpcServiceRequest | LookupRpcServiceRequest[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [LookupRpcServiceRequest.encode(p).finish()]
        }
      } else {
        yield* [LookupRpcServiceRequest.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, LookupRpcServiceRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<LookupRpcServiceRequest> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [LookupRpcServiceRequest.decode(p)]
        }
      } else {
        yield* [LookupRpcServiceRequest.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): LookupRpcServiceRequest {
    return {
      serviceId: isSet(object.serviceId)
        ? globalThis.String(object.serviceId)
        : '',
      serverId: isSet(object.serverId)
        ? globalThis.String(object.serverId)
        : '',
    }
  },

  toJSON(message: LookupRpcServiceRequest): unknown {
    const obj: any = {}
    if (message.serviceId !== '') {
      obj.serviceId = message.serviceId
    }
    if (message.serverId !== '') {
      obj.serverId = message.serverId
    }
    return obj
  },

  create<I extends Exact<DeepPartial<LookupRpcServiceRequest>, I>>(
    base?: I,
  ): LookupRpcServiceRequest {
    return LookupRpcServiceRequest.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<LookupRpcServiceRequest>, I>>(
    object: I,
  ): LookupRpcServiceRequest {
    const message = createBaseLookupRpcServiceRequest()
    message.serviceId = object.serviceId ?? ''
    message.serverId = object.serverId ?? ''
    return message
  },
}

function createBaseLookupRpcServiceResponse(): LookupRpcServiceResponse {
  return { idle: false, exists: false, removed: false }
}

export const LookupRpcServiceResponse = {
  encode(
    message: LookupRpcServiceResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.idle === true) {
      writer.uint32(8).bool(message.idle)
    }
    if (message.exists === true) {
      writer.uint32(16).bool(message.exists)
    }
    if (message.removed === true) {
      writer.uint32(24).bool(message.removed)
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): LookupRpcServiceResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseLookupRpcServiceResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.idle = reader.bool()
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.exists = reader.bool()
          continue
        case 3:
          if (tag !== 24) {
            break
          }

          message.removed = reader.bool()
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
  // Transform<LookupRpcServiceResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<LookupRpcServiceResponse | LookupRpcServiceResponse[]>
      | Iterable<LookupRpcServiceResponse | LookupRpcServiceResponse[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [LookupRpcServiceResponse.encode(p).finish()]
        }
      } else {
        yield* [LookupRpcServiceResponse.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, LookupRpcServiceResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<LookupRpcServiceResponse> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [LookupRpcServiceResponse.decode(p)]
        }
      } else {
        yield* [LookupRpcServiceResponse.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): LookupRpcServiceResponse {
    return {
      idle: isSet(object.idle) ? globalThis.Boolean(object.idle) : false,
      exists: isSet(object.exists) ? globalThis.Boolean(object.exists) : false,
      removed: isSet(object.removed)
        ? globalThis.Boolean(object.removed)
        : false,
    }
  },

  toJSON(message: LookupRpcServiceResponse): unknown {
    const obj: any = {}
    if (message.idle === true) {
      obj.idle = message.idle
    }
    if (message.exists === true) {
      obj.exists = message.exists
    }
    if (message.removed === true) {
      obj.removed = message.removed
    }
    return obj
  },

  create<I extends Exact<DeepPartial<LookupRpcServiceResponse>, I>>(
    base?: I,
  ): LookupRpcServiceResponse {
    return LookupRpcServiceResponse.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<LookupRpcServiceResponse>, I>>(
    object: I,
  ): LookupRpcServiceResponse {
    const message = createBaseLookupRpcServiceResponse()
    message.idle = object.idle ?? false
    message.exists = object.exists ?? false
    message.removed = object.removed ?? false
    return message
  },
}

/** AccessRpcService exposes services with LookupRpcService via RPC. */
export interface AccessRpcService {
  /**
   * LookupRpcService checks if a RPC service exists with the given info.
   * Usually translates to accessing the LookupRpcService directive.
   * If the service was not found (directive is idle) returns empty.
   */
  LookupRpcService(
    request: LookupRpcServiceRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<LookupRpcServiceResponse>
  /**
   * CallRpcService forwards an RPC call to the service with the component ID.
   * Component ID: json encoded LookupRpcServiceRequest.
   */
  CallRpcService(
    request: AsyncIterable<RpcStreamPacket>,
    abortSignal?: AbortSignal,
  ): AsyncIterable<RpcStreamPacket>
}

export const AccessRpcServiceServiceName = 'bifrost.rpc.access.AccessRpcService'
export class AccessRpcServiceClientImpl implements AccessRpcService {
  private readonly rpc: Rpc
  private readonly service: string
  constructor(rpc: Rpc, opts?: { service?: string }) {
    this.service = opts?.service || AccessRpcServiceServiceName
    this.rpc = rpc
    this.LookupRpcService = this.LookupRpcService.bind(this)
    this.CallRpcService = this.CallRpcService.bind(this)
  }
  LookupRpcService(
    request: LookupRpcServiceRequest,
    abortSignal?: AbortSignal,
  ): AsyncIterable<LookupRpcServiceResponse> {
    const data = LookupRpcServiceRequest.encode(request).finish()
    const result = this.rpc.serverStreamingRequest(
      this.service,
      'LookupRpcService',
      data,
      abortSignal || undefined,
    )
    return LookupRpcServiceResponse.decodeTransform(result)
  }

  CallRpcService(
    request: AsyncIterable<RpcStreamPacket>,
    abortSignal?: AbortSignal,
  ): AsyncIterable<RpcStreamPacket> {
    const data = RpcStreamPacket.encodeTransform(request)
    const result = this.rpc.bidirectionalStreamingRequest(
      this.service,
      'CallRpcService',
      data,
      abortSignal || undefined,
    )
    return RpcStreamPacket.decodeTransform(result)
  }
}

/** AccessRpcService exposes services with LookupRpcService via RPC. */
export type AccessRpcServiceDefinition = typeof AccessRpcServiceDefinition
export const AccessRpcServiceDefinition = {
  name: 'AccessRpcService',
  fullName: 'bifrost.rpc.access.AccessRpcService',
  methods: {
    /**
     * LookupRpcService checks if a RPC service exists with the given info.
     * Usually translates to accessing the LookupRpcService directive.
     * If the service was not found (directive is idle) returns empty.
     */
    lookupRpcService: {
      name: 'LookupRpcService',
      requestType: LookupRpcServiceRequest,
      requestStream: false,
      responseType: LookupRpcServiceResponse,
      responseStream: true,
      options: {},
    },
    /**
     * CallRpcService forwards an RPC call to the service with the component ID.
     * Component ID: json encoded LookupRpcServiceRequest.
     */
    callRpcService: {
      name: 'CallRpcService',
      requestType: RpcStreamPacket,
      requestStream: true,
      responseType: RpcStreamPacket,
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
