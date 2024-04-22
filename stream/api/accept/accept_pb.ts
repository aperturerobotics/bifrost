// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/api/accept/accept.proto (package stream.api.accept, syntax proto3)
/* eslint-disable */

import type {
  BinaryReadOptions,
  FieldList,
  JsonReadOptions,
  JsonValue,
  PartialMessage,
  PlainMessage,
} from '@bufbuild/protobuf'
import { Message, proto3, protoInt64 } from '@bufbuild/protobuf'

/**
 * Config configures the accept controller.
 *
 * @generated from message stream.api.accept.Config
 */
export class Config extends Message<Config> {
  /**
   * LocalPeerId is the peer ID to accept incoming connections with.
   * Can be empty to accept any peer.
   *
   * @generated from field: string local_peer_id = 1;
   */
  localPeerId = ''

  /**
   * RemotePeerIds are peer IDs to accept incoming connections from.
   * Can be empty to accept any remote peer IDs.
   *
   * @generated from field: repeated string remote_peer_ids = 2;
   */
  remotePeerIds: string[] = []

  /**
   * ProtocolId is the protocol ID to accept.
   *
   * @generated from field: string protocol_id = 3;
   */
  protocolId = ''

  /**
   * TransportId constrains the transport ID to accept from.
   * Can be empty.
   *
   * @generated from field: uint64 transport_id = 4;
   */
  transportId = protoInt64.zero

  constructor(data?: PartialMessage<Config>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.accept.Config'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'local_peer_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 2,
      name: 'remote_peer_ids',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
      repeated: true,
    },
    {
      no: 3,
      name: 'protocol_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 4,
      name: 'transport_id',
      kind: 'scalar',
      T: 4 /* ScalarType.UINT64 */,
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): Config {
    return new Config().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): Config {
    return new Config().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): Config {
    return new Config().fromJsonString(jsonString, options)
  }

  static equals(
    a: Config | PlainMessage<Config> | undefined,
    b: Config | PlainMessage<Config> | undefined,
  ): boolean {
    return proto3.util.equals(Config, a, b)
  }
}
