// @generated by protoc-gen-es v1.9.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/peer/controller/config.proto (package peer.controller, syntax proto3)
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
 * Config is the peer controller config.
 *
 * @generated from message peer.controller.Config
 */
export class Config extends Message<Config> {
  /**
   * PrivKey is the peer private key in either b58 or PEM format.
   * See confparse.MarshalPrivateKey.
   * If not set, the peer private key will be unavailable.
   *
   * @generated from field: string priv_key = 1;
   */
  privKey = ''

  /**
   * PubKey is the peer public key.
   * Ignored if priv_key is set.
   *
   * @generated from field: string pub_key = 2;
   */
  pubKey = ''

  /**
   * PeerId is the peer identifier.
   * Ignored if priv_key or pub_key are set.
   * The peer ID should contain the public key.
   *
   * @generated from field: string peer_id = 3;
   */
  peerId = ''

  constructor(data?: PartialMessage<Config>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'peer.controller.Config'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'priv_key', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
    { no: 2, name: 'pub_key', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
    { no: 3, name: 'peer_id', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
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