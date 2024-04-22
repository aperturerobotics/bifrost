// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/common/dialer/dialer.proto (package dialer, syntax proto3)
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
import { Backoff } from '../../../../util/backoff/backoff_pb.js'

/**
 * DialerOpts contains options relating to dialing a statically configured peer.
 *
 * @generated from message dialer.DialerOpts
 */
export class DialerOpts extends Message<DialerOpts> {
  /**
   * Address is the address of the peer, in the format expected by the transport.
   *
   * @generated from field: string address = 1;
   */
  address = ''

  /**
   * Backoff is the dialing backoff configuration.
   * Can be empty.
   *
   * @generated from field: backoff.Backoff backoff = 2;
   */
  backoff?: Backoff

  constructor(data?: PartialMessage<DialerOpts>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'dialer.DialerOpts'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'address', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
    { no: 2, name: 'backoff', kind: 'message', T: Backoff },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): DialerOpts {
    return new DialerOpts().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): DialerOpts {
    return new DialerOpts().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): DialerOpts {
    return new DialerOpts().fromJsonString(jsonString, options)
  }

  static equals(
    a: DialerOpts | PlainMessage<DialerOpts> | undefined,
    b: DialerOpts | PlainMessage<DialerOpts> | undefined,
  ): boolean {
    return proto3.util.equals(DialerOpts, a, b)
  }
}
