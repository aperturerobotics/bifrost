// @generated by protoc-gen-es-starpc none with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/rpc/access/access.proto (package bifrost.rpc.access, syntax proto3)
/* eslint-disable */

import {
  LookupRpcServiceRequest,
  LookupRpcServiceResponse,
} from './access_pb.js'
import { MethodKind } from '@bufbuild/protobuf'
import { RpcStreamPacket } from '@go/github.com/aperturerobotics/starpc/rpcstream/rpcstream_pb.js'
import { Message } from '@aptre/protobuf-es-lite'
import {
  buildDecodeMessageTransform,
  buildEncodeMessageTransform,
  MessageStream,
  ProtoRpc,
} from 'starpc'

/**
 * AccessRpcService exposes services with LookupRpcService via RPC.
 *
 * @generated from service bifrost.rpc.access.AccessRpcService
 */
export const AccessRpcServiceDefinition = {
  typeName: 'bifrost.rpc.access.AccessRpcService',
  methods: {
    /**
     * LookupRpcService checks if a RPC service exists with the given info.
     * Usually translates to accessing the LookupRpcService directive.
     * If the service was not found (directive is idle) returns empty.
     *
     * @generated from rpc bifrost.rpc.access.AccessRpcService.LookupRpcService
     */
    LookupRpcService: {
      name: 'LookupRpcService',
      I: LookupRpcServiceRequest,
      O: LookupRpcServiceResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * CallRpcService forwards an RPC call to the service with the component ID.
     * Component ID: json encoded LookupRpcServiceRequest.
     *
     * @generated from rpc bifrost.rpc.access.AccessRpcService.CallRpcService
     */
    CallRpcService: {
      name: 'CallRpcService',
      I: RpcStreamPacket,
      O: RpcStreamPacket,
      kind: MethodKind.BiDiStreaming,
    },
  },
} as const

/**
 * AccessRpcService exposes services with LookupRpcService via RPC.
 *
 * @generated from service bifrost.rpc.access.AccessRpcService
 */
export interface AccessRpcService {
  /**
   * LookupRpcService checks if a RPC service exists with the given info.
   * Usually translates to accessing the LookupRpcService directive.
   * If the service was not found (directive is idle) returns empty.
   *
   * @generated from rpc bifrost.rpc.access.AccessRpcService.LookupRpcService
   */
  LookupRpcService(
    request: Message<LookupRpcServiceRequest>,
    abortSignal?: AbortSignal,
  ): MessageStream<LookupRpcServiceResponse>

  /**
   * CallRpcService forwards an RPC call to the service with the component ID.
   * Component ID: json encoded LookupRpcServiceRequest.
   *
   * @generated from rpc bifrost.rpc.access.AccessRpcService.CallRpcService
   */
  CallRpcService(
    request: MessageStream<RpcStreamPacket>,
    abortSignal?: AbortSignal,
  ): MessageStream<RpcStreamPacket>
}

export const AccessRpcServiceServiceName = AccessRpcServiceDefinition.typeName

export class AccessRpcServiceClient implements AccessRpcService {
  private readonly rpc: ProtoRpc
  private readonly service: string
  constructor(rpc: ProtoRpc, opts?: { service?: string }) {
    this.service = opts?.service || AccessRpcServiceServiceName
    this.rpc = rpc
    this.LookupRpcService = this.LookupRpcService.bind(this)
    this.CallRpcService = this.CallRpcService.bind(this)
  }
  /**
   * LookupRpcService checks if a RPC service exists with the given info.
   * Usually translates to accessing the LookupRpcService directive.
   * If the service was not found (directive is idle) returns empty.
   *
   * @generated from rpc bifrost.rpc.access.AccessRpcService.LookupRpcService
   */
  LookupRpcService(
    request: Message<LookupRpcServiceRequest>,
    abortSignal?: AbortSignal,
  ): MessageStream<LookupRpcServiceResponse> {
    const requestMsg = LookupRpcServiceRequest.create(request)
    const result = this.rpc.serverStreamingRequest(
      this.service,
      AccessRpcServiceDefinition.methods.LookupRpcService.name,
      LookupRpcServiceRequest.toBinary(requestMsg),
      abortSignal || undefined,
    )
    return buildDecodeMessageTransform(LookupRpcServiceResponse)(result)
  }

  /**
   * CallRpcService forwards an RPC call to the service with the component ID.
   * Component ID: json encoded LookupRpcServiceRequest.
   *
   * @generated from rpc bifrost.rpc.access.AccessRpcService.CallRpcService
   */
  CallRpcService(
    request: MessageStream<RpcStreamPacket>,
    abortSignal?: AbortSignal,
  ): MessageStream<RpcStreamPacket> {
    const result = this.rpc.bidirectionalStreamingRequest(
      this.service,
      AccessRpcServiceDefinition.methods.CallRpcService.name,
      buildEncodeMessageTransform(RpcStreamPacket)(request),
      abortSignal || undefined,
    )
    return buildDecodeMessageTransform(RpcStreamPacket)(result)
  }
}
