// @generated by protoc-gen-es v1.9.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/peer/api/api.proto (package peer.api, syntax proto3)
/* eslint-disable */

import type {
  BinaryReadOptions,
  FieldList,
  JsonReadOptions,
  JsonValue,
  PartialMessage,
  PlainMessage,
} from '@bufbuild/protobuf'
import { Message, proto3 } from '@bufbuild/protobuf'
import { Config } from '../controller/config_pb.js'
import { ControllerStatus } from '../../../controllerbus/controller/exec/exec_pb.js'

/**
 * IdentifyRequest is a request to load an identity.
 *
 * @generated from message peer.api.IdentifyRequest
 */
export class IdentifyRequest extends Message<IdentifyRequest> {
  /**
   * Config is the request to configure the peer controller.
   *
   * @generated from field: peer.controller.Config config = 1;
   */
  config?: Config

  constructor(data?: PartialMessage<IdentifyRequest>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'peer.api.IdentifyRequest'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'config', kind: 'message', T: Config },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): IdentifyRequest {
    return new IdentifyRequest().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): IdentifyRequest {
    return new IdentifyRequest().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): IdentifyRequest {
    return new IdentifyRequest().fromJsonString(jsonString, options)
  }

  static equals(
    a: IdentifyRequest | PlainMessage<IdentifyRequest> | undefined,
    b: IdentifyRequest | PlainMessage<IdentifyRequest> | undefined,
  ): boolean {
    return proto3.util.equals(IdentifyRequest, a, b)
  }
}

/**
 * IdentifyResponse is a response to an identify request.
 *
 * @generated from message peer.api.IdentifyResponse
 */
export class IdentifyResponse extends Message<IdentifyResponse> {
  /**
   * ControllerStatus is the status of the peer controller.
   *
   * @generated from field: controller.exec.ControllerStatus controller_status = 1;
   */
  controllerStatus = ControllerStatus.ControllerStatus_UNKNOWN

  constructor(data?: PartialMessage<IdentifyResponse>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'peer.api.IdentifyResponse'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'controller_status',
      kind: 'enum',
      T: proto3.getEnumType(ControllerStatus),
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): IdentifyResponse {
    return new IdentifyResponse().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): IdentifyResponse {
    return new IdentifyResponse().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): IdentifyResponse {
    return new IdentifyResponse().fromJsonString(jsonString, options)
  }

  static equals(
    a: IdentifyResponse | PlainMessage<IdentifyResponse> | undefined,
    b: IdentifyResponse | PlainMessage<IdentifyResponse> | undefined,
  ): boolean {
    return proto3.util.equals(IdentifyResponse, a, b)
  }
}

/**
 * GetPeerInfoRequest is the request type for GetPeerInfo.
 *
 * @generated from message peer.api.GetPeerInfoRequest
 */
export class GetPeerInfoRequest extends Message<GetPeerInfoRequest> {
  /**
   * PeerId restricts the response to a specific peer ID.
   *
   * @generated from field: string peer_id = 1;
   */
  peerId = ''

  constructor(data?: PartialMessage<GetPeerInfoRequest>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'peer.api.GetPeerInfoRequest'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'peer_id', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): GetPeerInfoRequest {
    return new GetPeerInfoRequest().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): GetPeerInfoRequest {
    return new GetPeerInfoRequest().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): GetPeerInfoRequest {
    return new GetPeerInfoRequest().fromJsonString(jsonString, options)
  }

  static equals(
    a: GetPeerInfoRequest | PlainMessage<GetPeerInfoRequest> | undefined,
    b: GetPeerInfoRequest | PlainMessage<GetPeerInfoRequest> | undefined,
  ): boolean {
    return proto3.util.equals(GetPeerInfoRequest, a, b)
  }
}

/**
 * PeerInfo is basic information about a peer.
 *
 * @generated from message peer.api.PeerInfo
 */
export class PeerInfo extends Message<PeerInfo> {
  /**
   * PeerId is the b58 peer ID.
   *
   * @generated from field: string peer_id = 1;
   */
  peerId = ''

  constructor(data?: PartialMessage<PeerInfo>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'peer.api.PeerInfo'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'peer_id', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): PeerInfo {
    return new PeerInfo().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): PeerInfo {
    return new PeerInfo().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): PeerInfo {
    return new PeerInfo().fromJsonString(jsonString, options)
  }

  static equals(
    a: PeerInfo | PlainMessage<PeerInfo> | undefined,
    b: PeerInfo | PlainMessage<PeerInfo> | undefined,
  ): boolean {
    return proto3.util.equals(PeerInfo, a, b)
  }
}

/**
 * GetPeerInfoResponse is the response type for GetPeerInfo.
 *
 * @generated from message peer.api.GetPeerInfoResponse
 */
export class GetPeerInfoResponse extends Message<GetPeerInfoResponse> {
  /**
   * LocalPeers is the set of peers loaded.
   *
   * @generated from field: repeated peer.api.PeerInfo local_peers = 1;
   */
  localPeers: PeerInfo[] = []

  constructor(data?: PartialMessage<GetPeerInfoResponse>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'peer.api.GetPeerInfoResponse'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'local_peers',
      kind: 'message',
      T: PeerInfo,
      repeated: true,
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): GetPeerInfoResponse {
    return new GetPeerInfoResponse().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): GetPeerInfoResponse {
    return new GetPeerInfoResponse().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): GetPeerInfoResponse {
    return new GetPeerInfoResponse().fromJsonString(jsonString, options)
  }

  static equals(
    a: GetPeerInfoResponse | PlainMessage<GetPeerInfoResponse> | undefined,
    b: GetPeerInfoResponse | PlainMessage<GetPeerInfoResponse> | undefined,
  ): boolean {
    return proto3.util.equals(GetPeerInfoResponse, a, b)
  }
}
