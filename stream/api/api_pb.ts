// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/api/api.proto (package stream.api, syntax proto3)
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
import { Config } from '../forwarding/forwarding_pb.js'
import { ControllerStatus } from '../../../controllerbus/controller/exec/exec_pb.js'
import { Config as Config$1 } from '../listening/listening_pb.js'
import { Config as Config$2 } from './accept/accept_pb.js'
import { Data } from './rpc/rpc_pb.js'
import { Config as Config$3 } from './dial/dial_pb.js'

/**
 * ForwardStreamsRequest is the request type for ForwardStreams.
 *
 * @generated from message stream.api.ForwardStreamsRequest
 */
export class ForwardStreamsRequest extends Message<ForwardStreamsRequest> {
  /**
   * @generated from field: stream.forwarding.Config forwarding_config = 1;
   */
  forwardingConfig?: Config

  constructor(data?: PartialMessage<ForwardStreamsRequest>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.ForwardStreamsRequest'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'forwarding_config', kind: 'message', T: Config },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): ForwardStreamsRequest {
    return new ForwardStreamsRequest().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): ForwardStreamsRequest {
    return new ForwardStreamsRequest().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): ForwardStreamsRequest {
    return new ForwardStreamsRequest().fromJsonString(jsonString, options)
  }

  static equals(
    a: ForwardStreamsRequest | PlainMessage<ForwardStreamsRequest> | undefined,
    b: ForwardStreamsRequest | PlainMessage<ForwardStreamsRequest> | undefined,
  ): boolean {
    return proto3.util.equals(ForwardStreamsRequest, a, b)
  }
}

/**
 * ForwardStreamsResponse is the response type for ForwardStreams.
 *
 * @generated from message stream.api.ForwardStreamsResponse
 */
export class ForwardStreamsResponse extends Message<ForwardStreamsResponse> {
  /**
   * ControllerStatus is the status of the forwarding controller.
   *
   * @generated from field: controller.exec.ControllerStatus controller_status = 1;
   */
  controllerStatus = ControllerStatus.ControllerStatus_UNKNOWN

  constructor(data?: PartialMessage<ForwardStreamsResponse>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.ForwardStreamsResponse'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'controller_status',
      kind: 'enum',
      T: proto3.getEnumType(ControllerStatus),
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): ForwardStreamsResponse {
    return new ForwardStreamsResponse().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): ForwardStreamsResponse {
    return new ForwardStreamsResponse().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): ForwardStreamsResponse {
    return new ForwardStreamsResponse().fromJsonString(jsonString, options)
  }

  static equals(
    a:
      | ForwardStreamsResponse
      | PlainMessage<ForwardStreamsResponse>
      | undefined,
    b:
      | ForwardStreamsResponse
      | PlainMessage<ForwardStreamsResponse>
      | undefined,
  ): boolean {
    return proto3.util.equals(ForwardStreamsResponse, a, b)
  }
}

/**
 * ListenStreamsRequest is the request type for ListenStreams.
 *
 * @generated from message stream.api.ListenStreamsRequest
 */
export class ListenStreamsRequest extends Message<ListenStreamsRequest> {
  /**
   * @generated from field: stream.listening.Config listening_config = 1;
   */
  listeningConfig?: Config$1

  constructor(data?: PartialMessage<ListenStreamsRequest>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.ListenStreamsRequest'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'listening_config', kind: 'message', T: Config$1 },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): ListenStreamsRequest {
    return new ListenStreamsRequest().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): ListenStreamsRequest {
    return new ListenStreamsRequest().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): ListenStreamsRequest {
    return new ListenStreamsRequest().fromJsonString(jsonString, options)
  }

  static equals(
    a: ListenStreamsRequest | PlainMessage<ListenStreamsRequest> | undefined,
    b: ListenStreamsRequest | PlainMessage<ListenStreamsRequest> | undefined,
  ): boolean {
    return proto3.util.equals(ListenStreamsRequest, a, b)
  }
}

/**
 * ListenStreamsResponse is the response type for ListenStreams.
 *
 * @generated from message stream.api.ListenStreamsResponse
 */
export class ListenStreamsResponse extends Message<ListenStreamsResponse> {
  /**
   * ControllerStatus is the status of the forwarding controller.
   *
   * @generated from field: controller.exec.ControllerStatus controller_status = 1;
   */
  controllerStatus = ControllerStatus.ControllerStatus_UNKNOWN

  constructor(data?: PartialMessage<ListenStreamsResponse>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.ListenStreamsResponse'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'controller_status',
      kind: 'enum',
      T: proto3.getEnumType(ControllerStatus),
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): ListenStreamsResponse {
    return new ListenStreamsResponse().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): ListenStreamsResponse {
    return new ListenStreamsResponse().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): ListenStreamsResponse {
    return new ListenStreamsResponse().fromJsonString(jsonString, options)
  }

  static equals(
    a: ListenStreamsResponse | PlainMessage<ListenStreamsResponse> | undefined,
    b: ListenStreamsResponse | PlainMessage<ListenStreamsResponse> | undefined,
  ): boolean {
    return proto3.util.equals(ListenStreamsResponse, a, b)
  }
}

/**
 * AcceptStreamRequest is the request type for AcceptStream.
 *
 * @generated from message stream.api.AcceptStreamRequest
 */
export class AcceptStreamRequest extends Message<AcceptStreamRequest> {
  /**
   * Config is the configuration for the accept.
   * The first packet will contain this value.
   *
   * @generated from field: stream.api.accept.Config config = 1;
   */
  config?: Config$2

  /**
   * Data is a data packet.
   *
   * @generated from field: stream.api.rpc.Data data = 2;
   */
  data?: Data

  constructor(data?: PartialMessage<AcceptStreamRequest>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.AcceptStreamRequest'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'config', kind: 'message', T: Config$2 },
    { no: 2, name: 'data', kind: 'message', T: Data },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): AcceptStreamRequest {
    return new AcceptStreamRequest().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): AcceptStreamRequest {
    return new AcceptStreamRequest().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): AcceptStreamRequest {
    return new AcceptStreamRequest().fromJsonString(jsonString, options)
  }

  static equals(
    a: AcceptStreamRequest | PlainMessage<AcceptStreamRequest> | undefined,
    b: AcceptStreamRequest | PlainMessage<AcceptStreamRequest> | undefined,
  ): boolean {
    return proto3.util.equals(AcceptStreamRequest, a, b)
  }
}

/**
 * AcceptStreamResponse is the response type for AcceptStream.
 *
 * @generated from message stream.api.AcceptStreamResponse
 */
export class AcceptStreamResponse extends Message<AcceptStreamResponse> {
  /**
   * Data is a data packet.
   *
   * @generated from field: stream.api.rpc.Data data = 1;
   */
  data?: Data

  constructor(data?: PartialMessage<AcceptStreamResponse>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.AcceptStreamResponse'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'data', kind: 'message', T: Data },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): AcceptStreamResponse {
    return new AcceptStreamResponse().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): AcceptStreamResponse {
    return new AcceptStreamResponse().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): AcceptStreamResponse {
    return new AcceptStreamResponse().fromJsonString(jsonString, options)
  }

  static equals(
    a: AcceptStreamResponse | PlainMessage<AcceptStreamResponse> | undefined,
    b: AcceptStreamResponse | PlainMessage<AcceptStreamResponse> | undefined,
  ): boolean {
    return proto3.util.equals(AcceptStreamResponse, a, b)
  }
}

/**
 * DialStreamRequest is the request type for DialStream.
 *
 * @generated from message stream.api.DialStreamRequest
 */
export class DialStreamRequest extends Message<DialStreamRequest> {
  /**
   * Config is the configuration for the dial.
   * The first packet will contain this value.
   *
   * @generated from field: stream.api.dial.Config config = 1;
   */
  config?: Config$3

  /**
   * Data is a data packet.
   *
   * @generated from field: stream.api.rpc.Data data = 2;
   */
  data?: Data

  constructor(data?: PartialMessage<DialStreamRequest>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.DialStreamRequest'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'config', kind: 'message', T: Config$3 },
    { no: 2, name: 'data', kind: 'message', T: Data },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): DialStreamRequest {
    return new DialStreamRequest().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): DialStreamRequest {
    return new DialStreamRequest().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): DialStreamRequest {
    return new DialStreamRequest().fromJsonString(jsonString, options)
  }

  static equals(
    a: DialStreamRequest | PlainMessage<DialStreamRequest> | undefined,
    b: DialStreamRequest | PlainMessage<DialStreamRequest> | undefined,
  ): boolean {
    return proto3.util.equals(DialStreamRequest, a, b)
  }
}

/**
 * DialStreamResponse is the response type for DialStream.
 *
 * @generated from message stream.api.DialStreamResponse
 */
export class DialStreamResponse extends Message<DialStreamResponse> {
  /**
   * Data is a data packet.
   *
   * @generated from field: stream.api.rpc.Data data = 1;
   */
  data?: Data

  constructor(data?: PartialMessage<DialStreamResponse>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'stream.api.DialStreamResponse'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'data', kind: 'message', T: Data },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): DialStreamResponse {
    return new DialStreamResponse().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): DialStreamResponse {
    return new DialStreamResponse().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): DialStreamResponse {
    return new DialStreamResponse().fromJsonString(jsonString, options)
  }

  static equals(
    a: DialStreamResponse | PlainMessage<DialStreamResponse> | undefined,
    b: DialStreamResponse | PlainMessage<DialStreamResponse> | undefined,
  ): boolean {
    return proto3.util.equals(DialStreamResponse, a, b)
  }
}
