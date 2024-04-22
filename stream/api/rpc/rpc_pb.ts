// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/api/rpc/rpc.proto (package stream.api.rpc, syntax proto3)
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
 * StreamState is state for the stream related calls.
 *
 * @generated from enum stream.api.rpc.StreamState
 */
export enum StreamState {
  /**
   * StreamState_NONE indicates nothing about the state
   *
   * @generated from enum value: StreamState_NONE = 0;
   */
  StreamState_NONE = 0,

  /**
   * StreamState_ESTABLISHING indicates the stream is connecting.
   *
   * @generated from enum value: StreamState_ESTABLISHING = 1;
   */
  StreamState_ESTABLISHING = 1,

  /**
   * StreamState_ESTABLISHED indicates the stream is established.
   *
   * @generated from enum value: StreamState_ESTABLISHED = 2;
   */
  StreamState_ESTABLISHED = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(StreamState)
proto3.util.setEnumType(StreamState, 'stream.api.rpc.StreamState', [
  { no: 0, name: 'StreamState_NONE' },
  { no: 1, name: 'StreamState_ESTABLISHING' },
  { no: 2, name: 'StreamState_ESTABLISHED' },
])

/**
 * Data is a data packet.
 *
 * @generated from message stream.api.rpc.Data
 */
export class Data extends Message<Data> {
  /**
   * State indicates stream state in-band.
   * Data is packet data from the remote.
   *
   * @generated from field: bytes data = 1;
   */
  data = new Uint8Array(0)

  /**
   * State indicates the stream state.
   *
   * @generated from field: stream.api.rpc.StreamState state = 2;
   */
  state = StreamState.StreamState_NONE

  constructor(data?: PartialMessage<Data>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.rpc.Data'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'data', kind: 'scalar', T: 12 /* ScalarType.BYTES */ },
    { no: 2, name: 'state', kind: 'enum', T: proto3.getEnumType(StreamState) },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): Data {
    return new Data().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): Data {
    return new Data().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): Data {
    return new Data().fromJsonString(jsonString, options)
  }

  static equals(
    a: Data | PlainMessage<Data> | undefined,
    b: Data | PlainMessage<Data> | undefined,
  ): boolean {
    return proto3.util.equals(Data, a, b)
  }
}
