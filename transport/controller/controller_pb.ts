// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/controller/controller.proto (package transport.controller, syntax proto3)
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
 * StreamEstablish is the first message sent by the initiator of a stream.
 * Prefixed by a uint32 length.
 * Max size: 100kb
 *
 * @generated from message transport.controller.StreamEstablish
 */
export class StreamEstablish extends Message<StreamEstablish> {
  /**
   * ProtocolID is the protocol identifier string for the stream.
   *
   * @generated from field: string protocol_id = 1;
   */
  protocolId = ''

  constructor(data?: PartialMessage<StreamEstablish>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'transport.controller.StreamEstablish'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'protocol_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): StreamEstablish {
    return new StreamEstablish().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): StreamEstablish {
    return new StreamEstablish().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): StreamEstablish {
    return new StreamEstablish().fromJsonString(jsonString, options)
  }

  static equals(
    a: StreamEstablish | PlainMessage<StreamEstablish> | undefined,
    b: StreamEstablish | PlainMessage<StreamEstablish> | undefined,
  ): boolean {
    return proto3.util.equals(StreamEstablish, a, b)
  }
}
