// @generated by protoc-gen-es v1.9.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/common/quic/quic.proto (package transport.quic, syntax proto3)
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
 * @generated from message transport.quic.Opts
 */
export class Opts extends Message<Opts> {
  /**
   * MaxIdleTimeoutDur is the duration of idle after which conn is closed.
   *
   * Defaults to 10s.
   *
   * @generated from field: string max_idle_timeout_dur = 1;
   */
  maxIdleTimeoutDur = ''

  /**
   * MaxIncomingStreams is the maximum number of concurrent bidirectional
   * streams that a peer is allowed to open.
   *
   * If unset or negative, defaults to 100000.
   *
   * @generated from field: int32 max_incoming_streams = 2;
   */
  maxIncomingStreams = 0

  /**
   * DisableKeepAlive disables the keep alive packets.
   *
   *
   * @generated from field: bool disable_keep_alive = 3;
   */
  disableKeepAlive = false

  /**
   * KeepAliveDur is the duration between keep-alive pings.
   *
   * If disable_keep_alive is set, this value is ignored.
   * If unset, sets keep-alive to half of MaxIdleTimeout.
   *
   * @generated from field: string keep_alive_dur = 7;
   */
  keepAliveDur = ''

  /**
   * DisableDatagrams disables the unreliable datagrams feature.
   * Both peers must support it for it to be enabled, regardless of this flag.
   *
   * @generated from field: bool disable_datagrams = 4;
   */
  disableDatagrams = false

  /**
   * DisablePathMtuDiscovery disables sending packets to discover max packet size.
   *
   * @generated from field: bool disable_path_mtu_discovery = 5;
   */
  disablePathMtuDiscovery = false

  /**
   * Verbose indicates to use verbose logging.
   * Note: this is VERY verbose, logs every packet sent.
   *
   * @generated from field: bool verbose = 6;
   */
  verbose = false

  constructor(data?: PartialMessage<Opts>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'transport.quic.Opts'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'max_idle_timeout_dur',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 2,
      name: 'max_incoming_streams',
      kind: 'scalar',
      T: 5 /* ScalarType.INT32 */,
    },
    {
      no: 3,
      name: 'disable_keep_alive',
      kind: 'scalar',
      T: 8 /* ScalarType.BOOL */,
    },
    {
      no: 7,
      name: 'keep_alive_dur',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 4,
      name: 'disable_datagrams',
      kind: 'scalar',
      T: 8 /* ScalarType.BOOL */,
    },
    {
      no: 5,
      name: 'disable_path_mtu_discovery',
      kind: 'scalar',
      T: 8 /* ScalarType.BOOL */,
    },
    { no: 6, name: 'verbose', kind: 'scalar', T: 8 /* ScalarType.BOOL */ },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): Opts {
    return new Opts().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): Opts {
    return new Opts().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): Opts {
    return new Opts().fromJsonString(jsonString, options)
  }

  static equals(
    a: Opts | PlainMessage<Opts> | undefined,
    b: Opts | PlainMessage<Opts> | undefined,
  ): boolean {
    return proto3.util.equals(Opts, a, b)
  }
}