// @generated by protoc-gen-es v1.9.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/srpc/client/controller/config.proto (package stream.srpc.client.controller, syntax proto3)
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
import { Config as Config$1 } from '../client_pb.js'

/**
 * Config configures mounting a bifrost srpc RPC client to a bus.
 * Resolves the LookupRpcClient directive.
 *
 * @generated from message stream.srpc.client.controller.Config
 */
export class Config extends Message<Config> {
  /**
   * Client contains srpc.client configuration for the RPC client.
   *
   * @generated from field: stream.srpc.client.Config client = 1;
   */
  client?: Config$1

  /**
   * ProtocolId is the protocol ID to use to contact the remote RPC service.
   * Must be set.
   *
   * @generated from field: string protocol_id = 2;
   */
  protocolId = ''

  /**
   * ServiceIdPrefixes are the service ID prefixes to match.
   * The prefix will be stripped from the service id before being passed to the client.
   * This is used like: LookupRpcClient<remote/my/service> -> my/service.
   *
   * If empty slice or empty string: matches all LookupRpcClient calls ignoring service ID.
   * Optional.
   *
   * @generated from field: repeated string service_id_prefixes = 3;
   */
  serviceIdPrefixes: string[] = []

  constructor(data?: PartialMessage<Config>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.srpc.client.controller.Config'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'client', kind: 'message', T: Config$1 },
    {
      no: 2,
      name: 'protocol_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 3,
      name: 'service_id_prefixes',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
      repeated: true,
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