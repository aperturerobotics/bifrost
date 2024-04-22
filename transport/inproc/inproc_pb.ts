// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/inproc/inproc.proto (package inproc, syntax proto3)
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
import { Opts } from '../common/pconn/pconn_pb.js'
import { DialerOpts } from '../common/dialer/dialer_pb.js'

/**
 * Config is the configuration for the inproc testing transport.
 *
 * @generated from message inproc.Config
 */
export class Config extends Message<Config> {
  /**
   * TransportPeerID sets the peer ID to attach the transport to.
   * If unset, attaches to any running peer with a private key.
   *
   * @generated from field: string transport_peer_id = 1;
   */
  transportPeerId = ''

  /**
   * PacketOpts are options to set on the packet connection.
   *
   * @generated from field: pconn.Opts packet_opts = 2;
   */
  packetOpts?: Opts

  /**
   * Dialers maps peer IDs to dialers.
   *
   * @generated from field: map<string, dialer.DialerOpts> dialers = 3;
   */
  dialers: { [key: string]: DialerOpts } = {}

  constructor(data?: PartialMessage<Config>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'inproc.Config'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'transport_peer_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    { no: 2, name: 'packet_opts', kind: 'message', T: Opts },
    {
      no: 3,
      name: 'dialers',
      kind: 'map',
      K: 9 /* ScalarType.STRING */,
      V: { kind: 'message', T: DialerOpts },
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
