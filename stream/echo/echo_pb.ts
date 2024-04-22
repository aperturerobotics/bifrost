// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/echo/echo.proto (package stream.echo, syntax proto3)
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

/**
 * Config configures the echo controller.
 *
 * @generated from message stream.echo.Config
 */
export class Config extends Message<Config> {
  /**
   * PeerId is the peer ID to echo for.
   * Can be empty.
   *
   * @generated from field: string peer_id = 1;
   */
  peerId = ''

  /**
   * ProtocolId is the protocol ID to echo on.
   *
   * @generated from field: string protocol_id = 2;
   */
  protocolId = ''

  constructor(data?: PartialMessage<Config>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.echo.Config'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'peer_id', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
    {
      no: 2,
      name: 'protocol_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
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
