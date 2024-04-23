// @generated by protoc-gen-es v1.9.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/daemon/api/api.proto (package bifrost.api, syntax proto3)
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
import { Config as Config$1 } from '@go/github.com/aperturerobotics/controllerbus/bus/api/api_pb.js'

/**
 * Config configures the API.
 *
 * @generated from message bifrost.api.Config
 */
export class Config extends Message<Config> {
  /**
   * BusConfig configures the bus api.
   *
   * @generated from field: bus.api.Config bus_config = 1;
   */
  busConfig?: Config$1

  constructor(data?: PartialMessage<Config>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'bifrost.api.Config'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'bus_config', kind: 'message', T: Config$1 },
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
