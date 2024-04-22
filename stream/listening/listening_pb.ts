// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/listening/listening.proto (package stream.listening, syntax proto3)
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
 * Config configures the listening controller.
 *
 * @generated from message stream.listening.Config
 */
export class Config extends Message<Config> {
  /**
   * LocalPeerId is the peer ID to forward incoming connections with.
   * Can be empty.
   *
   * @generated from field: string local_peer_id = 1;
   */
  localPeerId = ''

  /**
   * RemotePeerId is the peer ID to forward incoming connections to.
   *
   * @generated from field: string remote_peer_id = 2;
   */
  remotePeerId = ''

  /**
   * ProtocolId is the protocol ID to assign to incoming connections.
   *
   * @generated from field: string protocol_id = 3;
   */
  protocolId = ''

  /**
   * ListenMultiaddr is the listening multiaddress.
   *
   * @generated from field: string listen_multiaddr = 4;
   */
  listenMultiaddr = ''

  /**
   * TransportId sets a transport ID constraint.
   * Can be empty.
   *
   * @generated from field: uint64 transport_id = 5;
   */
  transportId = protoInt64.zero

  /**
   * Reliable indicates the stream should be reliable.
   *
   * @generated from field: bool reliable = 6;
   */
  reliable = false

  /**
   * Encrypted indicates the stream should be encrypted.
   *
   * @generated from field: bool encrypted = 7;
   */
  encrypted = false

  constructor(data?: PartialMessage<Config>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.listening.Config'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'local_peer_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 2,
      name: 'remote_peer_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 3,
      name: 'protocol_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 4,
      name: 'listen_multiaddr',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 5,
      name: 'transport_id',
      kind: 'scalar',
      T: 4 /* ScalarType.UINT64 */,
    },
    { no: 6, name: 'reliable', kind: 'scalar', T: 8 /* ScalarType.BOOL */ },
    { no: 7, name: 'encrypted', kind: 'scalar', T: 8 /* ScalarType.BOOL */ },
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
